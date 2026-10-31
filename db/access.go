// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// AccessKey is the DB row for a workspace-scoped access key.
type AccessKey struct {
	Id          Id
	WorkspaceId Id
	Name        string
	Key         string
	ExpiresAt   *time.Time
	Meta        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AccessKeyRepository is the interface for workspace access key CRUD.
type AccessKeyRepository interface {
	Create(accessKey *AccessKey) error
	GetById(id Id) (*AccessKey, error)
	GetByKey(key string) (*AccessKey, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AccessKey, error)
	Delete(id Id) error
	DeleteExpired() (int64, error)
	Count() (int64, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type AccessKeyRepositoryPostgres struct {
	db *sql.DB
}

// NewAccessKeyRepository returns the repository for workspace access keys.
func NewAccessKeyRepository(db *sql.DB) AccessKeyRepository {
	return &AccessKeyRepositoryPostgres{db: db}
}

// scanAccessKey maps a row into a workspace access key. Column order must match the select list.
func scanAccessKey(item *AccessKey) []any {
	return []any{
		&item.Id,
		&item.WorkspaceId,
		&item.Name,
		&item.Key,
		&item.ExpiresAt,
		&item.Meta,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}

// Create inserts a workspace access key row.
func (r *AccessKeyRepositoryPostgres) Create(accessKey *AccessKey) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	accessKey.Id = id

	return r.db.QueryRow(
		`INSERT INTO access_keys
		(id, workspace_id, name, token, expires_at, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		accessKey.Id.String(),
		accessKey.WorkspaceId.String(),
		accessKey.Name,
		accessKey.Key,
		accessKey.ExpiresAt,
		accessKey.Meta,
	).Scan(&accessKey.CreatedAt, &accessKey.UpdatedAt)
}

// GetById returns a workspace access key by id.
func (r *AccessKeyRepositoryPostgres) GetById(id Id) (*AccessKey, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, name, token, expires_at, meta, created_at, updated_at
		FROM access_keys
		WHERE id = $1`,
		scanAccessKey,
		id.String(),
	)
}

// GetByKey returns a workspace access key by key.
func (r *AccessKeyRepositoryPostgres) GetByKey(key string) (*AccessKey, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, name, token, expires_at, meta, created_at, updated_at
		FROM access_keys
		WHERE token = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		scanAccessKey,
		key,
		time.Now().UTC(),
	)
}

// ListByWorkspaceId lists workspace access key rows by workspace id.
func (r *AccessKeyRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AccessKey, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, name, token, expires_at, meta, created_at, updated_at
		FROM access_keys
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanAccessKey,
		workspaceId.String(),
		limit,
		offset,
	)
}

// Delete deletes a workspace access key row.
func (r *AccessKeyRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM access_keys WHERE id = $1`, id.String())

	return err
}

// DeleteExpired deletes expired workspace access keys.
func (r *AccessKeyRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM access_keys
		WHERE expires_at IS NOT NULL AND expires_at < $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Count returns the total number of workspace access key rows.
func (r *AccessKeyRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM access_keys
		WHERE expires_at IS NULL OR expires_at > $1`,
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}

// CountByWorkspaceId counts workspace access key rows by workspace id.
func (r *AccessKeyRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM access_keys
		WHERE workspace_id = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		workspaceId.String(),
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}
