// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListChatsAction lists the user's chats in a workspace.
func (a *API) ListChatsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chats, err := a.Chat.ListChats(db.Id(wid), user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to list chats")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"chats": chats,
	})
}

// CreateChatAction starts an empty chat for the user.
func (a *API) CreateChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chat, err := a.Chat.CreateChat(db.Id(wid), user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to create chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusCreated, chat)
}

// GetChatAction returns one of the user's chats with its messages.
func (a *API) GetChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_chat_id"),
		})
		return
	}

	chat, err := a.Chat.GetChat(db.Id(wid), user.Id, db.Id(chatId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to get chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, chat)
}

// DeleteChatAction removes one of the user's chats with its messages.
func (a *API) DeleteChatAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_chat_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("chatId", chatId).
		Msg("Deleting chat")

	err := a.Chat.DeleteChat(db.Id(wid), user.Id, db.Id(chatId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to delete chat")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendChatMessageAction stores the user's question and streams the answer as server-sent events.
func (a *API) SendChatMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	user, _ := a.GetUser(r)

	chatId := chi.URLParam(r, "chatId")
	if lo.IsEmpty(chatId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_chat_id"),
		})
		return
	}

	var req module.SendChatMessageRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	chat, err := a.Chat.SendMessage(db.Id(wid), user.Id, db.Id(chatId), &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrChatNotFound, http.StatusNotFound, locale.TR(r, "chat_not_found")},
			{module.ErrChatNotConfigured, http.StatusServiceUnavailable, locale.TR(r, "chat_not_configured")},
			{module.ErrChatBusy, http.StatusConflict, locale.TR(r, "chat_busy")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("chatId", chatId).
			Msg("Failed to send chat message")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_chat_request"),
		})

		return
	}

	// An answer can outlive the server write timeout, so the stream gets its own deadline.
	a.Chat.Reply(chat, a.EventStream(w, module.ChatReplyTimeout+time.Minute))
}
