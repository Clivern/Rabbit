// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAgentRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewAgentRepository(database)

	t.Run("create and get", func(t *testing.T) {
		agent := createTestAgent(t, database, workspace.Id, "invoice-extractor")

		got, err := repo.GetById(agent.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, agent.Name, got.Name)
		assert.Equal(t, workspace.Id, got.WorkspaceId)
		assert.Equal(t, agent.CardChecksum, got.CardChecksum)
		assert.Equal(t, "1.0.0", got.Version)
		assert.Equal(t, "/.well-known/agent-card.json", got.CardPath)
		assert.Equal(t, "round_robin", got.LbStrategy)
		require.NotNil(t, got.Card)
		assert.Contains(t, *got.Card, "invoice-extractor")

		byName, err := repo.GetByWorkspaceAndName(workspace.Id, "invoice-extractor")
		require.NoError(t, err)
		require.NotNil(t, byName)
		assert.Equal(t, agent.Id, byName.Id)
	})

	t.Run("missing row", func(t *testing.T) {
		id, err := NewId()
		require.NoError(t, err)

		got, err := repo.GetById(id)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = repo.GetByWorkspaceAndName(workspace.Id, "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("update list count and delete", func(t *testing.T) {
		agent := createTestAgent(t, database, workspace.Id, "support-assistant")
		createTestAgent(t, database, workspace.Id, "other-agent")

		agent.Version = "1.1.0"
		agent.Card = lo.ToPtr(`{"name":"support-assistant","version":"1.1.0"}`)
		agent.CardChecksum = lo.ToPtr("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
		agent.CardPath = "/agent.json"
		agent.LbStrategy = "sticky"
		require.NoError(t, repo.Update(agent))

		got, err := repo.GetById(agent.Id)
		require.NoError(t, err)
		assert.Equal(t, "1.1.0", got.Version)
		assert.Equal(t, agent.CardChecksum, got.CardChecksum)
		assert.Equal(t, "/agent.json", got.CardPath)
		assert.Equal(t, "sticky", got.LbStrategy)

		list, err := repo.ListByWorkspaceId(workspace.Id, 50, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))

		require.NoError(t, repo.Delete(agent.Id))
		got, err = repo.GetById(agent.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("unique name per workspace", func(t *testing.T) {
		createTestAgent(t, database, workspace.Id, "unique-agent")
		duplicate := &Agent{
			WorkspaceId: workspace.Id,
			Name:        "unique-agent",
			CardPath:    "/.well-known/agent-card.json",
			Version:     "1.0.0",
		}
		assert.Error(t, repo.Create(duplicate))
	})

	t.Run("an agent can be created without a card", func(t *testing.T) {
		agent := &Agent{
			WorkspaceId: workspace.Id,
			Name:        "pending-agent",
			CardPath:    "/.well-known/agent-card.json",
			Version:     "1.0.0",
		}
		require.NoError(t, repo.Create(agent))

		got, err := repo.GetById(agent.Id)
		require.NoError(t, err)
		assert.Nil(t, got.Card)
		assert.Nil(t, got.CardChecksum)
	})

	t.Run("set card if missing only fills an empty card", func(t *testing.T) {
		agent := &Agent{
			WorkspaceId: workspace.Id,
			Name:        "fetched-agent",
			CardPath:    "/.well-known/agent-card.json",
			Version:     "1.0.0",
		}
		require.NoError(t, repo.Create(agent))

		stored, err := repo.SetCardIfMissing(agent.Id, `{"name":"first"}`, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
		require.NoError(t, err)
		assert.True(t, stored)

		stored, err = repo.SetCardIfMissing(agent.Id, `{"name":"second"}`, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
		require.NoError(t, err)
		assert.False(t, stored)

		got, err := repo.GetById(agent.Id)
		require.NoError(t, err)
		require.NotNil(t, got.Card)
		assert.Contains(t, *got.Card, "first")
		assert.Equal(t, "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", lo.FromPtr(got.CardChecksum))
	})

	t.Run("storing a card flattens its skills and clearing it removes them", func(t *testing.T) {
		agent := &Agent{
			WorkspaceId: workspace.Id,
			Name:        "skilled-agent",
			CardPath:    "/.well-known/agent-card.json",
			Version:     "1.0.0",
		}
		require.NoError(t, repo.Create(agent))

		card := `{
			"defaultInputModes": ["application/json"],
			"defaultOutputModes": ["application/json"],
			"skills": [
				{"id": "user.profile.get", "name": "Get user profile", "tags": ["profile"], "outputModes": ["text/plain"]},
				{"id": "user.delete"},
				{"id": "user.delete", "name": "Duplicate"},
				{"name": "No id"}
			]
		}`
		stored, err := repo.SetCardIfMissing(agent.Id, card, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
		require.NoError(t, err)
		require.True(t, stored)

		type skillRow struct {
			SkillId, Name, Tags, InputModes, OutputModes string
		}
		listSkills := func() []skillRow {
			rows, err := database.Query(
				`SELECT skill_id, name, tags::text, input_modes::text, output_modes::text
				FROM ai_agent_skills WHERE agent_id = $1 ORDER BY skill_id`,
				agent.Id.String(),
			)
			require.NoError(t, err)
			defer rows.Close()

			var list []skillRow
			for rows.Next() {
				var row skillRow
				require.NoError(t, rows.Scan(&row.SkillId, &row.Name, &row.Tags, &row.InputModes, &row.OutputModes))
				list = append(list, row)
			}

			return list
		}

		assert.Equal(t, []skillRow{
			{SkillId: "user.delete", Name: "user.delete", Tags: "[]", InputModes: `["application/json"]`, OutputModes: `["application/json"]`},
			{SkillId: "user.profile.get", Name: "Get user profile", Tags: `["profile"]`, InputModes: `["application/json"]`, OutputModes: `["text/plain"]`},
		}, listSkills())

		got, err := repo.GetById(agent.Id)
		require.NoError(t, err)
		got.Card = nil
		got.CardChecksum = nil
		require.NoError(t, repo.Update(got))
		assert.Empty(t, listSkills())
	})
}

func TestIntegrationAgentMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "meta-agent")
	repo := NewAgentMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(agent.Id, "region", "eu"))
		got, err := repo.Get(agent.Id, "region")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "eu", got.Value)
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(agent.Id, "region", "us"))
		require.NoError(t, repo.Upsert(agent.Id, "tier", "gold"))
		got, err := repo.Get(agent.Id, "region")
		require.NoError(t, err)
		assert.Equal(t, "us", got.Value)
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByAgentId(agent.Id)
		require.NoError(t, err)
		assert.Len(t, list, 2)

		require.NoError(t, repo.Delete(agent.Id, "tier"))
		got, err := repo.Get(agent.Id, "tier")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
