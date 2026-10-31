// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/stretchr/testify/assert"
)

func TestUnitTrafficActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	taskId := "task-1"
	stored := &db.TrafficCall{
		Id:                  "call-1",
		WorkspaceId:         "ws-1",
		SourceAgent:         "frontend",
		DestinationAgent:    "users",
		DestinationInstance: "users-1",
		Skill:               "lookup",
		Verb:                "POST",
		Decision:            "allow",
		Result:              "ok",
		LatencyMs:           42,
		TaskId:              &taskId,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	newAPI := func(workspace *db.Workspace, calls *fakeTrafficCalls) *API {
		return &API{
			Traffic: module.NewTraffic(calls, &fakeWorkspaceRecords{workspace: workspace}),
		}
	}

	t.Run("list returns calls with total", func(t *testing.T) {
		api := newAPI(workspace, &fakeTrafficCalls{calls: []*db.TrafficCall{stored}, total: 1})
		w := httptest.NewRecorder()
		newTrafficRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/traffic", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"latencyMs":42`)
		assert.Contains(t, w.Body.String(), `"destinationAgent":"users"`)
		assert.Contains(t, w.Body.String(), `"destinationInstance":"users-1"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeTrafficCalls{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newTrafficRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/traffic", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get returns one call", func(t *testing.T) {
		api := newAPI(workspace, &fakeTrafficCalls{call: stored})
		w := httptest.NewRecorder()
		newTrafficRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/traffic/call-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"taskId":"task-1"`)
	})

	t.Run("get returns 404 when the call belongs to another workspace", func(t *testing.T) {
		other := *stored
		other.WorkspaceId = "ws-2"
		api := newAPI(workspace, &fakeTrafficCalls{call: &other})
		w := httptest.NewRecorder()
		newTrafficRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/traffic/call-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "traffic_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeTrafficCalls{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newTrafficRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/traffic/call-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get returns 400 when the traffic id is missing", func(t *testing.T) {
		api := newAPI(workspace, &fakeTrafficCalls{})
		w := httptest.NewRecorder()
		api.GetTrafficAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
