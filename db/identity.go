// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AgentIdentityTypeAPIKey            = "api_key"
	AgentIdentityTypeBasicAuth         = "basic_auth"
	AgentIdentityTypeClientCredentials = "client_credentials"
)

// AgentIdentity is a single row in the agent_identity table.
type AgentIdentity struct {
	Id        Id
	AgentId   Id
	Name      string
	Type      string
	Config    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentIdentityRepository is the interface for agent identity CRUD.
type AgentIdentityRepository interface {
	Create(identity *AgentIdentity) error
	GetById(id Id) (*AgentIdentity, error)
	GetByAgentAndName(agentId Id, name string) (*AgentIdentity, error)
	Update(identity *AgentIdentity) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentIdentity, error)
	ListActiveByAgentId(agentId Id) ([]*AgentIdentity, error)
}

type AgentIdentityRepositoryPostgres struct {
	db *sql.DB
}

// AgentIdentityToken is a single row in the agent_identity_tokens table.
type AgentIdentityToken struct {
	Id         Id
	IdentityId Id
	TokenHash  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// AgentIdentityTokenRepository is the interface for tokens issued to client credentials identities.
type AgentIdentityTokenRepository interface {
	Create(token *AgentIdentityToken) error
	GetByTokenHash(tokenHash string) (*AgentIdentityToken, error)
	DeleteExpired() (int64, error)
}

type AgentIdentityTokenRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentIdentityRepository returns the repository for the agent_identity table.
func NewAgentIdentityRepository(db *sql.DB) AgentIdentityRepository {
	return &AgentIdentityRepositoryPostgres{db: db}
}

// scanAgentIdentity maps a row into an agent identity. Column order must match the select list.
func scanAgentIdentity(identity *AgentIdentity) []any {
	return []any{
		&identity.Id,
		&identity.AgentId,
		&identity.Name,
		&identity.Type,
		&identity.Config,
		&identity.IsActive,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	}
}

// Create inserts an agent identity row.
func (r *AgentIdentityRepositoryPostgres) Create(identity *AgentIdentity) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	identity.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_agent_identity (id, agent_id, name, type, config, is_active)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING created_at, updated_at`,
		identity.Id.String(),
		identity.AgentId.String(),
		identity.Name,
		identity.Type,
		identity.Config,
		identity.IsActive,
	).Scan(&identity.CreatedAt, &identity.UpdatedAt)
}

// GetById returns an agent identity by id.
func (r *AgentIdentityRepositoryPostgres) GetById(id Id) (*AgentIdentity, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM ai_agent_identity
		WHERE id = $1`,
		scanAgentIdentity,
		id.String(),
	)
}

// GetByAgentAndName returns an agent identity by agent and name.
func (r *AgentIdentityRepositoryPostgres) GetByAgentAndName(agentId Id, name string) (*AgentIdentity, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM ai_agent_identity
		WHERE agent_id = $1 AND name = $2`,
		scanAgentIdentity,
		agentId.String(),
		name,
	)
}

// Update updates an agent identity row.
func (r *AgentIdentityRepositoryPostgres) Update(identity *AgentIdentity) error {
	_, err := r.db.Exec(
		`UPDATE ai_agent_identity
		SET name = $1, type = $2, config = $3::jsonb, is_active = $4, updated_at = $5
		WHERE id = $6`,
		identity.Name,
		identity.Type,
		identity.Config,
		identity.IsActive,
		time.Now().UTC(),
		identity.Id.String(),
	)

	return err
}

// Delete removes an agent identity row.
func (r *AgentIdentityRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_agent_identity WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists the identity rows of an agent.
func (r *AgentIdentityRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentIdentity, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM ai_agent_identity
		WHERE agent_id = $1
		ORDER BY name`,
		scanAgentIdentity,
		agentId.String(),
	)
}

// ListActiveByAgentId lists the active identity rows of an agent. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) ListActiveByAgentId(agentId Id) ([]*AgentIdentity, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM ai_agent_identity
		WHERE agent_id = $1 AND is_active = TRUE
		ORDER BY name`,
		scanAgentIdentity,
		agentId.String(),
	)
}

// NewAgentIdentityTokenRepository returns the repository for the agent_identity_tokens table.
func NewAgentIdentityTokenRepository(db *sql.DB) AgentIdentityTokenRepository {
	return &AgentIdentityTokenRepositoryPostgres{db: db}
}

// scanAgentIdentityToken maps a row into an issued identity token. Column order must match the select list.
func scanAgentIdentityToken(token *AgentIdentityToken) []any {
	return []any{
		&token.Id,
		&token.IdentityId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	}
}

// Create inserts an issued token row.
func (r *AgentIdentityTokenRepositoryPostgres) Create(token *AgentIdentityToken) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	token.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_agent_identity_tokens (id, identity_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		token.Id.String(),
		token.IdentityId.String(),
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)
}

// GetByTokenHash returns an issued token by the SHA-256 hash of its value.
func (r *AgentIdentityTokenRepositoryPostgres) GetByTokenHash(tokenHash string) (*AgentIdentityToken, error) {
	return QueryOne(
		r.db,
		`SELECT id, identity_id, token_hash, expires_at, created_at
		FROM ai_agent_identity_tokens
		WHERE token_hash = $1`,
		scanAgentIdentityToken,
		tokenHash,
	)
}

// DeleteExpired removes issued tokens past their expiry.
func (r *AgentIdentityTokenRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM ai_agent_identity_tokens WHERE expires_at <= $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
