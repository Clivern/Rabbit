// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationChatRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	user := createTestUser(t, database)
	other := createTestUser(t, database)
	repo := NewChatRepository(database)

	chat := &Chat{
		WorkspaceId: workspace.Id,
		UserId:      user.Id,
		Status:      ChatStatusIdle,
	}

	t.Run("create and get belongs to its user", func(t *testing.T) {
		require.NoError(t, repo.CreateChat(chat))

		got, err := repo.GetChat(workspace.Id, user.Id, chat.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ChatStatusIdle, got.Status)

		got, err = repo.GetChat(workspace.Id, other.Id, chat.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("messages keep their order", func(t *testing.T) {
		require.NoError(t, repo.CreateMessage(&ChatMessage{ChatId: chat.Id, Kind: ChatMessageKindUser, Content: "Which agents do I have?"}))
		require.NoError(t, repo.CreateMessage(&ChatMessage{ChatId: chat.Id, Kind: ChatMessageKindAssistant, Content: "You have one agent, users.", Meta: stringPtr(`{"model": "muse-wonder-1.0"}`)}))

		messages, err := repo.ListMessages(chat.Id)
		require.NoError(t, err)
		require.Len(t, messages, 2)
		assert.Equal(t, ChatMessageKindUser, messages[0].Kind)
		assert.Nil(t, messages[0].Meta)
		assert.Equal(t, "You have one agent, users.", messages[1].Content)
		assert.JSONEq(t, `{"model": "muse-wonder-1.0"}`, *messages[1].Meta)
	})

	t.Run("list shows the latest message as preview", func(t *testing.T) {
		chat.Title = "Which agents do I have?"
		chat.Status = ChatStatusRunning
		require.NoError(t, repo.UpdateChat(chat))

		list, err := repo.ListChats(workspace.Id, user.Id, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, "Which agents do I have?", list[0].Title)
		assert.Equal(t, ChatStatusRunning, list[0].Status)
		assert.Equal(t, "You have one agent, users.", *list[0].Preview)

		list, err = repo.ListChats(workspace.Id, other.Id, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("delete belongs to its user and removes the messages", func(t *testing.T) {
		deleted, err := repo.DeleteChat(workspace.Id, other.Id, chat.Id)
		require.NoError(t, err)
		assert.False(t, deleted)

		deleted, err = repo.DeleteChat(workspace.Id, user.Id, chat.Id)
		require.NoError(t, err)
		assert.True(t, deleted)

		got, err := repo.GetChat(workspace.Id, user.Id, chat.Id)
		require.NoError(t, err)
		assert.Nil(t, got)

		messages, err := repo.ListMessages(chat.Id)
		require.NoError(t, err)
		assert.Empty(t, messages)
	})
}
