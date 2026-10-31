// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// AgentCheck is a single row in the agent_checks table.
type AgentCheck struct {
	Id         Id
	AgentId    Id
	CheckId    string
	Name       string
	Type       string
	Definition string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentCheckRepository is the interface for agent check template CRUD.
type AgentCheckRepository interface {
	Create(check *AgentCheck) error
	GetByAgentAndCheckId(agentId Id, checkId string) (*AgentCheck, error)
	Update(check *AgentCheck) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentCheck, error)
	ListPageByAgentId(agentId Id, limit, offset int) ([]*AgentCheck, error)
	CountByAgentId(agentId Id) (int64, error)
}

type AgentCheckRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentCheckRepository returns the repository for agent check templates.
func NewAgentCheckRepository(db *sql.DB) AgentCheckRepository {
	return &AgentCheckRepositoryPostgres{db: db}
}

// scanAgentCheck maps a row into an agent check template. Column order must match the select list.
func scanAgentCheck(check *AgentCheck) []any {
	return []any{
		&check.Id,
		&check.AgentId,
		&check.CheckId,
		&check.Name,
		&check.Type,
		&check.Definition,
		&check.CreatedAt,
		&check.UpdatedAt,
	}
}

// Create inserts an agent check template.
func (r *AgentCheckRepositoryPostgres) Create(check *AgentCheck) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	check.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_agent_checks (id, agent_id, check_id, name, type, definition)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING created_at, updated_at`,
		check.Id.String(),
		check.AgentId.String(),
		check.CheckId,
		check.Name,
		check.Type,
		check.Definition,
	).Scan(&check.CreatedAt, &check.UpdatedAt)
}

// GetByAgentAndCheckId returns an agent check template by its check id.
func (r *AgentCheckRepositoryPostgres) GetByAgentAndCheckId(agentId Id, checkId string) (*AgentCheck, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, check_id, name, type, definition, created_at, updated_at
		FROM ai_agent_checks
		WHERE agent_id = $1 AND check_id = $2`,
		scanAgentCheck,
		agentId.String(),
		checkId,
	)
}

// Update updates an agent check template.
func (r *AgentCheckRepositoryPostgres) Update(check *AgentCheck) error {
	check.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE ai_agent_checks
		SET name = $1, type = $2, definition = $3::jsonb, updated_at = $4
		WHERE id = $5`,
		check.Name,
		check.Type,
		check.Definition,
		check.UpdatedAt,
		check.Id.String(),
	)

	return err
}

// Delete removes an agent check template.
func (r *AgentCheckRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_agent_checks WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists the check templates of an agent.
func (r *AgentCheckRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentCheck, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, check_id, name, type, definition, created_at, updated_at
		FROM ai_agent_checks
		WHERE agent_id = $1
		ORDER BY check_id`,
		scanAgentCheck,
		agentId.String(),
	)
}

// ListPageByAgentId lists one page of the check templates of an agent.
func (r *AgentCheckRepositoryPostgres) ListPageByAgentId(agentId Id, limit, offset int) ([]*AgentCheck, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, check_id, name, type, definition, created_at, updated_at
		FROM ai_agent_checks
		WHERE agent_id = $1
		ORDER BY check_id
		LIMIT $2 OFFSET $3`,
		scanAgentCheck,
		agentId.String(),
		limit,
		offset,
	)
}

// CountByAgentId returns the number of check templates of an agent.
func (r *AgentCheckRepositoryPostgres) CountByAgentId(agentId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ai_agent_checks WHERE agent_id = $1`,
		agentId.String(),
	).Scan(&count)

	return count, err
}
