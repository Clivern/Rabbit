// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AsyncTaskStatusPending   = "pending"
	AsyncTaskStatusRunning   = "running"
	AsyncTaskStatusCompleted = "completed"
	AsyncTaskStatusFailed    = "failed"

	AsyncTaskTypeNoop = "sloth.noop"
)

// AsyncTask is a single row in the async_tasks table.
type AsyncTask struct {
	Id          Id
	WorkspaceId Id
	Type        string
	Status      string
	Payload     *string
	Result      *string
	Error       *string
	Attempts    int
	Priority    int
	RunAt       *time.Time
	LockedAt    *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AsyncTaskRepository is the interface for async task persistence.
type AsyncTaskRepository interface {
	Create(task *AsyncTask) error
	MarkRunning(id Id) error
	Complete(id Id, result string) error
	Fail(id Id, message string) error
	CountByStatus(status string) (int64, error)
	ListByStatus(status string) ([]*AsyncTask, error)
}

type AsyncTaskRepositoryPostgres struct {
	db *sql.DB
}

// NewAsyncTaskRepository returns an async task repository.
func NewAsyncTaskRepository(db *sql.DB) AsyncTaskRepository {
	return &AsyncTaskRepositoryPostgres{db: db}
}

// Create inserts an async task row.
func (r *AsyncTaskRepositoryPostgres) Create(task *AsyncTask) error {
	if task.Id == "" {
		id, err := NewId()
		if err != nil {
			return err
		}

		task.Id = id
	}
	if task.Status == "" {
		task.Status = AsyncTaskStatusPending
	}

	err := r.db.QueryRow(
		`INSERT INTO async_tasks (
			id, workspace_id, type, status, payload, result, error, attempts, priority,
			run_at, locked_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`,
		task.Id.String(),
		task.WorkspaceId,
		task.Type,
		task.Status,
		task.Payload,
		task.Result,
		task.Error,
		task.Attempts,
		task.Priority,
		task.RunAt,
		task.LockedAt,
		task.CompletedAt,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	return err
}

// MarkRunning sets a task to running.
func (r *AsyncTaskRepositoryPostgres) MarkRunning(id Id) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE async_tasks
		SET status = $1, locked_at = $2, run_at = $2, attempts = attempts + 1, updated_at = $2
		WHERE id = $3`,
		AsyncTaskStatusRunning,
		now,
		id.String(),
	)

	return err
}

// Complete marks a task as completed.
func (r *AsyncTaskRepositoryPostgres) Complete(id Id, result string) error {
	now := time.Now().UTC()
	var raw *string
	if result != "" {
		raw = &result
	}

	_, err := r.db.Exec(
		`UPDATE async_tasks
		SET status = $1, result = $2::jsonb, completed_at = $3, locked_at = NULL, updated_at = $3
		WHERE id = $4`,
		AsyncTaskStatusCompleted,
		raw,
		now,
		id.String(),
	)

	return err
}

// Fail marks a task as failed. TODO: remove if not used in future.
func (r *AsyncTaskRepositoryPostgres) Fail(id Id, message string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE async_tasks
		SET status = $1, error = to_jsonb($2::text), completed_at = $3, locked_at = NULL, updated_at = $3
		WHERE id = $4`,
		AsyncTaskStatusFailed,
		message,
		now,
		id.String(),
	)

	return err
}

// CountByStatus counts async task rows by status.
func (r *AsyncTaskRepositoryPostgres) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM async_tasks
		WHERE status = $1`,
		status,
	).Scan(&count)

	return count, err
}

// ListByStatus lists async tasks with the given status.
func (r *AsyncTaskRepositoryPostgres) ListByStatus(status string) ([]*AsyncTask, error) {
	rows, err := r.db.Query(
		`SELECT id, type, payload::text
		FROM async_tasks
		WHERE status = $1
		ORDER BY priority DESC, created_at`,
		status,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AsyncTask
	for rows.Next() {
		task := &AsyncTask{}
		var payload string
		err := rows.Scan(&task.Id, &task.Type, &payload)
		if err != nil {
			return nil, err
		}

		task.Payload = &payload
		list = append(list, task)
	}

	return list, rows.Err()
}

// AsyncTaskMeta is a single row in the async_tasks_meta table.
type AsyncTaskMeta struct {
	Id          Id
	AsyncTaskId Id
	Key         string
	Value       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AsyncTaskMetaRepository is the interface for async task metadata CRUD.
type AsyncTaskMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*AsyncTaskMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByAsyncTaskId(id Id) ([]*AsyncTaskMeta, error)
	Upsert(id Id, key, value string) error
}

type AsyncTaskMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewAsyncTaskMetaRepository returns the repository for async task metadata.
func NewAsyncTaskMetaRepository(db *sql.DB) AsyncTaskMetaRepository {
	return &AsyncTaskMetaRepositoryPostgres{db: db}
}

// scanAsyncTaskMeta maps a row into async task metadata. Column order must match the select list.
func scanAsyncTaskMeta(meta *AsyncTaskMeta) []any {
	return []any{
		&meta.Id,
		&meta.AsyncTaskId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// Create inserts an async task metadata row. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO async_tasks_meta (id, async_task_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(), id.String(), key, value,
	)

	return err
}

// Get returns async task metadata by key. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) Get(id Id, key string) (*AsyncTaskMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, async_task_id, key, value #>> '{}', created_at, updated_at
		FROM async_tasks_meta
		WHERE async_task_id = $1 AND key = $2`,
		scanAsyncTaskMeta,
		id.String(), key,
	)
}

// Update updates an existing async task metadata row. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE async_tasks_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE async_task_id = $3 AND key = $4`,
		value, time.Now().UTC(), id.String(), key,
	)

	return err
}

// Delete deletes an async task metadata row. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM async_tasks_meta
		WHERE async_task_id = $1 AND key = $2`,
		id.String(), key,
	)

	return err
}

// ListByAsyncTaskId lists async task metadata rows by async task id. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) ListByAsyncTaskId(id Id) ([]*AsyncTaskMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, async_task_id, key, value #>> '{}', created_at, updated_at
		FROM async_tasks_meta
		WHERE async_task_id = $1
		ORDER BY key`,
		scanAsyncTaskMeta,
		id.String(),
	)
}

// Upsert creates or updates async task metadata. TODO: remove if not used in future.
func (r *AsyncTaskMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
