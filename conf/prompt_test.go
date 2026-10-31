// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package conf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitAssistantAnswer(t *testing.T) {
	t.Run("lists abilities and data after the prompt", func(t *testing.T) {
		prompt := AssistantAnswer([]string{"List agents.", "List KV."}, []string{`list_agents {}: {"Total":0}`})

		assert.True(t, strings.HasPrefix(prompt, AssistantAnswerPrompt))
		assert.Contains(t, prompt, "What you can do:\n- List agents.\n- List KV.\n\nWorkspace data:\n"+`list_agents {}: {"Total":0}`)
	})

	t.Run("falls back when there is nothing to list", func(t *testing.T) {
		prompt := AssistantAnswer(nil, nil)

		assert.Contains(t, prompt, "What you can do:\n- answer general questions about Sloth")
		assert.True(t, strings.HasSuffix(prompt, "Workspace data:\nnone needed"))
	})
}
