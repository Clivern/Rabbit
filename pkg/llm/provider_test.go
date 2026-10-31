// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package llm

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitProvider(t *testing.T) {
	t.Run("openrouter", func(t *testing.T) {
		viper.Set("app.llm.provider", ProviderOpenRouter)

		assert.IsType(t, &OpenRouterChat{}, NewChatClient())
	})

	t.Run("muse", func(t *testing.T) {
		viper.Set("app.llm.provider", ProviderMuse)

		assert.IsType(t, &MuseChat{}, NewChatClient())
	})
}
