// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	IntegrationTypeWebhook   = "webhook"
	IntegrationTypeSlack     = "slack"
	IntegrationTypeDatadog   = "datadog"
	IntegrationTypeSplunk    = "splunk"
	IntegrationTypePagerDuty = "pagerduty"
)

// Integration is the DB row for a workspace integration (webhook, Slack, etc.).
type Integration struct {
	Id          Id
	WorkspaceId Id
	Type        string
	Name        string
	Config      *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IntegrationRepository is the interface for integration CRUD.
type IntegrationRepository interface {
	Create(integration *Integration) error
	GetById(id Id) (*Integration, error)
	Update(integration *Integration) error
	Delete(id Id) error
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Integration, error)
	ListActiveByEvent(workspaceId Id, event string) ([]*Integration, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

// IntegrationRepositoryPostgres implements IntegrationRepository for PostgreSQL
type IntegrationRepositoryPostgres struct {
	db *sql.DB
}

// IntegrationMeta is a single row in the integrations_meta table.
type IntegrationMeta struct {
	Id            Id
	IntegrationId Id
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IntegrationMetaRepository is the interface for integration metadata CRUD.
type IntegrationMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*IntegrationMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByIntegrationId(id Id) ([]*IntegrationMeta, error)
	Upsert(id Id, key, value string) error
}

type IntegrationMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewIntegrationRepository returns the repository for the current driver
func NewIntegrationRepository(db *sql.DB) IntegrationRepository {
	return &IntegrationRepositoryPostgres{db: db}
}

// scanIntegration maps a row into an integration. Column order must match the select list.
func scanIntegration(inv *Integration) []any {
	return []any{
		&inv.Id,
		&inv.WorkspaceId,
		&inv.Type,
		&inv.Name,
		&inv.Config,
		&inv.IsActive,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts a row.
func (r *IntegrationRepositoryPostgres) Create(integration *Integration) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	integration.Id = id

	err = r.db.QueryRow(
		`INSERT INTO integrations (id, workspace_id, type, name, config, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		integration.Id.String(),
		integration.WorkspaceId.String(),
		integration.Type,
		integration.Name,
		integration.Config,
		integration.IsActive,
	).Scan(&integration.CreatedAt, &integration.UpdatedAt)

	return err
}

// GetById returns a integration by Id.
func (r *IntegrationRepositoryPostgres) GetById(id Id) (*Integration, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, type, name, config, is_active, created_at, updated_at
		FROM integrations
		WHERE id = $1`,
		scanIntegration,
		id.String(),
	)
}

// Update updates a integration.
func (r *IntegrationRepositoryPostgres) Update(integration *Integration) error {
	_, err := r.db.Exec(
		`UPDATE integrations
		SET
			workspace_id = $1,
			type = $2,
			name = $3,
			config = $4,
			is_active = $5,
			updated_at = $6
		WHERE id = $7`,
		integration.WorkspaceId.String(),
		integration.Type,
		integration.Name,
		integration.Config,
		integration.IsActive,
		time.Now().UTC(),
		integration.Id.String(),
	)

	return err
}

// Delete removes a integration.
func (r *IntegrationRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations WHERE id = $1`,
		id.String(),
	)

	return err
}

// ListByWorkspaceId returns a list of integrations by workspace Id.
func (r *IntegrationRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Integration, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, type, name, config, is_active, created_at, updated_at
		FROM integrations
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanIntegration,
		workspaceId.String(),
		limit,
		offset,
	)
}

// ListActiveByEvent returns the active integrations of a workspace subscribed to an event.
func (r *IntegrationRepositoryPostgres) ListActiveByEvent(workspaceId Id, event string) ([]*Integration, error) {
	return QueryMany(
		r.db,
		`SELECT id, workspace_id, type, name, config, is_active, created_at, updated_at
		FROM integrations
		WHERE workspace_id = $1 AND is_active = TRUE AND config->'events' @> to_jsonb($2::text)
		ORDER BY created_at`,
		scanIntegration,
		workspaceId.String(),
		event,
	)
}

// CountByWorkspaceId returns the total number of integrations by workspace Id.
func (r *IntegrationRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM integrations
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// NewIntegrationMetaRepository returns the repository for integration metadata.
func NewIntegrationMetaRepository(db *sql.DB) IntegrationMetaRepository {
	return &IntegrationMetaRepositoryPostgres{db: db}
}

// scanIntegrationMeta maps a row into integration metadata. Column order must match the select list.
func scanIntegrationMeta(meta *IntegrationMeta) []any {
	return []any{
		&meta.Id,
		&meta.IntegrationId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts an integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO integrations_meta (id, integration_id, key, value)
		VALUES ($1, $2, $3, $4)`,
		metaId.String(), id.String(), key, value,
	)

	return err
}

// Get returns integration metadata by key.
func (r *IntegrationMetaRepositoryPostgres) Get(id Id, key string) (*IntegrationMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, integration_id, key, value, created_at, updated_at
		FROM integrations_meta
		WHERE integration_id = $1 AND key = $2`,
		scanIntegrationMeta,
		id.String(), key,
	)
}

// Update updates an existing integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE integrations_meta
		SET value = $1, updated_at = $2
		WHERE integration_id = $3 AND key = $4`,
		value, time.Now().UTC(), id.String(), key,
	)

	return err
}

// Delete deletes an integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations_meta
		WHERE integration_id = $1 AND key = $2`,
		id.String(), key,
	)

	return err
}

// ListByIntegrationId lists integration metadata rows by integration id.
func (r *IntegrationMetaRepositoryPostgres) ListByIntegrationId(id Id) ([]*IntegrationMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, integration_id, key, value, created_at, updated_at
		FROM integrations_meta
		WHERE integration_id = $1
		ORDER BY key`,
		scanIntegrationMeta,
		id.String(),
	)
}

// Upsert creates or updates integration metadata.
func (r *IntegrationMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
