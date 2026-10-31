// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AgentProtectionTypeAPIKey    = "api_key"
	AgentProtectionTypeBasicAuth = "basic_auth"
)

// AgentProtection is a single row in the agent_protection table.
type AgentProtection struct {
	Id        Id
	AgentId   Id
	Type      string
	Config    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentProtectionRepository is the interface for agent protection CRUD.
type AgentProtectionRepository interface {
	Create(protection *AgentProtection) error
	GetById(id Id) (*AgentProtection, error)
	GetByAgentId(agentId Id) (*AgentProtection, error)
	Update(protection *AgentProtection) error
	Delete(id Id) error
}

type AgentProtectionRepositoryPostgres struct {
	db *sql.DB
}

// AgentProtectionToken is a single row in the agent_protection_tokens table.
type AgentProtectionToken struct {
	Id           Id
	ProtectionId Id
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	Meta         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AgentProtectionTokenRepository is the interface for issued protection token CRUD.
type AgentProtectionTokenRepository interface {
	Create(token *AgentProtectionToken) error
	GetById(id Id) (*AgentProtectionToken, error)
	GetByTokenHash(tokenHash string) (*AgentProtectionToken, error)
	ListByProtectionId(protectionId Id) ([]*AgentProtectionToken, error)
	Revoke(id Id) error
	DeleteExpired() (int64, error)
}

type AgentProtectionTokenRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentProtectionRepository returns the repository for the agent_protection table.
func NewAgentProtectionRepository(db *sql.DB) AgentProtectionRepository {
	return &AgentProtectionRepositoryPostgres{db: db}
}

// scanAgentProtection maps a row into an agent protection. Column order must match the select list.
func scanAgentProtection(protection *AgentProtection) []any {
	return []any{
		&protection.Id,
		&protection.AgentId,
		&protection.Type,
		&protection.Config,
		&protection.IsActive,
		&protection.CreatedAt,
		&protection.UpdatedAt,
	}
}

// Create inserts an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Create(protection *AgentProtection) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	protection.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_agent_protection (id, agent_id, type, config, is_active)
		VALUES ($1, $2, $3, $4::jsonb, $5)
		RETURNING created_at, updated_at`,
		protection.Id.String(),
		protection.AgentId.String(),
		protection.Type,
		protection.Config,
		protection.IsActive,
	).Scan(&protection.CreatedAt, &protection.UpdatedAt)
}

// GetById returns an agent protection by id.
func (r *AgentProtectionRepositoryPostgres) GetById(id Id) (*AgentProtection, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, type, config, is_active, created_at, updated_at
		FROM ai_agent_protection
		WHERE id = $1`,
		scanAgentProtection,
		id.String(),
	)
}

// GetByAgentId returns the protection of an agent.
func (r *AgentProtectionRepositoryPostgres) GetByAgentId(agentId Id) (*AgentProtection, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, type, config, is_active, created_at, updated_at
		FROM ai_agent_protection
		WHERE agent_id = $1`,
		scanAgentProtection,
		agentId.String(),
	)
}

// Update updates an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Update(protection *AgentProtection) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE ai_agent_protection
		SET type = $1, config = $2::jsonb, is_active = $3, updated_at = $4
		WHERE id = $5`,
		protection.Type,
		protection.Config,
		protection.IsActive,
		now,
		protection.Id.String(),
	)

	protection.UpdatedAt = now

	return err
}

// Delete removes an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_agent_protection WHERE id = $1`, id.String())

	return err
}

// NewAgentProtectionTokenRepository returns the repository for agent_protection_tokens.
func NewAgentProtectionTokenRepository(db *sql.DB) AgentProtectionTokenRepository {
	return &AgentProtectionTokenRepositoryPostgres{db: db}
}

// scanAgentProtectionToken maps a row into a protection token. Column order must match the select list.
func scanAgentProtectionToken(token *AgentProtectionToken) []any {
	return []any{
		&token.Id,
		&token.ProtectionId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.Meta,
		&token.CreatedAt,
		&token.UpdatedAt,
	}
}

// Create inserts an agent protection token row. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) Create(token *AgentProtectionToken) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	token.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_agent_protection_tokens
		(id, protection_id, token_hash, expires_at, revoked_at, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		token.Id.String(),
		token.ProtectionId.String(),
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.Meta,
	).Scan(&token.CreatedAt, &token.UpdatedAt)
}

// GetById returns an agent protection token by id. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) GetById(id Id) (*AgentProtectionToken, error) {
	return QueryOne(
		r.db,
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM ai_agent_protection_tokens
		WHERE id = $1`,
		scanAgentProtectionToken,
		id.String(),
	)
}

// GetByTokenHash returns a non-revoked, non-expired token by hash. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) GetByTokenHash(tokenHash string) (*AgentProtectionToken, error) {
	return QueryOne(
		r.db,
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM ai_agent_protection_tokens
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > $2`,
		scanAgentProtectionToken,
		tokenHash,
		time.Now().UTC(),
	)
}

// ListByProtectionId lists tokens for a protection row. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) ListByProtectionId(protectionId Id) ([]*AgentProtectionToken, error) {
	return QueryMany(
		r.db,
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM ai_agent_protection_tokens
		WHERE protection_id = $1
		ORDER BY created_at DESC`,
		scanAgentProtectionToken,
		protectionId.String(),
	)
}

// Revoke marks a token as revoked. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) Revoke(id Id) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE ai_agent_protection_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE id = $2 AND revoked_at IS NULL`,
		now,
		id.String(),
	)

	return err
}

// DeleteExpired removes expired token rows. TODO: remove if not used in future.
func (r *AgentProtectionTokenRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM ai_agent_protection_tokens WHERE expires_at <= $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
