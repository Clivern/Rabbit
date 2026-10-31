// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// WorkspaceUser is a single row in the workspace_users table
type WorkspaceUser struct {
	Id          Id
	WorkspaceId Id
	UserId      Id
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceUserRepository is the interface for workspace-user membership CRUD.
type WorkspaceUserRepository interface {
	Create(m *WorkspaceUser) error
	GetById(id Id) (*WorkspaceUser, error)
	GetByWorkspaceAndUser(workspaceId, userId Id) (*WorkspaceUser, error)
	Update(m *WorkspaceUser) error
	Delete(id Id) error
	List(limit, offset int) ([]*WorkspaceUser, error)
	Count() (int64, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*WorkspaceUser, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

// WorkspaceUserRepositoryPostgres implements WorkspaceUserRepository for PostgreSQL
type WorkspaceUserRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceUserRepository returns the repository for the current driver
func NewWorkspaceUserRepository(db *sql.DB) WorkspaceUserRepository {
	return &WorkspaceUserRepositoryPostgres{db: db}
}

// scanWorkspaceUser maps a row into a workspace user. Column order must match the select list.
func scanWorkspaceUser(m *WorkspaceUser) []any {
	return []any{
		&m.Id,
		&m.WorkspaceId,
		&m.UserId,
		&m.Role,
		&m.CreatedAt,
		&m.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a row
func (r *WorkspaceUserRepositoryPostgres) Create(m *WorkspaceUser) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	m.Id = id

	err = r.db.QueryRow(
		`INSERT INTO workspace_users (id, workspace_id, user_id, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`,
		m.Id.String(),
		m.WorkspaceId.String(),
		m.UserId.String(),
		m.Role,
	).Scan(&m.CreatedAt, &m.UpdatedAt)

	return err
}

// GetById returns a workspace user by Id. TODO: remove if not used in future.
func (r *WorkspaceUserRepositoryPostgres) GetById(id Id) (*WorkspaceUser, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE id = $1`,
		scanWorkspaceUser,
		id.String(),
	)
}

// GetByWorkspaceAndUser returns a workspace user by workspace and user Id
func (r *WorkspaceUserRepositoryPostgres) GetByWorkspaceAndUser(workspaceId, userId Id) (*WorkspaceUser, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users WHERE workspace_id = $1 AND user_id = $2`,
		scanWorkspaceUser,
		workspaceId.String(),
		userId.String(),
	)
}

// Update updates a workspace user
func (r *WorkspaceUserRepositoryPostgres) Update(m *WorkspaceUser) error {
	_, err := r.db.Exec(
		`UPDATE workspace_users SET workspace_id = $1, user_id = $2, role = $3, updated_at = $4
		WHERE id = $5`,
		m.WorkspaceId.String(),
		m.UserId.String(),
		m.Role,
		time.Now().UTC(),
		m.Id.String(),
	)

	return err
}

// Delete removes a workspace user
func (r *WorkspaceUserRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_users WHERE id = $1`,
		id.String(),
	)

	return err
}

// List returns a list of workspace users. TODO: remove if not used in future.
func (r *WorkspaceUserRepositoryPostgres) List(limit, offset int) ([]*WorkspaceUser, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		scanWorkspaceUser,
		limit,
		offset,
	)
}

// Count returns the total number of workspace users. TODO: remove if not used in future.
func (r *WorkspaceUserRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_users`,
	).Scan(&count)

	return count, err
}

// ListByWorkspaceId returns workspace users for the given workspace with pagination
func (r *WorkspaceUserRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*WorkspaceUser, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, user_id, role, created_at, updated_at
		FROM workspace_users
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanWorkspaceUser,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns the number of users in the workspace
func (r *WorkspaceUserRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspace_users
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
