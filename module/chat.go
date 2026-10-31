// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/sloth/conf"
	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/llm"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	ChatMaxCalls     = 3
	ChatOutputLength = 8000
	ChatTitleLength  = 60
	ChatListLimit    = 50
	ChatReplyTimeout = 3 * time.Minute

	ChatEventChat  = "chat"
	ChatEventDelta = "delta"
	ChatEventDone  = "done"
	ChatEventError = "error"
)

var (
	ErrChatNotFound      = errors.New("assistant chat not found")
	ErrChatNotConfigured = errors.New("assistant is not configured")
	ErrChatBusy          = errors.New("assistant is still answering")
	ErrFailedChatRequest = errors.New("failed assistant request")
)

// Chat is the module for the dashboard assistant chats.
type Chat struct {
	ChatRepository db.ChatRepository
	Tools          []*ChatTool
	Can            ChatPermissionCheck
	NewModel       func() llm.ChatClient
}

// ChatEmit sends one named event with its data to the user while the assistant answers.
type ChatEmit func(event string, data any)

// ChatPermissionCheck reports whether a user holds a workspace permission.
type ChatPermissionCheck func(workspaceId, userId db.Id, permission string) (bool, error)

// SendChatMessageRequest is the body for asking the assistant a question.
type SendChatMessageRequest struct {
	Content string `json:"content" validate:"required,max=4000" label:"Message"`
}

// ChatPlan is the tool calls the model picked for a question.
type ChatPlan struct {
	Calls []ChatToolCall `json:"calls"`
}

// ChatToolCall is one tool call in a plan.
type ChatToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

// ChatResponse is a chat shaped for API responses.
type ChatResponse struct {
	Id        db.Id                  `json:"id"`
	Title     string                 `json:"title"`
	Status    string                 `json:"status"`
	Preview   string                 `json:"preview,omitempty"`
	Messages  []*ChatMessageResponse `json:"messages,omitempty"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
}

// ChatMessageResponse is a message shaped for API responses.
type ChatMessageResponse struct {
	Id        db.Id  `json:"id"`
	Kind      string `json:"kind"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// NewChat creates an assistant module whose tools run under the user's workspace permissions.
func NewChat(assistant db.ChatRepository, workspaces db.WorkspaceRepository, tools *ChatTools) *Chat {
	return &Chat{
		ChatRepository: assistant,
		Tools:          tools.List(),
		Can: func(workspaceId, userId db.Id, permission string) (bool, error) {
			perm := &Perm{WorkspaceRepository: workspaces}

			return perm.WithUser(&db.User{Id: userId}).WithWorkspaceId(workspaceId).Can(permission)
		},
		NewModel: llm.NewChatClient,
	}
}

// ListChats returns a user's recent chats in a workspace.
func (c *Chat) ListChats(workspaceId, userId db.Id) ([]*ChatResponse, error) {
	chats, err := c.ChatRepository.ListChats(workspaceId, userId, ChatListLimit, 0)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list chats")
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return lo.Map(chats, func(chat *db.Chat, _ int) *ChatResponse {
		return ToChatResponse(chat, nil)
	}), nil
}

// CreateChat starts an empty chat for a user.
func (c *Chat) CreateChat(workspaceId, userId db.Id) (*ChatResponse, error) {
	chat := &db.Chat{
		WorkspaceId: workspaceId,
		UserId:      userId,
		Status:      db.ChatStatusIdle,
	}

	err := c.ChatRepository.CreateChat(chat)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create chat")
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return ToChatResponse(chat, []*db.ChatMessage{}), nil
}

// GetChat returns a user's chat with its messages.
func (c *Chat) GetChat(workspaceId, userId, id db.Id) (*ChatResponse, error) {
	chat, err := c.ChatRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}

	messages, err := c.ChatRepository.ListMessages(chat.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return ToChatResponse(chat, messages), nil
}

// DeleteChat removes a user's chat and its messages. An answer still being written for it is dropped.
func (c *Chat) DeleteChat(workspaceId, userId, id db.Id) error {
	deleted, err := c.ChatRepository.DeleteChat(workspaceId, userId, id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("chatId", id.String()).
			Msg("Failed to delete chat")
		return fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}
	if !deleted {
		return ErrChatNotFound
	}

	return nil
}

// SendMessage stores the user's question and marks the chat running; Reply then answers it.
func (c *Chat) SendMessage(workspaceId, userId, id db.Id, req *SendChatMessageRequest) (*db.Chat, error) {
	if !llm.IsConfigured() {
		return nil, ErrChatNotConfigured
	}

	chat, err := c.ChatRepository.GetChat(workspaceId, userId, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}
	if chat.Status == db.ChatStatusRunning {
		return nil, ErrChatBusy
	}

	content := strings.TrimSpace(req.Content)

	err = c.ChatRepository.CreateMessage(&db.ChatMessage{
		ChatId:  chat.Id,
		Kind:    db.ChatMessageKindUser,
		Content: content,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	if lo.IsEmpty(chat.Title) {
		chat.Title = lo.Substring(content, 0, ChatTitleLength)
	}

	chat.Status = db.ChatStatusRunning

	err = c.ChatRepository.UpdateChat(chat)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedChatRequest, err)
	}

	return chat, nil
}

// Reply streams the chat, the answer and its outcome, and stores the answer even if the user leaves mid-stream.
func (c *Chat) Reply(chat *db.Chat, emit ChatEmit) {
	ctx, cancel := context.WithTimeout(context.Background(), ChatReplyTimeout)
	defer cancel()

	started, err := c.GetChat(chat.WorkspaceId, chat.UserId, chat.Id)
	if err != nil {
		emit(ChatEventError, map[string]string{"status": db.ChatStatusFailed})
		return
	}

	emit(ChatEventChat, started)

	err = c.Answer(ctx, chat, emit)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", chat.WorkspaceId.String()).
			Str("chatId", chat.Id.String()).
			Msg("Failed to answer chat")
	}

	chat.Status = lo.Ternary(err == nil, db.ChatStatusIdle, db.ChatStatusFailed)

	err = c.ChatRepository.UpdateChat(chat)
	if err != nil {
		log.Error().
			Err(err).
			Str("chatId", chat.Id.String()).
			Msg("Failed to update chat")
	}

	finished, err := c.GetChat(chat.WorkspaceId, chat.UserId, chat.Id)
	if err != nil {
		emit(ChatEventError, map[string]string{"status": db.ChatStatusFailed})
		return
	}

	emit(lo.Ternary(finished.Status == db.ChatStatusFailed, ChatEventError, ChatEventDone), finished)
}

// Answer plans one round of tools, runs them, then streams and stores the answer; tool results are never stored.
func (c *Chat) Answer(ctx context.Context, chat *db.Chat, emit ChatEmit) error {
	messages, err := c.ChatRepository.ListMessages(chat.Id)
	if err != nil {
		return err
	}

	tools, err := c.AllowedTools(chat)
	if err != nil {
		return err
	}

	model := c.NewModel()
	history := ChatHistory(messages)

	plan, err := c.Plan(ctx, model, history, tools, nil)
	if err != nil {
		return err
	}

	results := []string{}
	for _, call := range plan.Calls {
		result, err := c.Call(chat, tools, call)
		if err != nil {
			return err
		}

		results = append(results, result)
	}

	// Looked-up data goes in the instructions, not the chat, so the model answers from it instead of narrating it.
	prompt := conf.AssistantAnswer(lo.Map(tools, func(tool *ChatTool, _ int) string { return tool.Description }), results)

	answer, _, err := model.Stream(ctx, append([]llm.Message{{Role: "system", Content: prompt}}, history...), func(delta string) {
		emit(ChatEventDelta, map[string]string{"content": delta})
	})
	if err != nil {
		return err
	}

	return c.ChatRepository.CreateMessage(&db.ChatMessage{
		ChatId:  chat.Id,
		Kind:    db.ChatMessageKindAssistant,
		Content: strings.TrimSpace(answer),
	})
}

// AllowedTools returns the tools the chat's user holds the permission for.
func (c *Chat) AllowedTools(chat *db.Chat) ([]*ChatTool, error) {
	allowed := make([]*ChatTool, 0, len(c.Tools))
	for _, tool := range c.Tools {
		can, err := c.Can(chat.WorkspaceId, chat.UserId, tool.Permission)
		if err != nil {
			return nil, err
		}
		if can {
			allowed = append(allowed, tool)
		}
	}

	return allowed, nil
}

// Plan asks the model which tools to run next; a reply that is not a plan means no calls.
func (c *Chat) Plan(ctx context.Context, model llm.ChatClient, history []llm.Message, tools []*ChatTool, results []string) (*ChatPlan, error) {
	if len(tools) == 0 {
		return &ChatPlan{}, nil
	}

	choices, _ := json.Marshal(tools)

	reply, _, err := model.Complete(ctx, append(
		append([]llm.Message{{Role: "system", Content: conf.AssistantPlanPrompt}}, history...),
		llm.Message{Role: "user", Content: fmt.Sprintf(
			"Tools you may run: %s\nData already looked up for this message:\n%s",
			choices,
			lo.Ternary(len(results) == 0, "none", strings.Join(results, "\n")),
		)},
	))
	if err != nil {
		return nil, err
	}

	plan := &ChatPlan{}
	start, end := strings.Index(reply, "{"), strings.LastIndex(reply, "}")
	if start == -1 || end < start {
		return plan, nil
	}

	_ = json.Unmarshal([]byte(reply[start:end+1]), plan)

	names := lo.SliceToMap(tools, func(tool *ChatTool) (string, bool) { return tool.Name, true })
	plan.Calls = lo.Filter(plan.Calls, func(call ChatToolCall, _ int) bool { return names[call.Tool] })
	plan.Calls = lo.Slice(plan.Calls, 0, ChatMaxCalls)

	return plan, nil
}

// Call runs one tool for the chat's user and returns its outcome as a line of workspace data for the model.
func (c *Chat) Call(chat *db.Chat, tools []*ChatTool, call ChatToolCall) (string, error) {
	tool, _ := lo.Find(tools, func(tool *ChatTool) bool { return tool.Name == call.Tool })
	arguments := lo.CoalesceMapOrEmpty(call.Arguments)
	input, _ := json.Marshal(arguments)

	// The user's role may have changed since the tools were offered.
	can, err := c.Can(chat.WorkspaceId, chat.UserId, tool.Permission)
	if err != nil {
		return "", err
	}
	if !can {
		return fmt.Sprintf("%s %s: the user is not allowed to see this", call.Tool, input), nil
	}

	result, err := tool.Run(chat.WorkspaceId, chat.UserId, arguments)
	if err != nil {
		return fmt.Sprintf("%s %s: could not be read (%s)", call.Tool, input, err), nil
	}

	output, _ := json.Marshal(result)

	return fmt.Sprintf(
		"%s %s: %s",
		call.Tool,
		input,
		lo.Substring(string(output), 0, ChatOutputLength),
	), nil
}

// ChatHistory turns stored messages into model messages.
func ChatHistory(messages []*db.ChatMessage) []llm.Message {
	return lo.Map(messages, func(message *db.ChatMessage, _ int) llm.Message {
		return llm.Message{Role: message.Kind, Content: message.Content}
	})
}

// ToChatResponse shapes a chat and, when given, its messages for API responses.
func ToChatResponse(chat *db.Chat, messages []*db.ChatMessage) *ChatResponse {
	response := &ChatResponse{
		Id:        chat.Id,
		Title:     chat.Title,
		Status:    chat.Status,
		Preview:   lo.FromPtr(chat.Preview),
		CreatedAt: chat.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: chat.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if messages != nil {
		response.Messages = lo.Map(messages, func(message *db.ChatMessage, _ int) *ChatMessageResponse {
			return &ChatMessageResponse{
				Id:        message.Id,
				Kind:      message.Kind,
				Content:   message.Content,
				CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339),
			}
		})
	}

	return response
}
