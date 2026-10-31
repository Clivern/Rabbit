// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	EventActorTypeUser      = "user"
	EventActorTypeAccessKey = "access_key"
	EventActorTypeSystem    = "system"

	EventResourceTypeAgentInstance = "agent_instance"
	EventResourceTypeHealthCheck   = "health_check"
)

// Event is a single row in the events table.
type Event struct {
	Id           Id
	WorkspaceId  Id
	Type         string
	ActorId      *Id
	ActorName    *string
	ActorType    *string
	ResourceType *string
	ResourceId   *Id
	IPAddress    *string
	UserAgent    *string
	Data         *string
	CreatedAt    time.Time
}

// EventRepository is the interface for event persistence.
type EventRepository interface {
	Create(event *Event) error
	GetById(id Id) (*Event, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Event, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type EventRepositoryPostgres struct {
	db *sql.DB
}

// NewEventRepository returns an event repository.
func NewEventRepository(db *sql.DB) EventRepository {
	return &EventRepositoryPostgres{db: db}
}

// scanEvent maps a row into an event. Column order must match the select list.
func scanEvent(event *Event) []any {
	return []any{
		&event.Id,
		&event.WorkspaceId,
		&event.Type,
		&event.ActorId,
		&event.ActorName,
		&event.ActorType,
		&event.ResourceType,
		&event.ResourceId,
		&event.IPAddress,
		&event.UserAgent,
		&event.Data,
		&event.CreatedAt,
	}
}

// Create inserts an event row.
func (r *EventRepositoryPostgres) Create(event *Event) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	event.Id = id

	err = r.db.QueryRow(
		`INSERT INTO events (
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`,
		event.Id.String(),
		event.WorkspaceId.String(),
		event.Type,
		event.ActorId,
		event.ActorName,
		event.ActorType,
		event.ResourceType,
		event.ResourceId,
		event.IPAddress,
		event.UserAgent,
		event.Data,
	).Scan(&event.CreatedAt)

	return err
}

// GetById returns an event by id.
func (r *EventRepositoryPostgres) GetById(id Id) (*Event, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data, created_at
		FROM events
		WHERE id = $1`,
		scanEvent,
		id.String(),
	)
}

// ListByWorkspaceId returns the newest events of a workspace.
func (r *EventRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Event, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data, created_at
		FROM events
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanEvent,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns the total number of events in a workspace.
func (r *EventRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM events
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
