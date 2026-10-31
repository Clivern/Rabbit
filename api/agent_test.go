// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUnitAgentActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	stored := &db.Agent{
		Id:           "agent-1",
		WorkspaceId:  "ws-1",
		Name:         "user-directory",
		Card:         lo.ToPtr(`{"name":"User Directory","version":"1.0.0"}`),
		CardChecksum: lo.ToPtr("9f2c"),
		CardPath:     "/.well-known/agent-card.json",
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	validBody := `{"version":"1.0.0"}`

	t.Run("list returns agents with pagination meta", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{list: []*db.Agent{stored}, total: 1})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"user-directory"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns an empty page for unknown workspace", func(t *testing.T) {
		api := newAgentAPI(nil, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_agent_request")
	})

	t.Run("get returns one agent", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{agent: stored})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"cardChecksum":"9f2c"`)
	})

	t.Run("get returns 404 for unknown agent", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_not_found")
	})

	t.Run("get returns 404 for an invalid agent name", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/-bad", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_agent_request")
	})

	t.Run("upsert returns 201 when the agent is new", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(validBody),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"version":"1.0.0"`)
	})

	t.Run("upsert returns 200 when the agent already exists", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{agent: stored})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(validBody),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("upsert returns 400 for invalid body", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(`{`),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("upsert returns 400 for a card path that is not absolute", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(`{"version":"1.0.0","cardPath":"agent.json"}`),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("upsert returns 400 when no version is given", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(`{}`),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("upsert returns 400 for invalid agent name", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/-bad",
			strings.NewReader(validBody),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_agent_name")
	})

	t.Run("upsert stores the load balancing strategy", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(`{"version":"1.0.0","lbStrategy":"sticky"}`),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"lbStrategy":"sticky"`)
	})

	t.Run("upsert returns 400 for an unknown load balancing strategy", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(`{"version":"1.0.0","lbStrategy":"weighted"}`),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("upsert returns 500 when the repository fails", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{createErr: errors.New("boom")})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodPut,
			"/api/v1/workspaces/ws-1/agents/user-directory",
			strings.NewReader(validBody),
		)
		newAgentRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_agent_request")
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{agent: stored})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown agent", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_not_found")
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newAgentAPI(workspace, &fakeAgentCatalog{agent: stored, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newAgentRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
