// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	IntentionActionAllow = "allow"
	IntentionActionDeny  = "deny"
)

// Intention is a single row in the intentions table.
type Intention struct {
	Id               Id
	WorkspaceId      Id
	SourceAgent      string
	DestinationAgent string
	Skill            *string
	Action           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IntentionRepository is the interface for intention persistence.
type IntentionRepository interface {
	Create(intention *Intention) error
	GetById(id Id) (*Intention, error)
	Update(intention *Intention) error
	Delete(id Id) error
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Intention, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	Match(workspaceId Id, source, destination, skill string) (*Intention, error)
}

type IntentionRepositoryPostgres struct {
	db *sql.DB
}

// IntentionMeta is a single row in the intentions_meta table.
type IntentionMeta struct {
	Id          Id
	IntentionId Id
	Key         string
	Value       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IntentionMetaRepository is the interface for intention metadata CRUD.
type IntentionMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*IntentionMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByIntentionId(id Id) ([]*IntentionMeta, error)
	Upsert(id Id, key, value string) error
}

type IntentionMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewIntentionRepository returns the repository for intentions.
func NewIntentionRepository(db *sql.DB) IntentionRepository {
	return &IntentionRepositoryPostgres{db: db}
}

// scanIntention maps a row into an intention. Column order must match the select list.
func scanIntention(intention *Intention) []any {
	return []any{
		&intention.Id,
		&intention.WorkspaceId,
		&intention.SourceAgent,
		&intention.DestinationAgent,
		&intention.Skill,
		&intention.Action,
		&intention.CreatedAt,
		&intention.UpdatedAt,
	}
}

// Create inserts an intention row.
func (r *IntentionRepositoryPostgres) Create(intention *Intention) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	intention.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_intentions (id, workspace_id, source_agent, destination_agent, skill, action)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		intention.Id.String(),
		intention.WorkspaceId.String(),
		intention.SourceAgent,
		intention.DestinationAgent,
		intention.Skill,
		intention.Action,
	).Scan(&intention.CreatedAt, &intention.UpdatedAt)
}

// GetById returns an intention by id.
func (r *IntentionRepositoryPostgres) GetById(id Id) (*Intention, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, source_agent, destination_agent, skill, action, created_at, updated_at
		FROM ai_intentions
		WHERE id = $1`,
		scanIntention,
		id.String(),
	)
}

// Update updates an intention rule.
func (r *IntentionRepositoryPostgres) Update(intention *Intention) error {
	_, err := r.db.Exec(
		`UPDATE ai_intentions
		SET source_agent = $1, destination_agent = $2, skill = $3, action = $4, updated_at = $5
		WHERE id = $6`,
		intention.SourceAgent,
		intention.DestinationAgent,
		intention.Skill,
		intention.Action,
		time.Now().UTC(),
		intention.Id.String(),
	)

	return err
}

// Delete removes an intention row.
func (r *IntentionRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_intentions WHERE id = $1`, id.String())

	return err
}

// ListByWorkspaceId lists intentions in a workspace.
func (r *IntentionRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Intention, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, source_agent, destination_agent, skill, action, created_at, updated_at
		FROM ai_intentions
		WHERE workspace_id = $1
		ORDER BY source_agent, destination_agent, skill
		LIMIT $2 OFFSET $3`,
		scanIntention,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns the number of intentions in a workspace.
func (r *IntentionRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ai_intentions WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// Match returns the intention the gateway should apply, preferring a row for the named skill.
func (r *IntentionRepositoryPostgres) Match(workspaceId Id, source, destination, skill string) (*Intention, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, source_agent, destination_agent, skill, action, created_at, updated_at
		FROM ai_intentions
		WHERE workspace_id = $1
			AND source_agent = $2
			AND destination_agent = $3
			AND (skill IS NULL OR skill = $4)
		ORDER BY CASE WHEN skill = $4 THEN 0 ELSE 1 END
		LIMIT 1`,
		scanIntention,
		workspaceId.String(),
		source,
		destination,
		skill,
	)
}

// NewIntentionMetaRepository returns the repository for intention metadata.
func NewIntentionMetaRepository(db *sql.DB) IntentionMetaRepository {
	return &IntentionMetaRepositoryPostgres{db: db}
}

// scanIntentionMeta maps a row into intention metadata. Column order must match the select list.
func scanIntentionMeta(meta *IntentionMeta) []any {
	return []any{
		&meta.Id,
		&meta.IntentionId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts an intention metadata row. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO ai_intentions_meta (id, intention_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns intention metadata by key. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) Get(id Id, key string) (*IntentionMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, intention_id, key, value #>> '{}', created_at, updated_at
		FROM ai_intentions_meta
		WHERE intention_id = $1 AND key = $2`,
		scanIntentionMeta,
		id.String(),
		key,
	)
}

// Update updates an existing intention metadata row. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE ai_intentions_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE intention_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes an intention metadata row. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_intentions_meta WHERE intention_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByIntentionId lists intention metadata rows. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) ListByIntentionId(id Id) ([]*IntentionMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, intention_id, key, value #>> '{}', created_at, updated_at
		FROM ai_intentions_meta
		WHERE intention_id = $1
		ORDER BY key`,
		scanIntentionMeta,
		id.String(),
	)
}

// Upsert creates or updates intention metadata. TODO: remove if not used in future.
func (r *IntentionMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
