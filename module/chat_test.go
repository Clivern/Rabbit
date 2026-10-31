// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/llm"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitChat(t *testing.T) {
	// newAssistant wires an assistant with two tools to a scripted model; the user holds every permission in allowed.
	newAssistant := func(allowed map[string]bool, replies ...string) (*Chat, *fakeChats, *fakeModel, *[]map[string]string) {
		var ran []map[string]string
		repository := &fakeChats{
			chat: &db.Chat{Id: "chat-1", WorkspaceId: "ws-1", UserId: "user-1", Status: db.ChatStatusIdle},
		}
		chat := &fakeModel{replies: replies}

		assistant := &Chat{
			ChatRepository: repository,
			Tools: []*ChatTool{
				{
					Name:       "list_agents",
					Permission: CanListAgents,
					Run: func(workspaceId, userId db.Id, args map[string]string) (any, error) {
						ran = append(ran, args)
						return map[string]any{"agents": []string{"users", "weather"}}, nil
					},
				},
				{
					Name:       "get_agent",
					Arguments:  map[string]string{"name": "agent name"},
					Permission: CanGetAgent,
					Run: func(workspaceId, userId db.Id, args map[string]string) (any, error) {
						ran = append(ran, args)
						return nil, errors.New("agent not found")
					},
				},
			},
			Can: func(workspaceId, userId db.Id, permission string) (bool, error) {
				return allowed[permission], nil
			},
			NewModel: func() llm.ChatClient { return chat },
		}

		return assistant, repository, chat, &ran
	}

	everything := map[string]bool{CanListAgents: true, CanGetAgent: true}

	t.Run("answer runs the planned tools once and stores only the answer", func(t *testing.T) {
		assistant, repository, chat, ran := newAssistant(everything,
			"```json\n{\"calls\": [{\"tool\": \"list_agents\"}, {\"tool\": \"drop_database\"}, {\"tool\": \"get_agent\", \"arguments\": {\"name\": \"users\"}}]}\n```",
			"You have two agents, users and weather.",
		)
		repository.messages = []*db.ChatMessage{{Kind: db.ChatMessageKindUser, Content: "Which agents do I have?"}}

		require.NoError(t, assistant.Answer(context.Background(), repository.chat, func(string, any) {}))

		// One plan, one answer; unknown tools are dropped, the rest run in order.
		require.Len(t, chat.asked, 2)
		require.Len(t, *ran, 2)
		assert.Equal(t, "users", (*ran)[1]["name"])

		require.Len(t, repository.messages, 2)
		assert.Equal(t, db.ChatMessageKindAssistant, repository.messages[1].Kind)
		assert.Equal(t, "You have two agents, users and weather.", repository.messages[1].Content)

		// Looked-up data reaches the answer as instructions, never as chat turns.
		answerPrompt := chat.asked[1]
		require.Len(t, answerPrompt, 2)
		assert.Contains(t, answerPrompt[0].Content, "weather")
		assert.Contains(t, answerPrompt[0].Content, "could not be read (agent not found)")
	})

	t.Run("send message stores the question and reply streams the chat and answer, then stores it", func(t *testing.T) {
		viper.Set("app.llm.api_key", "key")
		t.Cleanup(func() { viper.Set("app.llm.api_key", "") })
		assistant, repository, _, _ := newAssistant(everything,
			`{"calls": [{"tool": "list_agents"}]}`,
			"You have two agents.",
		)

		chat, err := assistant.SendMessage("ws-1", "user-1", "chat-1", &SendChatMessageRequest{Content: " Which agents do I have? "})

		require.NoError(t, err)
		assert.Equal(t, db.ChatStatusRunning, chat.Status)
		require.Len(t, repository.messages, 1)

		var events []string
		var deltas strings.Builder
		var finished *ChatResponse
		assistant.Reply(chat, func(event string, data any) {
			events = append(events, event)
			switch event {
			case ChatEventDelta:
				deltas.WriteString(data.(map[string]string)["content"])
			case ChatEventDone:
				finished = data.(*ChatResponse)
			}
		})

		assert.Equal(t, []string{"chat", "delta", "delta", "delta", "delta", "done"}, events)
		assert.Equal(t, "You have two agents.", deltas.String())
		assert.Equal(t, "Which agents do I have?", repository.chat.Title)
		assert.Equal(t, db.ChatStatusIdle, finished.Status)
		require.Len(t, finished.Messages, 2)
		assert.Equal(t, "You have two agents.", finished.Messages[1].Content)
	})

	t.Run("answer only offers tools the user may run", func(t *testing.T) {
		assistant, repository, chat, _ := newAssistant(map[string]bool{CanListAgents: true},
			`{"calls": []}`,
			"I can list your agents.",
		)
		repository.messages = []*db.ChatMessage{{Kind: db.ChatMessageKindUser, Content: "What can you do?"}}

		require.NoError(t, assistant.Answer(context.Background(), repository.chat, func(string, any) {}))

		offered := chat.asked[0][len(chat.asked[0])-1].Content
		assert.Contains(t, offered, "list_agents")
		assert.NotContains(t, offered, "get_agent")

		// The persona only claims what this user may do.
		answerPrompt := chat.asked[1][0].Content
		assert.Contains(t, answerPrompt, "What you can do:\n- ")
		assert.Equal(t, 1, strings.Count(answerPrompt[strings.Index(answerPrompt, "What you can do:"):strings.Index(answerPrompt, "Workspace data:")], "\n- "))
	})

	t.Run("call is denied when the user lost the permission", func(t *testing.T) {
		allowed := map[string]bool{CanListAgents: true}
		assistant, repository, _, ran := newAssistant(allowed)
		tools, err := assistant.AllowedTools(repository.chat)
		require.NoError(t, err)
		allowed[CanListAgents] = false

		result, err := assistant.Call(repository.chat, tools, ChatToolCall{Tool: "list_agents"})

		require.NoError(t, err)
		assert.Empty(t, *ran)
		assert.Contains(t, result, "not allowed to see")
	})

	t.Run("answer skips planning when the user may run no tool", func(t *testing.T) {
		assistant, repository, chat, _ := newAssistant(map[string]bool{}, "Sloth is a gateway for AI agents.")
		repository.messages = []*db.ChatMessage{{Kind: db.ChatMessageKindUser, Content: "What is Sloth?"}}

		require.NoError(t, assistant.Answer(context.Background(), repository.chat, func(string, any) {}))

		assert.Len(t, chat.asked, 1)
		assert.Len(t, repository.messages, 2)
	})

	t.Run("plan that is not json means no calls", func(t *testing.T) {
		assistant, _, chat, _ := newAssistant(everything, "No tool helps here.")

		plan, err := assistant.Plan(context.Background(), chat, nil, assistant.Tools, nil)

		require.NoError(t, err)
		assert.Empty(t, plan.Calls)
	})

	t.Run("delete chat removes it and a missing chat is not found", func(t *testing.T) {
		assistant, repository, _, _ := newAssistant(everything)

		require.NoError(t, assistant.DeleteChat("ws-1", "user-1", "chat-1"))
		assert.Nil(t, repository.chat)

		assert.ErrorIs(t, assistant.DeleteChat("ws-1", "user-1", "chat-1"), ErrChatNotFound)
	})

	t.Run("send message needs a configured provider", func(t *testing.T) {
		viper.Set("app.llm.api_key", "")
		assistant, _, _, _ := newAssistant(everything)

		_, err := assistant.SendMessage("ws-1", "user-1", "chat-1", &SendChatMessageRequest{Content: "hi"})
		assert.ErrorIs(t, err, ErrChatNotConfigured)
	})

	t.Run("send message refuses while the assistant is answering", func(t *testing.T) {
		viper.Set("app.llm.api_key", "key")
		t.Cleanup(func() { viper.Set("app.llm.api_key", "") })
		assistant, repository, _, _ := newAssistant(everything)
		repository.chat.Status = db.ChatStatusRunning

		_, err := assistant.SendMessage("ws-1", "user-1", "chat-1", &SendChatMessageRequest{Content: "hi"})
		assert.ErrorIs(t, err, ErrChatBusy)
	})

	t.Run("send message to a missing chat is not found", func(t *testing.T) {
		viper.Set("app.llm.api_key", "key")
		t.Cleanup(func() { viper.Set("app.llm.api_key", "") })
		assistant, repository, _, _ := newAssistant(everything)
		repository.chat = nil

		_, err := assistant.SendMessage("ws-1", "user-1", "chat-1", &SendChatMessageRequest{Content: "hi"})
		assert.ErrorIs(t, err, ErrChatNotFound)
	})
}

func TestUnitChatTools(t *testing.T) {
	workspaces := &fakeWorkspaces{
		workspace:  &db.Workspace{Id: "ws-1", Name: "Acme", Handle: "acme"},
		membership: &db.WorkspaceUser{Role: db.UserRoleOwner},
	}
	agents := &fakeAgents{list: []*db.Agent{{Id: "agent-1", Name: "users"}}, total: 1}
	intentions := &fakeIntentions{
		list:  []*db.Intention{{Action: db.IntentionActionAllow}, {Action: db.IntentionActionAllow}, {Action: db.IntentionActionDeny}},
		total: 3,
	}

	tools := NewChatTools(
		NewWorkspace(workspaces, &fakeWorkspaceUsers{total: 3}, nil, nil),
		NewAgent(agents, &fakeInstances{}, &fakeHealthChecks{}, workspaces),
		nil,
		NewIntention(intentions, workspaces),
		NewTraffic(&fakeTrafficCalls{total: 42}, workspaces),
		NewKeyValue(&fakeWorkspaceKeyValues{items: []*db.WorkspaceKeyValue{{Key: "feature/assistant"}}}, workspaces),
		nil,
		NewIntegration(&fakeIntegrations{list: []*db.Integration{{
			Type:     "slack",
			Name:     "Alerts",
			IsActive: true,
			Config:   lo.ToPtr(`{"url":"https://hooks.slack.com/services/T0/B0/secret-token","events":["agent.health.changed"]}`),
		}}, total: 1}),
	)

	tool, ok := lo.Find(tools.List(), func(tool *ChatTool) bool { return tool.Name == "get_workspace" })
	require.True(t, ok)
	assert.Equal(t, CanGetWorkspace, tool.Permission)

	result, err := tool.Run("ws-1", "user-1", nil)

	require.NoError(t, err)
	summary := result.(*ChatWorkspaceSummary)
	assert.Equal(t, "Acme", summary.Workspace.Name)
	assert.Equal(t, db.Id("ws-1"), summary.Workspace.Id)
	assert.Equal(t, db.UserRoleOwner, summary.Workspace.Role)
	assert.Equal(t, int64(3), summary.Workspace.MembersCount)
	assert.Equal(t, int64(1), summary.Agents)
	assert.Len(t, summary.AgentsByHealth, 1)
	assert.Equal(t, int64(3), summary.Intentions)
	assert.Equal(t, map[string]int{db.IntentionActionAllow: 2, db.IntentionActionDeny: 1}, summary.IntentionsByAction)
	assert.Equal(t, int64(1), summary.KeyValues)
	assert.Equal(t, int64(42), summary.GatewayCalls)

	// Integration URLs can carry credentials, so they never reach the model.
	integrations, ok := lo.Find(tools.List(), func(tool *ChatTool) bool { return tool.Name == "list_integrations" })
	require.True(t, ok)
	assert.Equal(t, CanListIntegrations, integrations.Permission)

	listed, err := integrations.Run("ws-1", "user-1", nil)
	require.NoError(t, err)
	body, _ := json.Marshal(listed)
	assert.Contains(t, string(body), `"name":"Alerts"`)
	assert.Contains(t, string(body), "agent.health.changed")
	assert.NotContains(t, string(body), "hooks.slack.com")

	members, ok := lo.Find(tools.List(), func(tool *ChatTool) bool { return tool.Name == "list_members" })
	require.True(t, ok)
	assert.Equal(t, CanListWorkspaceMembers, members.Permission)
}
