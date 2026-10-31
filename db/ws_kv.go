// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// WorkspaceKeyValue is a workspace-scoped key/value row with an optional expiry.
type WorkspaceKeyValue struct {
	Id          Id
	WorkspaceId Id
	Key         string
	Value       string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceKeyValueRepository is the interface for workspace key-value persistence.
type WorkspaceKeyValueRepository interface {
	Upsert(item *WorkspaceKeyValue) error
	Get(workspaceId Id, key string) (*WorkspaceKeyValue, error)
	Delete(workspaceId Id, key string) error
	ListByPrefix(workspaceId Id, prefix string) ([]*WorkspaceKeyValue, error)
	DeleteExpired() (int64, error)
}

type WorkspaceKeyValueRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceKeyValueRepository returns the repository for workspace_kv.
func NewWorkspaceKeyValueRepository(db *sql.DB) WorkspaceKeyValueRepository {
	return &WorkspaceKeyValueRepositoryPostgres{db: db}
}

// scanWorkspaceKeyValue maps a row into a workspace key/value. Column order must match the select list.
func scanWorkspaceKeyValue(item *WorkspaceKeyValue) []any {
	return []any{
		&item.Id,
		&item.WorkspaceId,
		&item.Key,
		&item.Value,
		&item.ExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}

// Upsert inserts or replaces a workspace key/value row.
func (r *WorkspaceKeyValueRepositoryPostgres) Upsert(item *WorkspaceKeyValue) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	return r.db.QueryRow(
		`INSERT INTO workspace_kv (id, workspace_id, key, value, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (workspace_id, key) DO UPDATE SET
			value = EXCLUDED.value,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		RETURNING id, created_at, updated_at`,
		id.String(),
		item.WorkspaceId.String(),
		item.Key,
		item.Value,
		item.ExpiresAt,
	).Scan(
		&item.Id,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
}

// Get returns a non-expired value for a workspace key.
func (r *WorkspaceKeyValueRepositoryPostgres) Get(workspaceId Id, key string) (*WorkspaceKeyValue, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, key, value, expires_at, created_at, updated_at
		FROM workspace_kv
		WHERE workspace_id = $1 AND key = $2
			AND (expires_at IS NULL OR expires_at > $3)`,
		scanWorkspaceKeyValue,
		workspaceId.String(),
		key,
		time.Now().UTC(),
	)
}

// Delete removes a workspace key/value row.
func (r *WorkspaceKeyValueRepositoryPostgres) Delete(workspaceId Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_kv WHERE workspace_id = $1 AND key = $2`,
		workspaceId.String(),
		key,
	)

	return err
}

// ListByPrefix lists non-expired keys in a workspace that start with prefix.
func (r *WorkspaceKeyValueRepositoryPostgres) ListByPrefix(workspaceId Id, prefix string) ([]*WorkspaceKeyValue, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, key, value, expires_at, created_at, updated_at
		FROM workspace_kv
		WHERE workspace_id = $1
			AND key LIKE $2 ESCAPE '\'
			AND (expires_at IS NULL OR expires_at > $3)
		ORDER BY key`,
		scanWorkspaceKeyValue,
		workspaceId.String(),
		prefix+"%",
		time.Now().UTC(),
	)
}

// DeleteExpired removes rows that have passed their expiry. TODO: remove if not used in future.
func (r *WorkspaceKeyValueRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM workspace_kv WHERE expires_at IS NOT NULL AND expires_at <= $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
