// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	ChatStatusIdle    = "idle"
	ChatStatusRunning = "running"
	ChatStatusFailed  = "failed"

	ChatMessageKindUser      = "user"
	ChatMessageKindAssistant = "assistant"
)

// Chat is a single row in the ai_assistant_chats table.
type Chat struct {
	Id          Id
	WorkspaceId Id
	UserId      Id
	Title       string
	Status      string
	Preview     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ChatMessage is a single row in the ai_assistant_messages table.
type ChatMessage struct {
	Id        Id
	ChatId    Id
	Kind      string
	Content   string
	Meta      *string
	CreatedAt time.Time
}

// ChatRepository is the interface for assistant chat persistence.
type ChatRepository interface {
	CreateChat(chat *Chat) error
	GetChat(workspaceId, userId, id Id) (*Chat, error)
	ListChats(workspaceId, userId Id, limit, offset int) ([]*Chat, error)
	UpdateChat(chat *Chat) error
	DeleteChat(workspaceId, userId, id Id) (bool, error)
	CreateMessage(message *ChatMessage) error
	ListMessages(chatId Id) ([]*ChatMessage, error)
}

type ChatRepositoryPostgres struct {
	db *sql.DB
}

// NewChatRepository returns an assistant repository.
func NewChatRepository(db *sql.DB) ChatRepository {
	return &ChatRepositoryPostgres{db: db}
}

// scanChat maps a row into a chat. Column order must match the select list.
func scanChat(chat *Chat) []any {
	return []any{
		&chat.Id,
		&chat.WorkspaceId,
		&chat.UserId,
		&chat.Title,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	}
}

// scanChatPreview maps a row into a chat with its last message. Column order must match the select list.
func scanChatPreview(chat *Chat) []any {
	return []any{
		&chat.Id,
		&chat.WorkspaceId,
		&chat.UserId,
		&chat.Title,
		&chat.Status,
		&chat.Preview,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	}
}

// scanChatMessage maps a row into a chat message. Column order must match the select list.
func scanChatMessage(message *ChatMessage) []any {
	return []any{
		&message.Id,
		&message.ChatId,
		&message.Kind,
		&message.Content,
		&message.Meta,
		&message.CreatedAt,
	}
}

// CreateChat inserts a chat row.
func (r *ChatRepositoryPostgres) CreateChat(chat *Chat) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	chat.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_assistant_chats (id, workspace_id, user_id, title, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		chat.Id.String(),
		chat.WorkspaceId.String(),
		chat.UserId.String(),
		chat.Title,
		chat.Status,
	).Scan(&chat.CreatedAt, &chat.UpdatedAt)
}

// GetChat returns a user's chat in a workspace.
func (r *ChatRepositoryPostgres) GetChat(workspaceId, userId, id Id) (*Chat, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, user_id, title, status, created_at, updated_at
		FROM ai_assistant_chats
		WHERE workspace_id = $1 AND user_id = $2 AND id = $3`,
		scanChat,
		workspaceId.String(),
		userId.String(),
		id.String(),
	)
}

// ListChats returns a user's chats in a workspace with their last message, most recently active first.
func (r *ChatRepositoryPostgres) ListChats(workspaceId, userId Id, limit, offset int) ([]*Chat, error) {
	return QueryMany(
		r.db,
		`SELECT
			c.id, c.workspace_id, c.user_id, c.title, c.status,
			(
				SELECT m.content
				FROM ai_assistant_messages m
				WHERE m.chat_id = c.id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			),
			c.created_at, c.updated_at
		FROM ai_assistant_chats c
		WHERE c.workspace_id = $1 AND c.user_id = $2
		ORDER BY c.updated_at DESC
		LIMIT $3 OFFSET $4`,
		scanChatPreview,
		workspaceId.String(),
		userId.String(),
		limit,
		offset,
	)
}

// UpdateChat updates a chat's title and status and marks it active now.
func (r *ChatRepositoryPostgres) UpdateChat(chat *Chat) error {
	chat.UpdatedAt = time.Now().UTC()

	_, err := r.db.Exec(
		`UPDATE ai_assistant_chats
		SET title = $1, status = $2, updated_at = $3
		WHERE id = $4`,
		chat.Title,
		chat.Status,
		chat.UpdatedAt,
		chat.Id.String(),
	)

	return err
}

// DeleteChat removes a user's chat and its messages and reports whether it existed.
func (r *ChatRepositoryPostgres) DeleteChat(workspaceId, userId, id Id) (bool, error) {
	result, err := r.db.Exec(
		`DELETE FROM ai_assistant_chats
		WHERE workspace_id = $1 AND user_id = $2 AND id = $3`,
		workspaceId.String(),
		userId.String(),
		id.String(),
	)
	if err != nil {
		return false, err
	}

	deleted, err := result.RowsAffected()

	return deleted > 0, err
}

// CreateMessage inserts a message row.
func (r *ChatRepositoryPostgres) CreateMessage(message *ChatMessage) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	message.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_assistant_messages (id, chat_id, kind, content, meta)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`,
		message.Id.String(),
		message.ChatId.String(),
		message.Kind,
		message.Content,
		message.Meta,
	).Scan(&message.CreatedAt)
}

// ListMessages returns a chat's messages, oldest first.
func (r *ChatRepositoryPostgres) ListMessages(chatId Id) ([]*ChatMessage, error) {
	return QueryMany(
		r.db,
		`SELECT id, chat_id, kind, content, meta, created_at
		FROM ai_assistant_messages
		WHERE chat_id = $1
		ORDER BY created_at, id`,
		scanChatMessage,
		chatId.String(),
	)
}
