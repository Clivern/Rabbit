// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package openrouter

import (
	orsdk "github.com/OpenRouterTeam/go-sdk"
)

// Config holds OpenRouter API credentials and the model to call.
type Config struct {
	APIKey string
	Model  string
}

// Client calls OpenRouter chat APIs for a single model.
type Client struct {
	api   *orsdk.OpenRouter
	model string
}

// Message is a chat message passed to Complete.
type Message struct {
	Role    string
	Content string
}

// Usage is token and USD cost usage reported by OpenRouter.
type Usage struct {
	PromptTokens int64
	TotalTokens  int64
	Cost         float64
}

// New returns an OpenRouter client.
func New(cfg Config) *Client {
	return &Client{
		api:   orsdk.New(orsdk.WithSecurity(cfg.APIKey)),
		model: cfg.Model,
	}
}
