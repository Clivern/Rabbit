// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const AgentLbStrategyDefault = "round_robin"

// Agent is a single row in the agents table.
type Agent struct {
	Id           Id
	WorkspaceId  Id
	Name         string
	Card         *string
	CardChecksum *string
	CardPath     string
	Version      string
	LbStrategy   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AgentRepository is the interface for agent catalog CRUD.
type AgentRepository interface {
	Create(agent *Agent) error
	GetById(id Id) (*Agent, error)
	GetByWorkspaceAndName(workspaceId Id, name string) (*Agent, error)
	Update(agent *Agent) error
	SetCardIfMissing(id Id, card, checksum string) (bool, error)
	Delete(id Id) error
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Agent, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type AgentRepositoryPostgres struct {
	db *sql.DB
}

// AgentMeta is a single row in the agents_meta table.
type AgentMeta struct {
	Id        Id
	AgentId   Id
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentMetaRepository is the interface for agent metadata CRUD.
type AgentMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*AgentMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByAgentId(id Id) ([]*AgentMeta, error)
	Upsert(id Id, key, value string) error
}

type AgentMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentRepository returns the repository for the agents table.
func NewAgentRepository(db *sql.DB) AgentRepository {
	return &AgentRepositoryPostgres{db: db}
}

// scanAgent maps a row into an agent. Column order must match the select list.
func scanAgent(agent *Agent) []any {
	return []any{
		&agent.Id,
		&agent.WorkspaceId,
		&agent.Name,
		&agent.Card,
		&agent.CardChecksum,
		&agent.CardPath,
		&agent.Version,
		&agent.LbStrategy,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	}
}

// Create inserts an agent row.
func (r *AgentRepositoryPostgres) Create(agent *Agent) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	agent.Id = id
	if agent.LbStrategy == "" {
		agent.LbStrategy = AgentLbStrategyDefault
	}

	return r.db.QueryRow(
		`INSERT INTO ai_agents (id, workspace_id, name, card, card_checksum, card_path, version, lb_strategy)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		agent.Id.String(),
		agent.WorkspaceId.String(),
		agent.Name,
		agent.Card,
		agent.CardChecksum,
		agent.CardPath,
		agent.Version,
		agent.LbStrategy,
	).Scan(&agent.CreatedAt, &agent.UpdatedAt)
}

// GetById returns an agent by id.
func (r *AgentRepositoryPostgres) GetById(id Id) (*Agent, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, name, card, card_checksum, card_path, version, lb_strategy, created_at, updated_at
		FROM ai_agents
		WHERE id = $1`,
		scanAgent,
		id.String(),
	)
}

// GetByWorkspaceAndName returns an agent by workspace and name.
func (r *AgentRepositoryPostgres) GetByWorkspaceAndName(workspaceId Id, name string) (*Agent, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, name, card, card_checksum, card_path, version, lb_strategy, created_at, updated_at
		FROM ai_agents
		WHERE workspace_id = $1 AND name = $2`,
		scanAgent,
		workspaceId.String(),
		name,
	)
}

// Update updates an agent card, card path, version and load balancing strategy.
func (r *AgentRepositoryPostgres) Update(agent *Agent) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE ai_agents
		SET name = $1, card = $2::jsonb, card_checksum = $3, card_path = $4, version = $5, lb_strategy = $6, updated_at = $7
		WHERE id = $8`,
		agent.Name,
		agent.Card,
		agent.CardChecksum,
		agent.CardPath,
		agent.Version,
		agent.LbStrategy,
		time.Now().UTC(),
		agent.Id.String(),
	)
	if err != nil {
		return err
	}

	err = r.SyncSkills(tx, agent.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SetCardIfMissing stores a fetched card only when the agent has none yet and reports whether it did.
func (r *AgentRepositoryPostgres) SetCardIfMissing(id Id, card, checksum string) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	result, err := tx.Exec(
		`UPDATE ai_agents
		SET card = $1::jsonb, card_checksum = $2, updated_at = $3
		WHERE id = $4 AND card IS NULL`,
		card,
		checksum,
		time.Now().UTC(),
		id.String(),
	)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return false, err
	}

	err = r.SyncSkills(tx, id)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

// SyncSkills rebuilds the agent's ai_agent_skills rows from its stored card inside tx; a missing card leaves none.
func (r *AgentRepositoryPostgres) SyncSkills(tx *sql.Tx, id Id) error {
	_, err := tx.Exec(`DELETE FROM ai_agent_skills WHERE agent_id = $1`, id.String())
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO ai_agent_skills (id, workspace_id, agent_id, skill_id, name, description, tags, input_modes, output_modes)
		SELECT
			gen_random_uuid(),
			a.workspace_id,
			a.id,
			s.skill->>'id',
			COALESCE(s.skill->>'name', s.skill->>'id'),
			s.skill->>'description',
			COALESCE(s.skill->'tags', '[]'::jsonb),
			COALESCE(s.skill->'inputModes', a.card->'defaultInputModes', '[]'::jsonb),
			COALESCE(s.skill->'outputModes', a.card->'defaultOutputModes', '[]'::jsonb)
		FROM ai_agents a
		CROSS JOIN LATERAL jsonb_array_elements(a.card->'skills') AS s(skill)
		WHERE a.id = $1 AND a.card IS NOT NULL AND s.skill->>'id' IS NOT NULL
		ON CONFLICT (agent_id, skill_id) DO NOTHING`,
		id.String(),
	)

	return err
}

// Delete removes an agent row.
func (r *AgentRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_agents WHERE id = $1`, id.String())

	return err
}

// ListByWorkspaceId lists agents in a workspace.
func (r *AgentRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Agent, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, name, card, card_checksum, card_path, version, lb_strategy, created_at, updated_at
		FROM ai_agents
		WHERE workspace_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`,
		scanAgent,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns the number of agents in a workspace.
func (r *AgentRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ai_agents WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// NewAgentMetaRepository returns the repository for agent metadata.
func NewAgentMetaRepository(db *sql.DB) AgentMetaRepository {
	return &AgentMetaRepositoryPostgres{db: db}
}

// scanAgentMeta maps a row into agent metadata. Column order must match the select list.
func scanAgentMeta(meta *AgentMeta) []any {
	return []any{
		&meta.Id,
		&meta.AgentId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts an agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO ai_agents_meta (id, agent_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns agent metadata by key. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Get(id Id, key string) (*AgentMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, key, value #>> '{}', created_at, updated_at
		FROM ai_agents_meta
		WHERE agent_id = $1 AND key = $2`,
		scanAgentMeta,
		id.String(),
		key,
	)
}

// Update updates an existing agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE ai_agents_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE agent_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes an agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_agents_meta WHERE agent_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByAgentId lists agent metadata rows. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) ListByAgentId(id Id) ([]*AgentMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, key, value #>> '{}', created_at, updated_at
		FROM ai_agents_meta
		WHERE agent_id = $1
		ORDER BY key`,
		scanAgentMeta,
		id.String(),
	)
}

// Upsert creates or updates agent metadata. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
