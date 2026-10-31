// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"context"

	"github.com/clivern/sloth/pkg/llm/muse"
)

// MuseChat adapts the Meta Muse client to ChatClient.
type MuseChat struct {
	client *muse.Client
}

// NewMuseChat returns a chat client backed by Meta Muse.
func NewMuseChat(client *muse.Client) *MuseChat {
	return &MuseChat{client: client}
}

// Complete sends messages and returns the reply and usage.
func (c *MuseChat) Complete(ctx context.Context, messages []Message) (string, Usage, error) {
	text, usage, err := c.client.Complete(ctx, MuseMessages(messages))

	return text, MuseUsage(usage), err
}

// Stream sends messages, passes each piece of the reply to onDelta and returns the whole reply and usage.
func (c *MuseChat) Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error) {
	text, usage, err := c.client.Stream(ctx, MuseMessages(messages), onDelta)

	return text, MuseUsage(usage), err
}

// MuseMessages converts messages into Meta Muse messages.
func MuseMessages(messages []Message) []muse.Message {
	items := make([]muse.Message, 0, len(messages))
	for _, message := range messages {
		items = append(items, muse.Message{Role: message.Role, Content: message.Content})
	}

	return items
}

// MuseUsage converts Meta Muse usage into Usage; Muse reports no cost.
func MuseUsage(usage muse.Usage) Usage {
	return Usage{PromptTokens: usage.InputTokens, TotalTokens: usage.TotalTokens}
}
