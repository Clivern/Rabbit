// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"strings"
	"testing"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/samber/lo"
	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAgent(t *testing.T) {
	viper.Set("app.url", "https://sloth.test")

	t.Run("list returns an empty page for a missing workspace", func(t *testing.T) {
		f := newAgentFixture(nil, nil)

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Empty(t, result.Agents)
	})

	t.Run("list reports unknown health when the agent has no instances", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.agents.total = 1

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		require.Len(t, result.Agents, 1)
		assert.Equal(t, db.AgentInstanceStatusUnknown, result.Agents[0].Health)
		assert.NotEmpty(t, result.Agents[0].UpdatedAt)
		assert.Equal(t, map[string]int{
			db.AgentInstanceStatusPassing:  0,
			db.AgentInstanceStatusWarning:  0,
			db.AgentInstanceStatusUnknown:  0,
			db.AgentInstanceStatusCritical: 0,
		}, result.Agents[0].StatusCounts)
	})

	t.Run(
		"list reports passing health when an instance has no failing checks",
		func(t *testing.T) {
			f := defaultAgentFixture()
			f.agents.list = []*db.Agent{f.agents.agent}
			f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}

			result, err := f.agent.ListAgents("ws-1", 20, 0)
			require.NoError(t, err)
			assert.Equal(t, db.AgentInstanceStatusPassing, result.Agents[0].Health)
			assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusPassing])
		},
	)

	t.Run("list marks an instance with a critical check as critical", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Type:   db.HealthCheckTypeHTTP,
				Source: db.HealthCheckSourceAgent,
				Status: db.HealthCheckStatusCritical,
			},
		}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusCritical])
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Health)
	})

	t.Run("list makes the agent warning when its worst instance is warning", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}, {Id: "inst-2"}}
		f.healthChecks.checks["inst-2"] = []*db.HealthCheck{
			{
				Type:   db.HealthCheckTypeHTTP,
				Source: db.HealthCheckSourceAgent,
				Status: db.HealthCheckStatusWarning,
			},
		}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusPassing])
		assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusWarning])
		assert.Equal(t, db.AgentInstanceStatusWarning, result.Agents[0].Health)
	})

	t.Run("list makes the agent unknown when its worst instance is unknown", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Type:   db.HealthCheckTypeHTTP,
				Source: db.HealthCheckSourceAgent,
				Status: db.HealthCheckStatusUnknown,
			},
		}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusUnknown])
		assert.Equal(t, db.AgentInstanceStatusUnknown, result.Agents[0].Health)
	})

	t.Run("list treats an expired ttl check as critical", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		expired := time.Now().Add(-time.Minute)
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Type:         db.HealthCheckTypeTTL,
				Source:       db.HealthCheckSourceAgent,
				Status:       db.HealthCheckStatusPassing,
				TTLExpiresAt: &expired,
			},
		}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Agents[0].StatusCounts[db.AgentInstanceStatusCritical])
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Health)
	})

	t.Run("get returns not found for a missing agent", func(t *testing.T) {
		noAgent := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)
		_, err := noAgent.agent.GetAgent("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("get health status returns agent health and status counts", func(t *testing.T) {
		f := defaultAgentFixture()
		f.instances.list = []*db.AgentInstance{
			{Id: "inst-1", InstanceId: "node-1"},
			{Id: "inst-2", InstanceId: "node-2"},
		}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Type:   db.HealthCheckTypeHTTP,
				Source: db.HealthCheckSourceAgent,
				Status: db.HealthCheckStatusWarning,
			},
		}
		f.healthChecks.checks["inst-2"] = []*db.HealthCheck{
			{
				Type:   db.HealthCheckTypeHTTP,
				Source: db.HealthCheckSourceAgent,
				Status: db.HealthCheckStatusCritical,
			},
		}

		health, counts, err := f.agent.GetHealthStatus("agent-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, health)
		assert.Equal(t, 1, counts[db.AgentInstanceStatusWarning])
		assert.Equal(t, 1, counts[db.AgentInstanceStatusCritical])
		assert.Equal(t, 0, counts[db.AgentInstanceStatusPassing])

		f.instances.list = nil
		health, counts, err = f.agent.GetHealthStatus("agent-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusUnknown, health)
		assert.Equal(t, 0, counts[db.AgentInstanceStatusCritical])
	})

	t.Run("get returns the agent with its status breakdown", func(t *testing.T) {
		f := defaultAgentFixture()
		f.instances.list = []*db.AgentInstance{
			{Id: "inst-1", InstanceId: "node-1", Address: "10.0.0.12", Port: 8080},
		}

		response, err := f.agent.GetAgent("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, db.Id("agent-1"), response.Id)
		assert.Equal(t, "1.0.0", response.Version)
		assert.Equal(t, db.AgentInstanceStatusPassing, response.Health)
		assert.Equal(t, 1, response.StatusCounts[db.AgentInstanceStatusPassing])
	})

	t.Run("upsert rejects invalid names", func(t *testing.T) {
		req := &UpsertAgentRequest{Version: "1.0.0"}

		_, _, err := defaultAgentFixture().agent.UpsertAgent("ws-1", "-bad", req)
		assert.ErrorIs(t, err, ErrInvalidAgentName)
	})

	t.Run("upsert requires a version of at most 40 characters", func(t *testing.T) {
		f := defaultAgentFixture()

		_, _, err := f.agent.UpsertAgent("ws-1", "support-assistant", nil)
		assert.ErrorIs(t, err, ErrAgentVersionRequired)

		_, _, err = f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{})
		assert.ErrorIs(t, err, ErrAgentVersionRequired)

		_, _, err = f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: strings.Repeat("v", 41)},
		)
		assert.ErrorIs(t, err, ErrAgentVersionRequired)
	})

	t.Run("upsert rejects a card path that is not absolute", func(t *testing.T) {
		_, _, err := defaultAgentFixture().agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0", CardPath: "agent.json"},
		)
		assert.ErrorIs(t, err, ErrInvalidCardPath)
	})

	t.Run("upsert creates a new agent without a card and the default card path", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		response, created, err := f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "2.1.0"},
		)
		require.NoError(t, err)
		assert.True(t, created)
		require.Len(t, f.agents.created, 1)
		assert.Empty(t, f.agents.updated)

		assert.Equal(t, "2.1.0", response.Version)
		assert.Equal(t, "/.well-known/agent-card.json", response.CardPath)
		assert.Nil(t, response.Card)
		assert.Empty(t, response.CardChecksum)
	})

	t.Run("upsert keeps a custom card path", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		response, _, err := f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0", CardPath: "/a2a/agent.json"},
		)
		require.NoError(t, err)
		assert.Equal(t, "/a2a/agent.json", response.CardPath)
	})

	t.Run("upsert keeps the card when version and card path are unchanged", func(t *testing.T) {
		f := defaultAgentFixture()

		_, created, err := f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Version: "1.0.0"})
		require.NoError(t, err)
		assert.False(t, created)
		require.Len(t, f.agents.updated, 1)
		assert.Equal(t, `{"name":"Support"}`, lo.FromPtr(f.agents.updated[0].Card))
	})

	t.Run("upsert clears the card when the version changes", func(t *testing.T) {
		f := defaultAgentFixture()

		_, created, err := f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Version: "3.0.0"})
		require.NoError(t, err)
		assert.False(t, created)
		require.Len(t, f.agents.updated, 1)
		assert.Empty(t, f.agents.created)
		assert.Equal(t, "3.0.0", f.agents.updated[0].Version)
		assert.Nil(t, f.agents.updated[0].Card)
		assert.Nil(t, f.agents.updated[0].CardChecksum)
	})

	t.Run("upsert clears the card when the card path changes", func(t *testing.T) {
		f := defaultAgentFixture()

		_, _, err := f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0", CardPath: "/agent.json"},
		)
		require.NoError(t, err)
		require.Len(t, f.agents.updated, 1)
		assert.Equal(t, "/agent.json", f.agents.updated[0].CardPath)
		assert.Nil(t, f.agents.updated[0].Card)
	})

	t.Run("upsert rejects an unknown load balancing strategy", func(t *testing.T) {
		_, _, err := defaultAgentFixture().agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0", LbStrategy: "weighted"},
		)
		assert.ErrorIs(t, err, ErrInvalidLbStrategy)
	})

	t.Run("upsert defaults a new agent to round robin", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		response, created, err := f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0"},
		)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "round_robin", response.LbStrategy)
		require.Len(t, f.agents.created, 1)
		assert.Equal(t, "round_robin", f.agents.created[0].LbStrategy)
	})

	t.Run("upsert stores a custom strategy and keeps it when omitted", func(t *testing.T) {
		f := defaultAgentFixture()

		response, _, err := f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.0", LbStrategy: "least_outstanding_tasks"},
		)
		require.NoError(t, err)
		assert.Equal(t, "least_outstanding_tasks", response.LbStrategy)
		require.Len(t, f.agents.updated, 1)
		assert.Equal(t, "least_outstanding_tasks", f.agents.updated[0].LbStrategy)

		f.agents.agent.LbStrategy = "least_outstanding_tasks"
		response, _, err = f.agent.UpsertAgent(
			"ws-1",
			"support-assistant",
			&UpsertAgentRequest{Version: "1.0.1"},
		)
		require.NoError(t, err)
		assert.Equal(t, "least_outstanding_tasks", response.LbStrategy)
	})

	t.Run("delete returns not found for a missing agent", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		err := f.agent.DeleteAgent("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("delete removes the agent", func(t *testing.T) {
		f := defaultAgentFixture()

		err := f.agent.DeleteAgent("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"agent-1"}, f.agents.deleted)
	})
}
