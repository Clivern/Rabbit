// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AgentInstanceStatusPassing  = "passing"
	AgentInstanceStatusWarning  = "warning"
	AgentInstanceStatusCritical = "critical"
	AgentInstanceStatusUnknown  = "unknown"
)

// AgentInstance is a single row in the agent_instances table.
type AgentInstance struct {
	Id         Id
	AgentId    Id
	InstanceId string
	Address    string
	Port       int
	Scheme     string
	Path       string
	Meta       *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentInstanceRepository is the interface for agent instance persistence.
type AgentInstanceRepository interface {
	Create(instance *AgentInstance) error
	GetById(id Id) (*AgentInstance, error)
	GetByAgentAndInstanceId(agentId Id, instanceId string) (*AgentInstance, error)
	Update(instance *AgentInstance) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentInstance, error)
	ListLiveByAgentId(agentId Id, now time.Time) ([]*AgentInstance, error)
	CountByAgentId(agentId Id) (int64, error)
	ListPageByAgentId(agentId Id, limit, offset int) ([]*AgentInstance, error)
	CountLiveByAgentId(agentId Id, now time.Time) (int64, error)
	ListLivePageByAgentId(agentId Id, now time.Time, limit, offset int) ([]*AgentInstance, error)
}

type AgentInstanceRepositoryPostgres struct {
	db *sql.DB
}

// AgentInstanceMeta is a single row in the agent_instances_meta table.
type AgentInstanceMeta struct {
	Id              Id
	AgentInstanceId Id
	Key             string
	Value           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AgentInstanceMetaRepository is the interface for agent instance metadata CRUD.
type AgentInstanceMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*AgentInstanceMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByAgentInstanceId(id Id) ([]*AgentInstanceMeta, error)
	Upsert(id Id, key, value string) error
}

type AgentInstanceMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentInstanceRepository returns the repository for agent instances.
func NewAgentInstanceRepository(db *sql.DB) AgentInstanceRepository {
	return &AgentInstanceRepositoryPostgres{db: db}
}

// scanAgentInstance maps a row into an agent instance. Column order must match the select list.
func scanAgentInstance(instance *AgentInstance) []any {
	return []any{
		&instance.Id,
		&instance.AgentId,
		&instance.InstanceId,
		&instance.Address,
		&instance.Port,
		&instance.Scheme,
		&instance.Path,
		&instance.Meta,
		&instance.Status,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	}
}

// Create inserts an agent instance row.
func (r *AgentInstanceRepositoryPostgres) Create(instance *AgentInstance) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	instance.Id = id
	if instance.Status == "" {
		instance.Status = AgentInstanceStatusPassing
	}

	return r.db.QueryRow(
		`INSERT INTO ai_agent_instances
		(id, agent_id, instance_id, address, port, scheme, path, meta, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
		RETURNING created_at, updated_at`,
		instance.Id.String(),
		instance.AgentId.String(),
		instance.InstanceId,
		instance.Address,
		instance.Port,
		instance.Scheme,
		instance.Path,
		instance.Meta,
		instance.Status,
	).Scan(&instance.CreatedAt, &instance.UpdatedAt)
}

// GetById returns an agent instance by id.
func (r *AgentInstanceRepositoryPostgres) GetById(id Id) (*AgentInstance, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE id = $1`,
		scanAgentInstance,
		id.String(),
	)
}

// GetByAgentAndInstanceId returns an agent instance by agent and instance id.
func (r *AgentInstanceRepositoryPostgres) GetByAgentAndInstanceId(agentId Id, instanceId string) (*AgentInstance, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE agent_id = $1 AND instance_id = $2`,
		scanAgentInstance,
		agentId.String(),
		instanceId,
	)
}

// Update updates an agent instance address, endpoint and status.
func (r *AgentInstanceRepositoryPostgres) Update(instance *AgentInstance) error {
	_, err := r.db.Exec(
		`UPDATE ai_agent_instances
		SET address = $1, port = $2, scheme = $3, path = $4, meta = $5::jsonb, status = $6, updated_at = $7
		WHERE id = $8`,
		instance.Address,
		instance.Port,
		instance.Scheme,
		instance.Path,
		instance.Meta,
		instance.Status,
		time.Now().UTC(),
		instance.Id.String(),
	)

	return err
}

// Delete removes an agent instance row.
func (r *AgentInstanceRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM ai_agent_instances WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists instances for an agent.
func (r *AgentInstanceRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentInstance, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE agent_id = $1
		ORDER BY instance_id`,
		scanAgentInstance,
		agentId.String(),
	)
}

// ListLiveByAgentId lists instances with no critical, unknown, or expired checks.
func (r *AgentInstanceRepositoryPostgres) ListLiveByAgentId(agentId Id, now time.Time) ([]*AgentInstance, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE agent_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM ai_health_checks hc
				WHERE hc.instance_id = ai_agent_instances.id
					AND (
						hc.status IN ($3, $5)
						OR (hc.type = $4 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $2))
					)
			)
		ORDER BY instance_id`,
		scanAgentInstance,
		agentId.String(),
		now,
		HealthCheckStatusCritical,
		HealthCheckTypeTTL,
		HealthCheckStatusUnknown,
	)
}

// CountByAgentId counts the instances of an agent.
func (r *AgentInstanceRepositoryPostgres) CountByAgentId(agentId Id) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ai_agent_instances WHERE agent_id = $1`,
		agentId.String(),
	).Scan(&total)

	return total, err
}

// ListPageByAgentId lists one page of instances for an agent.
func (r *AgentInstanceRepositoryPostgres) ListPageByAgentId(agentId Id, limit, offset int) ([]*AgentInstance, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE agent_id = $1
		ORDER BY instance_id
		LIMIT $2 OFFSET $3`,
		scanAgentInstance,
		agentId.String(),
		limit,
		offset,
	)
}

// CountLiveByAgentId counts the instances with no critical or expired checks.
func (r *AgentInstanceRepositoryPostgres) CountLiveByAgentId(agentId Id, now time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM ai_agent_instances
		WHERE agent_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM ai_health_checks hc
				WHERE hc.instance_id = ai_agent_instances.id
					AND (
						hc.status IN ($3, $5)
						OR (hc.type = $4 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $2))
					)
			)`,
		agentId.String(),
		now,
		HealthCheckStatusCritical,
		HealthCheckTypeTTL,
		HealthCheckStatusUnknown,
	).Scan(&total)

	return total, err
}

// ListLivePageByAgentId lists one page of instances with no critical, unknown, or expired checks.
func (r *AgentInstanceRepositoryPostgres) ListLivePageByAgentId(agentId Id, now time.Time, limit, offset int) ([]*AgentInstance, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_id, instance_id, address, port, scheme, path, meta, status, created_at, updated_at
		FROM ai_agent_instances
		WHERE agent_id = $1
			AND NOT EXISTS (
				SELECT 1 FROM ai_health_checks hc
				WHERE hc.instance_id = ai_agent_instances.id
					AND (
						hc.status IN ($3, $5)
						OR (hc.type = $4 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $2))
					)
			)
		ORDER BY instance_id
		LIMIT $6 OFFSET $7`,
		scanAgentInstance,
		agentId.String(),
		now,
		HealthCheckStatusCritical,
		HealthCheckTypeTTL,
		HealthCheckStatusUnknown,
		limit,
		offset,
	)
}

// NewAgentInstanceMetaRepository returns the repository for agent instance metadata.
func NewAgentInstanceMetaRepository(db *sql.DB) AgentInstanceMetaRepository {
	return &AgentInstanceMetaRepositoryPostgres{db: db}
}

// scanAgentInstanceMeta maps a row into agent instance metadata. Column order must match the select list.
func scanAgentInstanceMeta(meta *AgentInstanceMeta) []any {
	return []any{
		&meta.Id,
		&meta.AgentInstanceId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts an agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO ai_agent_instances_meta (id, agent_instance_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns agent instance metadata by key. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Get(id Id, key string) (*AgentInstanceMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, agent_instance_id, key, value #>> '{}', created_at, updated_at
		FROM ai_agent_instances_meta
		WHERE agent_instance_id = $1 AND key = $2`,
		scanAgentInstanceMeta,
		id.String(),
		key,
	)
}

// Update updates an existing agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE ai_agent_instances_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE agent_instance_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes an agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM ai_agent_instances_meta WHERE agent_instance_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByAgentInstanceId lists agent instance metadata rows. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) ListByAgentInstanceId(id Id) ([]*AgentInstanceMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, agent_instance_id, key, value #>> '{}', created_at, updated_at
		FROM ai_agent_instances_meta
		WHERE agent_instance_id = $1
		ORDER BY key`,
		scanAgentInstanceMeta,
		id.String(),
	)
}

// Upsert creates or updates agent instance metadata. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
