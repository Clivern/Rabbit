// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	HealthCheckStatusPassing  = "passing"
	HealthCheckStatusWarning  = "warning"
	HealthCheckStatusCritical = "critical"
	HealthCheckStatusUnknown  = "unknown"

	HealthCheckTypeTTL   = "ttl"
	HealthCheckTypeHTTP  = "http"
	HealthCheckTypeTCP   = "tcp"
	HealthCheckTypeA2A   = "a2a"
	HealthCheckTypeCard  = "card"
	HealthCheckTypeSkill = "skill"

	// HealthCheckSourceAgent marks a check copied from an agent check template.
	HealthCheckSourceAgent = "agent"
)

// HealthCheck is a single row in the health_checks table.
type HealthCheck struct {
	Id           Id
	InstanceId   Id
	CheckId      string
	Name         string
	Type         string
	Source       string
	Status       string
	Definition   *string
	Output       *string
	TTLExpiresAt *time.Time
	LastRunAt    *time.Time
	NextRunAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// HealthCheckRepository is the interface for health check persistence.
type HealthCheckRepository interface {
	Create(check *HealthCheck) error
	GetById(id Id) (*HealthCheck, error)
	GetByInstanceAndCheckId(instanceId Id, checkId string) (*HealthCheck, error)
	Update(check *HealthCheck) error
	Delete(id Id) error
	ListByInstanceId(instanceId Id) ([]*HealthCheck, error)
	DeleteByAgentAndCheckId(agentId Id, checkId string) error
	Pass(id Id, output string, ttlExpiresAt *time.Time) error
	Warn(id Id, output string) error
	Fail(id Id, output string) error
	Report(id Id, status, output string, ttlExpiresAt *time.Time) error
	ClaimDue(now time.Time, limit int, hold time.Duration) ([]*HealthCheck, error)
	Record(id Id, status, output string, nextRunAt time.Time) error
}

type HealthCheckRepositoryPostgres struct {
	db *sql.DB
}

// HealthCheckMeta is a single row in the health_checks_meta table.
type HealthCheckMeta struct {
	Id            Id
	HealthCheckId Id
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// HealthCheckMetaRepository is the interface for health check metadata CRUD.
type HealthCheckMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*HealthCheckMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByHealthCheckId(id Id) ([]*HealthCheckMeta, error)
	Upsert(id Id, key, value string) error
}

type HealthCheckMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewHealthCheckRepository returns the repository for health checks.
func NewHealthCheckRepository(db *sql.DB) HealthCheckRepository {
	return &HealthCheckRepositoryPostgres{db: db}
}

// scanHealthCheck maps a row into a health check. Column order must match the select list.
func scanHealthCheck(check *HealthCheck) []any {
	return []any{
		&check.Id,
		&check.InstanceId,
		&check.CheckId,
		&check.Name,
		&check.Type,
		&check.Source,
		&check.Status,
		&check.Definition,
		&check.Output,
		&check.TTLExpiresAt,
		&check.LastRunAt,
		&check.NextRunAt,
		&check.CreatedAt,
		&check.UpdatedAt,
	}
}

// Create inserts a health check row. A new check starts unknown until its first result.
func (r *HealthCheckRepositoryPostgres) Create(check *HealthCheck) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	check.Id = id
	if check.Status == "" {
		check.Status = HealthCheckStatusUnknown
	}
	if check.Source == "" {
		check.Source = HealthCheckSourceAgent
	}

	return r.db.QueryRow(
		`INSERT INTO ai_health_checks (
			id, instance_id, check_id, name, type, source, status, definition,
			output, ttl_expires_at, last_run_at, next_run_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12)
		RETURNING created_at, updated_at`,
		check.Id.String(),
		check.InstanceId.String(),
		check.CheckId,
		check.Name,
		check.Type,
		check.Source,
		check.Status,
		check.Definition,
		check.Output,
		check.TTLExpiresAt,
		check.LastRunAt,
		check.NextRunAt,
	).Scan(&check.CreatedAt, &check.UpdatedAt)
}

// GetById returns a health check by id. TODO: remove if not used in future.
func (r *HealthCheckRepositoryPostgres) GetById(id Id) (*HealthCheck, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, instance_id, check_id, name, type, source, status, definition,
			output, ttl_expires_at, last_run_at, next_run_at, created_at, updated_at
		FROM ai_health_checks
		WHERE id = $1`,
		scanHealthCheck,
		id.String(),
	)
}

// GetByInstanceAndCheckId returns a health check by instance and check id.
func (r *HealthCheckRepositoryPostgres) GetByInstanceAndCheckId(instanceId Id, checkId string) (*HealthCheck, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, instance_id, check_id, name, type, source, status, definition,
			output, ttl_expires_at, last_run_at, next_run_at, created_at, updated_at
		FROM ai_health_checks
		WHERE instance_id = $1 AND check_id = $2`,
		scanHealthCheck,
		instanceId.String(),
		checkId,
	)
}

// Update updates a health check definition and when it next runs.
func (r *HealthCheckRepositoryPostgres) Update(check *HealthCheck) error {
	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET name = $1, type = $2, definition = $3::jsonb, next_run_at = $4, updated_at = $5
		WHERE id = $6`,
		check.Name,
		check.Type,
		check.Definition,
		check.NextRunAt,
		time.Now().UTC(),
		check.Id.String(),
	)

	return err
}

// Delete removes a health check row.
func (r *HealthCheckRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_health_checks WHERE id = $1`, id.String())

	return err
}

// ListByInstanceId lists health checks for an instance.
func (r *HealthCheckRepositoryPostgres) ListByInstanceId(instanceId Id) ([]*HealthCheck, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, instance_id, check_id, name, type, source, status, definition,
			output, ttl_expires_at, last_run_at, next_run_at, created_at, updated_at
		FROM ai_health_checks
		WHERE instance_id = $1
		ORDER BY check_id`,
		scanHealthCheck,
		instanceId.String(),
	)
}

// DeleteByAgentAndCheckId removes a template check from every instance of an agent.
func (r *HealthCheckRepositoryPostgres) DeleteByAgentAndCheckId(agentId Id, checkId string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_health_checks hc
		USING ai_agent_instances ai
		WHERE hc.instance_id = ai.id
			AND ai.agent_id = $1
			AND hc.check_id = $2
			AND hc.source = $3`,
		agentId.String(),
		checkId,
		HealthCheckSourceAgent,
	)

	return err
}

// Pass marks a check passing and optionally extends a TTL.
func (r *HealthCheckRepositoryPostgres) Pass(id Id, output string, ttlExpiresAt *time.Time) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET status = $1, output = $2, ttl_expires_at = COALESCE($3, ttl_expires_at),
			last_run_at = $4, updated_at = $4
		WHERE id = $5`,
		HealthCheckStatusPassing,
		raw,
		ttlExpiresAt,
		now,
		id.String(),
	)

	return err
}

// Warn marks a check warning. The instance stays in discovery. TODO: remove if not used in future.
func (r *HealthCheckRepositoryPostgres) Warn(id Id, output string) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET status = $1, output = $2, last_run_at = $3, updated_at = $3
		WHERE id = $4`,
		HealthCheckStatusWarning,
		raw,
		now,
		id.String(),
	)

	return err
}

// Fail marks a check critical and drops the instance from discovery. TODO: remove if not used in future.
func (r *HealthCheckRepositoryPostgres) Fail(id Id, output string) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET status = $1, output = $2, last_run_at = $3, updated_at = $3
		WHERE id = $4`,
		HealthCheckStatusCritical,
		raw,
		now,
		id.String(),
	)

	return err
}

// Report sets a check status and output, optionally extending its TTL.
func (r *HealthCheckRepositoryPostgres) Report(id Id, status, output string, ttlExpiresAt *time.Time) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET status = $1, output = $2, ttl_expires_at = COALESCE($3, ttl_expires_at),
			last_run_at = $4, updated_at = $4
		WHERE id = $5`,
		status,
		raw,
		ttlExpiresAt,
		now,
		id.String(),
	)

	return err
}

// ClaimDue locks due http and tcp checks and pushes their next run out by hold. TODO: remove if not used in future.
func (r *HealthCheckRepositoryPostgres) ClaimDue(now time.Time, limit int, hold time.Duration) ([]*HealthCheck, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	list, err := QueryMany(
		tx,
		`SELECT
			id, instance_id, check_id, name, type, source, status, definition,
			output, ttl_expires_at, last_run_at, next_run_at, created_at, updated_at
		FROM ai_health_checks
		WHERE next_run_at <= $1 AND type IN ($2, $3)
		ORDER BY next_run_at
		LIMIT $4
		FOR UPDATE SKIP LOCKED`,
		scanHealthCheck,
		now,
		HealthCheckTypeHTTP,
		HealthCheckTypeTCP,
		limit,
	)
	if err != nil {
		return nil, err
	}

	held := now.Add(hold)
	for _, check := range list {
		_, err := tx.Exec(`UPDATE ai_health_checks SET next_run_at = $1 WHERE id = $2`, held, check.Id.String())
		if err != nil {
			return nil, err
		}
	}

	return list, tx.Commit()
}

// Record stores the result of a probe and schedules the next one. TODO: remove if not used in future.
func (r *HealthCheckRepositoryPostgres) Record(id Id, status, output string, nextRunAt time.Time) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE ai_health_checks
		SET status = $1, output = $2, last_run_at = $3, next_run_at = $4, updated_at = $3
		WHERE id = $5`,
		status,
		raw,
		now,
		nextRunAt,
		id.String(),
	)

	return err
}

// NewHealthCheckMetaRepository returns the repository for health check metadata.
func NewHealthCheckMetaRepository(db *sql.DB) HealthCheckMetaRepository {
	return &HealthCheckMetaRepositoryPostgres{db: db}
}

// scanHealthCheckMeta maps a row into health check metadata. Column order must match the select list.
func scanHealthCheckMeta(meta *HealthCheckMeta) []any {
	return []any{
		&meta.Id,
		&meta.HealthCheckId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts a health check metadata row. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO ai_health_checks_meta (id, health_check_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns health check metadata by key. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) Get(id Id, key string) (*HealthCheckMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, health_check_id, key, value #>> '{}', created_at, updated_at
		FROM ai_health_checks_meta
		WHERE health_check_id = $1 AND key = $2`,
		scanHealthCheckMeta,
		id.String(),
		key,
	)
}

// Update updates an existing health check metadata row. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE ai_health_checks_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE health_check_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes a health check metadata row. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_health_checks_meta WHERE health_check_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByHealthCheckId lists health check metadata rows. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) ListByHealthCheckId(id Id) ([]*HealthCheckMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, health_check_id, key, value #>> '{}', created_at, updated_at
		FROM ai_health_checks_meta
		WHERE health_check_id = $1
		ORDER BY key`,
		scanHealthCheckMeta,
		id.String(),
	)
}

// Upsert creates or updates health check metadata. TODO: remove if not used in future.
func (r *HealthCheckMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
