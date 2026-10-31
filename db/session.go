// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Session is the DB row for a user session.
type Session struct {
	Id        Id
	Token     string
	UserId    Id
	IPAddress *string
	UserAgent *string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SessionRepository is the interface for session CRUD.
type SessionRepository interface {
	Create(session *Session) error
	GetByToken(token string) (*Session, error)
	GetById(id Id) (*Session, error)
	GetByUserId(userId Id) ([]*Session, error)
	Delete(id Id) error
	DeleteByToken(token string) error
	DeleteByUserId(userId Id) error
	DeleteExpired() (int64, error)
	IsValid(token string) (bool, error)
	UpdateExpiration(id Id, expiresAt time.Time) error
	Count() (int64, error)
	CountByUserId(userId Id) (int64, error)
}

// SessionRepositoryPostgres implements SessionRepository for PostgreSQL
type SessionRepositoryPostgres struct {
	db *sql.DB
}

// NewSessionRepository returns the repository for the current driver
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &SessionRepositoryPostgres{db: db}
}

// scanSession maps a row into a session. Column order must match the select list.
func scanSession(s *Session) []any {
	return []any{
		&s.Id,
		&s.Token,
		&s.UserId,
		&s.IPAddress,
		&s.UserAgent,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a row
func (r *SessionRepositoryPostgres) Create(session *Session) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	session.Id = id

	_, err = r.db.Exec(
		`INSERT INTO user_sessions
		(id, token, user_id, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		session.Id.String(),
		session.Token,
		session.UserId.String(),
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
	)

	return err
}

// GetByToken returns a session by token
func (r *SessionRepositoryPostgres) GetByToken(token string) (*Session, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, token, user_id, ip_address, user_agent,
			expires_at, created_at, updated_at
		FROM user_sessions
		WHERE token = $1`,
		scanSession,
		token,
	)
}

// GetById returns a session by Id. TODO: remove if not used in future.
func (r *SessionRepositoryPostgres) GetById(id Id) (*Session, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, token, user_id, ip_address, user_agent,
			expires_at, created_at, updated_at
		FROM user_sessions
		WHERE id = $1`,
		scanSession,
		id.String(),
	)
}

// GetByUserId returns a list of sessions by user Id
func (r *SessionRepositoryPostgres) GetByUserId(userId Id) ([]*Session, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, token, user_id, ip_address, user_agent,
			expires_at, created_at, updated_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		scanSession,
		userId.String(),
	)
}

// Delete removes a session
func (r *SessionRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM user_sessions WHERE id = $1`,
		id.String(),
	)

	return err
}

// DeleteByToken removes a session by token. TODO: remove if not used in future.
func (r *SessionRepositoryPostgres) DeleteByToken(token string) error {
	_, err := r.db.Exec(
		`DELETE FROM user_sessions WHERE token = $1`,
		token,
	)

	return err
}

// DeleteByUserId removes a session by user Id
func (r *SessionRepositoryPostgres) DeleteByUserId(userId Id) error {
	_, err := r.db.Exec(
		`DELETE FROM user_sessions WHERE user_id = $1`,
		userId.String(),
	)

	return err
}

// DeleteExpired removes expired sessions
func (r *SessionRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM user_sessions
		WHERE expires_at < $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// IsValid checks if a session is valid. TODO: remove if not used in future.
func (r *SessionRepositoryPostgres) IsValid(token string) (bool, error) {
	session, err := r.GetByToken(token)
	if err != nil {
		return false, err
	}
	if session == nil {
		return false, nil
	}

	return session.ExpiresAt.After(time.Now().UTC()), nil
}

// UpdateExpiration updates the expiration time of a session. TODO: remove if not used in future.
func (r *SessionRepositoryPostgres) UpdateExpiration(id Id, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE user_sessions
		SET
			expires_at = $1,
			updated_at = $2
		WHERE id = $3`,
		expiresAt, time.Now().UTC(), id.String(),
	)

	return err
}

// Count returns the total number of sessions
func (r *SessionRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_sessions
		WHERE expires_at > $1`,
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}

// CountByUserId returns the total number of sessions by user Id. TODO: remove if not used in future.
func (r *SessionRepositoryPostgres) CountByUserId(userId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_sessions
		WHERE user_id = $1 AND expires_at > $2`,
		userId.String(), time.Now().UTC(),
	).Scan(&count)

	return count, err
}
