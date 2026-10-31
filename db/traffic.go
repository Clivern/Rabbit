// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	TrafficDecisionAllow = "allow"
	TrafficDecisionDeny  = "deny"

	TrafficResultCompleted  = "completed"
	TrafficResultFailed     = "failed"
	TrafficResultDenied     = "denied"
	TrafficResultTimeout    = "timeout"
	TrafficResultNoInstance = "no_instance"
)

// TrafficCall is a single row in the traffic table.
type TrafficCall struct {
	Id                  Id
	WorkspaceId         Id
	SourceAgent         string
	SourceInstance      string
	DestinationAgent    string
	DestinationInstance string
	Skill               string
	Verb                string
	Decision            string
	Result              string
	LatencyMs           int
	TaskId              *string
	Meta                *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TrafficRepository is the interface for gateway traffic persistence.
type TrafficRepository interface {
	Create(call *TrafficCall) error
	GetById(id Id) (*TrafficCall, error)
	GetAllowedByTaskId(workspaceId Id, taskId string) (*TrafficCall, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*TrafficCall, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type TrafficRepositoryPostgres struct {
	db *sql.DB
}

// NewTrafficRepository returns the repository for traffic.
func NewTrafficRepository(db *sql.DB) TrafficRepository {
	return &TrafficRepositoryPostgres{db: db}
}

// scanTrafficCall maps a row into a traffic call. Column order must match the select list.
func scanTrafficCall(call *TrafficCall) []any {
	return []any{
		&call.Id,
		&call.WorkspaceId,
		&call.SourceAgent,
		&call.SourceInstance,
		&call.DestinationAgent,
		&call.DestinationInstance,
		&call.Skill,
		&call.Verb,
		&call.Decision,
		&call.Result,
		&call.LatencyMs,
		&call.TaskId,
		&call.Meta,
		&call.CreatedAt,
		&call.UpdatedAt,
	}
}

// Create inserts a traffic call row.
func (r *TrafficRepositoryPostgres) Create(call *TrafficCall) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	call.Id = id

	return r.db.QueryRow(
		`INSERT INTO ai_traffic (
			id, workspace_id, source_agent, source_instance, destination_agent, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb)
		RETURNING created_at, updated_at`,
		call.Id.String(),
		call.WorkspaceId.String(),
		call.SourceAgent,
		call.SourceInstance,
		call.DestinationAgent,
		call.DestinationInstance,
		call.Skill,
		call.Verb,
		call.Decision,
		call.Result,
		call.LatencyMs,
		call.TaskId,
		call.Meta,
	).Scan(&call.CreatedAt, &call.UpdatedAt)
}

// GetById returns a traffic call by id.
func (r *TrafficRepositoryPostgres) GetById(id Id) (*TrafficCall, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, source_agent, source_instance, destination_agent, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta,
			created_at, updated_at
		FROM ai_traffic
		WHERE id = $1`,
		scanTrafficCall,
		id.String(),
	)
}

// GetAllowedByTaskId returns the latest allowed call that carried a task id in a workspace.
func (r *TrafficRepositoryPostgres) GetAllowedByTaskId(workspaceId Id, taskId string) (*TrafficCall, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, source_agent, source_instance, destination_agent, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta,
			created_at, updated_at
		FROM ai_traffic
		WHERE workspace_id = $1 AND task_id = $2 AND decision = $3
		ORDER BY created_at DESC
		LIMIT 1`,
		scanTrafficCall,
		workspaceId.String(),
		taskId,
		TrafficDecisionAllow,
	)
}

// ListByWorkspaceId lists traffic calls for a workspace, newest first.
func (r *TrafficRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*TrafficCall, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, source_agent, source_instance, destination_agent, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta,
			created_at, updated_at
		FROM ai_traffic
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanTrafficCall,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns the number of traffic calls in a workspace.
func (r *TrafficRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ai_traffic WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
