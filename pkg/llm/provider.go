// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"github.com/clivern/sloth/pkg/llm/muse"
	"github.com/clivern/sloth/pkg/llm/openrouter"
	"github.com/spf13/viper"
)

const (
	ProviderOpenRouter = "openrouter"
	ProviderMuse       = "muse"
)

// IsConfigured reports whether an AI provider API key is set.
func IsConfigured() bool {
	return viper.GetString("app.llm.api_key") != ""
}

// NewChatClient returns the chat client for the app.llm.provider config.
func NewChatClient() ChatClient {
	apiKey := viper.GetString("app.llm.api_key")
	model := viper.GetString("app.llm.chat_model")

	if viper.GetString("app.llm.provider") == ProviderMuse {
		return NewMuseChat(muse.New(muse.Config{
			APIKey: apiKey,
			Model:  model,
			Effort: viper.GetString("app.llm.reasoning_effort"),
		}))
	}

	return NewOpenRouterChat(openrouter.New(openrouter.Config{APIKey: apiKey, Model: model}))
}
