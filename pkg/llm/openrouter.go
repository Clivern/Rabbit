// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"context"

	"github.com/clivern/sloth/pkg/llm/openrouter"
)

// OpenRouterChat adapts the OpenRouter client to ChatClient.
type OpenRouterChat struct {
	client *openrouter.Client
}

// NewOpenRouterChat returns a chat client backed by OpenRouter.
func NewOpenRouterChat(client *openrouter.Client) *OpenRouterChat {
	return &OpenRouterChat{client: client}
}

// Complete sends messages and returns the reply and usage.
func (c *OpenRouterChat) Complete(ctx context.Context, messages []Message) (string, Usage, error) {
	text, usage, err := c.client.Complete(ctx, OpenRouterMessages(messages))

	return text, OpenRouterUsage(usage), err
}

// Stream sends messages, passes each piece of the reply to onDelta and returns the whole reply and usage.
func (c *OpenRouterChat) Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error) {
	text, usage, err := c.client.Stream(ctx, OpenRouterMessages(messages), onDelta)

	return text, OpenRouterUsage(usage), err
}

// OpenRouterMessages converts messages into OpenRouter messages.
func OpenRouterMessages(messages []Message) []openrouter.Message {
	items := make([]openrouter.Message, 0, len(messages))
	for _, message := range messages {
		items = append(items, openrouter.Message{Role: message.Role, Content: message.Content})
	}

	return items
}

// OpenRouterUsage converts OpenRouter usage into Usage with nano-USD cost.
func OpenRouterUsage(usage openrouter.Usage) Usage {
	return Usage{
		PromptTokens: usage.PromptTokens,
		TotalTokens:  usage.TotalTokens,
		Cost:         CostFromUSD(usage.Cost),
	}
}
