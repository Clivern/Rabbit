// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package openrouter

import (
	"context"
	"fmt"
	"strings"

	orsdk "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
)

// Complete sends messages using this client's model and returns the reply and usage.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, Usage, error) {
	items, err := Items(messages)
	if err != nil {
		return "", Usage{}, err
	}

	response, err := c.api.Chat.Send(ctx, components.ChatRequest{
		Model:    orsdk.Pointer(c.model),
		Messages: items,
	}, nil)
	if err != nil {
		return "", Usage{}, fmt.Errorf("openrouter chat: %w", err)
	}

	if response == nil || response.ChatResult == nil {
		return "", Usage{}, ErrNoCompletion
	}

	choices := response.ChatResult.GetChoices()
	if len(choices) == 0 {
		return "", Usage{}, ErrNoCompletion
	}

	assistantMessage := choices[0].GetMessage()
	content, ok := assistantMessage.GetContent().GetOrZero()
	if !ok {
		return "", Usage{}, ErrNoCompletion
	}

	usage := ChatUsage(response.ChatResult.GetUsage())

	if content.Type == components.ChatAssistantMessageContentTypeStr && content.Str != nil {
		return *content.Str, usage, nil
	}

	return "", usage, ErrNoCompletion
}

// Stream sends messages, passes each piece of the reply to onDelta as it arrives and returns the whole reply and usage.
func (c *Client) Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error) {
	items, err := Items(messages)
	if err != nil {
		return "", Usage{}, err
	}

	response, err := c.api.Chat.Send(ctx, components.ChatRequest{
		Model:    orsdk.Pointer(c.model),
		Messages: items,
		Stream:   orsdk.Pointer(true),
	}, nil)
	if err != nil {
		return "", Usage{}, fmt.Errorf("openrouter chat: %w", err)
	}

	events := response.EventStream
	defer events.Close()

	var reply strings.Builder
	usage := Usage{}

	for events.Next() {
		chunk := events.Value().Data
		if chunk.Error != nil {
			return "", Usage{}, fmt.Errorf("openrouter chat: %s", chunk.Error.Message)
		}

		for _, choice := range chunk.Choices {
			delta, _ := choice.Delta.Content.GetOrZero()
			if delta != "" {
				reply.WriteString(delta)
				onDelta(delta)
			}
		}

		if chunk.Usage != nil {
			usage = ChatUsage(chunk.Usage)
		}
	}

	if err := events.Err(); err != nil {
		return "", Usage{}, fmt.Errorf("openrouter chat: %w", err)
	}

	if reply.Len() == 0 {
		return "", usage, ErrNoCompletion
	}

	return reply.String(), usage, nil
}

// Items converts messages into OpenRouter chat messages.
func Items(messages []Message) ([]components.ChatMessages, error) {
	items := make([]components.ChatMessages, 0, len(messages))

	for _, message := range messages {
		var item components.ChatMessages
		switch message.Role {
		case "system":
			item = components.CreateChatMessagesSystem(components.ChatSystemMessage{
				Role:    components.ChatSystemMessageRoleSystem,
				Content: components.CreateChatSystemMessageContentStr(message.Content),
			})
		case "assistant":
			assistantContent := components.CreateChatAssistantMessageContentStr(message.Content)
			item = components.CreateChatMessagesAssistant(components.ChatAssistantMessage{
				Role:    components.ChatAssistantMessageRoleAssistant,
				Content: optionalnullable.From(&assistantContent),
			})
		case "user", "":
			item = components.CreateChatMessagesUser(components.ChatUserMessage{
				Role:    components.ChatUserMessageRoleUser,
				Content: components.CreateChatUserMessageContentStr(message.Content),
			})
		default:
			return nil, fmt.Errorf("openrouter chat: unsupported role %q", message.Role)
		}

		items = append(items, item)
	}

	return items, nil
}

// ChatUsage converts OpenRouter chat usage into Usage.
func ChatUsage(chatUsage *components.ChatUsage) Usage {
	if chatUsage == nil {
		return Usage{}
	}

	usage := Usage{
		PromptTokens: chatUsage.GetPromptTokens(),
		TotalTokens:  chatUsage.GetTotalTokens(),
	}
	if amount, ok := chatUsage.GetCost().GetOrZero(); ok {
		usage.Cost = amount
	}

	return usage
}
