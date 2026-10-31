// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	EventAuditLog              = "audit.log"
	EventAgentHealthChanged    = "agent.health.changed"
	EventAgentInstanceChanged  = "agent.instance.changed"
	EventCatalogChanged        = "catalog.changed"
	EventGatewayCallDenied     = "gateway.call.denied"
	EventSystemHealthDegraded  = "system.health.degraded"
	IntegrationDeliverTaskType = "integration.deliver"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrFailedListEvents = errors.New("failed list events")
	ErrFailedGetEvent   = errors.New("failed get event")
)

// Event is the module for reading workspace events.
type Event struct {
	EventRepository db.EventRepository
}

// EventRequest is who did what to which resource, passed when triggering an event.
type EventRequest struct {
	ActorId      db.Id
	ActorName    string
	ActorType    string
	ResourceType string
	ResourceId   db.Id
	IPAddress    string
	UserAgent    string
	Data         map[string]string
}

// EventTrigger stores and delivers one workspace event. TriggerEvent is the default.
type EventTrigger func(workspaceId db.Id, event string, request EventRequest)

// EventResponse is an event shaped for API responses.
type EventResponse struct {
	Id           db.Id           `json:"id"`
	WorkspaceId  db.Id           `json:"workspaceId"`
	Type         string          `json:"type"`
	ActorId      *db.Id          `json:"actorId,omitempty"`
	ActorName    *string         `json:"actorName,omitempty"`
	ActorType    *string         `json:"actorType,omitempty"`
	ResourceType *string         `json:"resourceType,omitempty"`
	ResourceId   *db.Id          `json:"resourceId,omitempty"`
	IPAddress    *string         `json:"ipAddress,omitempty"`
	UserAgent    *string         `json:"userAgent,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	CreatedAt    string          `json:"createdAt"`
}

// ListEventsResponse is returned when listing events.
type ListEventsResponse struct {
	Events []*EventResponse
	Total  int64
}

// NewEvent creates an event module with the given repository.
func NewEvent(events db.EventRepository) *Event {
	return &Event{
		EventRepository: events,
	}
}

// toEventResponse shapes an event row for API responses.
func toEventResponse(item *db.Event) *EventResponse {
	return &EventResponse{
		Id:           item.Id,
		WorkspaceId:  item.WorkspaceId,
		Type:         item.Type,
		ActorId:      item.ActorId,
		ActorName:    item.ActorName,
		ActorType:    item.ActorType,
		ResourceType: item.ResourceType,
		ResourceId:   item.ResourceId,
		IPAddress:    item.IPAddress,
		UserAgent:    item.UserAgent,
		Data:         util.JSONRawFromString(item.Data),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ListEvents returns paginated events in a workspace, newest first.
func (e *Event) ListEvents(workspaceId db.Id, limit, offset int) (*ListEventsResponse, error) {
	total, err := e.EventRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count events")
		return nil, fmt.Errorf("%w: %v", ErrFailedListEvents, err)
	}

	events, err := e.EventRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list events")
		return nil, fmt.Errorf("%w: %v", ErrFailedListEvents, err)
	}

	list := make([]*EventResponse, 0, len(events))
	for _, item := range events {
		list = append(list, toEventResponse(item))
	}

	return &ListEventsResponse{Events: list, Total: total}, nil
}

// GetEvent returns one event by id.
func (e *Event) GetEvent(workspaceId, eventId db.Id) (*EventResponse, error) {
	event, err := e.EventRepository.GetById(eventId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("eventId", eventId.String()).
			Msg("Failed to get event")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetEvent, err)
	}
	if event == nil || event.WorkspaceId != workspaceId {
		return nil, ErrEventNotFound
	}

	return toEventResponse(event), nil
}

// TriggerEvent stores the event and enqueues a delivery per subscribed integration.
func TriggerEvent(workspaceId db.Id, event string, request EventRequest) {
	events := db.NewEventRepository(db.GetDB(true))
	integrations := db.NewIntegrationRepository(db.GetDB())

	raw, _ := json.Marshal(request.Data)

	stored := &db.Event{
		WorkspaceId:  workspaceId,
		Type:         event,
		ActorId:      lo.EmptyableToPtr(request.ActorId),
		ActorName:    lo.EmptyableToPtr(request.ActorName),
		ActorType:    lo.EmptyableToPtr(request.ActorType),
		ResourceType: lo.EmptyableToPtr(request.ResourceType),
		ResourceId:   lo.EmptyableToPtr(request.ResourceId),
		IPAddress:    lo.EmptyableToPtr(request.IPAddress),
		UserAgent:    lo.EmptyableToPtr(request.UserAgent),
		Data:         new(string(raw)),
	}

	err := events.Create(stored)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("event", event).
			Msg("Failed to store event")
		return
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("eventId", stored.Id.String()).
		Str("event", event).
		Msg("Event stored")

	subscribed, err := integrations.ListActiveByEvent(workspaceId, event)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("event", event).
			Msg("Failed to list subscribed integrations")
		return
	}

	for _, integration := range subscribed {
		err = EnqueueTask(IntegrationDeliverTaskType, map[string]string{
			"integrationId": integration.Id.String(),
			"eventId":       stored.Id.String(),
			"event":         event,
			"actorId":       request.ActorId.String(),
			"actorName":     request.ActorName,
			"actorType":     request.ActorType,
			"resourceType":  request.ResourceType,
			"resourceId":    request.ResourceId.String(),
			"ipAddress":     request.IPAddress,
			"userAgent":     request.UserAgent,
			"data":          string(raw),
			"occurredAt":    stored.CreatedAt.UTC().Format(time.RFC3339),
		}, workspaceId)
		if err != nil {
			return
		}
	}
}
