// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"context"
)

// Message is a chat message passed to Complete.
type Message struct {
	Role    string
	Content string
}

// ChatClient completes chat messages with the configured provider.
type ChatClient interface {
	Complete(ctx context.Context, messages []Message) (string, Usage, error)
	Stream(ctx context.Context, messages []Message, onDelta func(string)) (string, Usage, error)
}
