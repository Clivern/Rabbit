// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// UserInvite is the DB row for an invite (by email).
type UserInvite struct {
	Id            Id
	Email         string
	Role          string
	Token         string
	Status        string
	InviterUserId Id
	WorkspaceId   Id
	ExpiresAt     time.Time
	AcceptedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UserInviteRepository is the interface for user invite CRUD.
type UserInviteRepository interface {
	Create(invite *UserInvite) error
	GetById(id Id) (*UserInvite, error)
	GetByToken(token string) (*UserInvite, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*UserInvite, error)
	ListByEmail(email string, limit, offset int) ([]*UserInvite, error)
	ListPendingByEmail(email string) ([]*UserInvite, error)
	UpdateStatus(id Id, status string, acceptedAt *time.Time) error
	MarkExpiredAsExpired() (int64, error)
	Delete(id Id) error
	Count() (int64, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	CountByEmail(email string) (int64, error)
	CountPendingByEmailInWorkspace(workspaceId Id, email string) (int64, error)
}

// UserInviteRepositoryPostgres implements UserInviteRepository for PostgreSQL
type UserInviteRepositoryPostgres struct {
	db *sql.DB
}

// NewUserInviteRepository returns the repository for the current driver
func NewUserInviteRepository(db *sql.DB) UserInviteRepository {
	return &UserInviteRepositoryPostgres{db: db}
}

// scanUserInvite maps a row into a user invite. Column order must match the select list.
func scanUserInvite(u *UserInvite) []any {
	return []any{
		&u.Id,
		&u.Email,
		&u.Role,
		&u.Token,
		&u.Status,
		&u.InviterUserId,
		&u.WorkspaceId,
		&u.ExpiresAt,
		&u.AcceptedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a row
func (r *UserInviteRepositoryPostgres) Create(invite *UserInvite) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	invite.Id = id

	_, err = r.db.Exec(
		`INSERT INTO user_invites
		(id, email, role, token, status, inviter_user_id, workspace_id, expires_at, accepted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		invite.Id.String(),
		invite.Email,
		invite.Role,
		invite.Token,
		invite.Status,
		invite.InviterUserId.String(),
		invite.WorkspaceId.String(),
		invite.ExpiresAt,
		invite.AcceptedAt,
	)

	return err
}

// GetById returns a user invite by Id
func (r *UserInviteRepositoryPostgres) GetById(id Id) (*UserInvite, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE id = $1`,
		scanUserInvite,
		id.String(),
	)
}

// GetByToken returns a user invite by token. TODO: remove if not used in future.
func (r *UserInviteRepositoryPostgres) GetByToken(token string) (*UserInvite, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE token = $1`,
		scanUserInvite,
		token,
	)
}

// ListByWorkspaceId returns a paginated list of invites for a workspace.
func (r *UserInviteRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*UserInvite, error) {
	if limit <= 0 {
		limit = 50
	}

	return QueryMany(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanUserInvite,
		workspaceId.String(),
		limit,
		offset,
	)
}

// ListByEmail returns a paginated list of invites for an email address. TODO: remove if not used in future.
func (r *UserInviteRepositoryPostgres) ListByEmail(email string, limit, offset int) ([]*UserInvite, error) {
	if limit <= 0 {
		limit = 50
	}

	return QueryMany(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanUserInvite,
		email,
		limit,
		offset,
	)
}

// ListPendingByEmail returns pending, unexpired invites for an email address.
func (r *UserInviteRepositoryPostgres) ListPendingByEmail(email string) ([]*UserInvite, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE LOWER(email) = LOWER($1) AND status = 'pending' AND expires_at > $2
		ORDER BY created_at ASC`,
		scanUserInvite,
		email,
		time.Now().UTC(),
	)
}

// UpdateStatus updates the status of a user invite (e.g. to "accepted" with acceptedAt set).
func (r *UserInviteRepositoryPostgres) UpdateStatus(id Id, status string, acceptedAt *time.Time) error {
	_, err := r.db.Exec(
		`UPDATE user_invites
		SET
			status = $1,
			accepted_at = $2,
			updated_at = $3
		WHERE id = $4`,
		status,
		acceptedAt,
		time.Now().UTC(),
		id.String(),
	)

	return err
}

// MarkExpiredAsExpired sets status to 'expired' for all pending invites past expires_at. Returns the number updated.
func (r *UserInviteRepositoryPostgres) MarkExpiredAsExpired() (int64, error) {
	result, err := r.db.Exec(
		`UPDATE user_invites
		SET
			status = 'expired',
			updated_at = $1
		WHERE status = 'pending' AND expires_at < $2`,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete removes a user invite
func (r *UserInviteRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM user_invites WHERE id = $1`,
		id.String(),
	)

	return err
}

// Count returns the total number of user invites.
func (r *UserInviteRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites`,
	).Scan(&count)

	return count, err
}

// CountByWorkspaceId returns the total number of user invites in a workspace.
func (r *UserInviteRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// CountByEmail returns the total number of user invites for an email address. TODO: remove if not used in future.
func (r *UserInviteRepositoryPostgres) CountByEmail(email string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites
		WHERE email = $1`,
		email,
	).Scan(&count)

	return count, err
}

// CountPendingByEmailInWorkspace returns pending invites for an email in a workspace.
func (r *UserInviteRepositoryPostgres) CountPendingByEmailInWorkspace(workspaceId Id, email string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites
		WHERE workspace_id = $1 AND email = $2 AND status = 'pending' AND expires_at > $3`,
		workspaceId.String(),
		email,
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}
