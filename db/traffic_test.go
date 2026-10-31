// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationTrafficRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewTrafficRepository(database)

	t.Run("create get list count", func(t *testing.T) {
		taskId := "task-8f2a"
		call := &TrafficCall{
			WorkspaceId:         workspace.Id,
			SourceAgent:         "support-assistant-2",
			SourceInstance:      "support-assistant-2-1",
			DestinationAgent:    "user-directory",
			DestinationInstance: "user-directory-1",
			Skill:               "user.profile.get",
			Verb:                "message:send",
			Decision:            TrafficDecisionAllow,
			Result:              TrafficResultCompleted,
			LatencyMs:           42,
			TaskId:              &taskId,
		}
		require.NoError(t, repo.Create(call))
		assert.NotEmpty(t, call.Id.String())

		got, err := repo.GetById(call.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "support-assistant-2", got.SourceAgent)
		assert.Equal(t, "support-assistant-2-1", got.SourceInstance)
		assert.Equal(t, "user-directory", got.DestinationAgent)
		assert.Equal(t, "user-directory-1", got.DestinationInstance)
		assert.Equal(t, "user.profile.get", got.Skill)
		assert.Equal(t, TrafficDecisionAllow, got.Decision)
		assert.Equal(t, TrafficResultCompleted, got.Result)
		assert.Equal(t, 42, got.LatencyMs)

		require.NoError(t, repo.Create(&TrafficCall{
			WorkspaceId:         workspace.Id,
			SourceAgent:         "support-assistant-2",
			DestinationInstance: "invoice-extractor-1",
			Skill:               "invoice.extract",
			Verb:                "message:send",
			Decision:            TrafficDecisionDeny,
			Result:              TrafficResultDenied,
			LatencyMs:           0,
		}))

		list, err := repo.ListByWorkspaceId(workspace.Id, 50, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("get allowed by task id skips denied calls and other workspaces", func(t *testing.T) {
		taskId := "task-b41c"
		other := createTestWorkspace(t, database)

		require.NoError(t, repo.Create(&TrafficCall{
			WorkspaceId:         workspace.Id,
			SourceAgent:         "support-assistant",
			DestinationInstance: "user-directory-1",
			Verb:                "SendMessage",
			Decision:            TrafficDecisionAllow,
			Result:              TrafficResultCompleted,
			TaskId:              &taskId,
			Meta:                new(`{"agentId":"a","instanceId":"i"}`),
		}))
		require.NoError(t, repo.Create(&TrafficCall{
			WorkspaceId:         workspace.Id,
			SourceAgent:         "support-assistant",
			DestinationInstance: "user-directory",
			Verb:                "GetTask",
			Decision:            TrafficDecisionDeny,
			Result:              TrafficResultDenied,
			TaskId:              &taskId,
		}))

		got, err := repo.GetAllowedByTaskId(workspace.Id, taskId)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, TrafficDecisionAllow, got.Decision)
		assert.Equal(t, "user-directory-1", got.DestinationInstance)
		assert.JSONEq(t, `{"agentId":"a","instanceId":"i"}`, *got.Meta)

		got, err = repo.GetAllowedByTaskId(other.Id, taskId)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("missing row", func(t *testing.T) {
		id, err := NewId()
		require.NoError(t, err)
		got, err := repo.GetById(id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
