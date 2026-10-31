// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrTrafficNotFound   = errors.New("traffic not found")
	ErrFailedListTraffic = errors.New("failed list traffic")
	ErrFailedGetTraffic  = errors.New("failed get traffic")
)

// Traffic is the module for recorded gateway calls.
type Traffic struct {
	TrafficRepository   db.TrafficRepository
	WorkspaceRepository db.WorkspaceRepository
}

// TrafficCallResponse is one gateway call shaped for API responses.
type TrafficCallResponse struct {
	Id                  db.Id  `json:"id"`
	WorkspaceId         db.Id  `json:"workspaceId"`
	SourceAgent         string `json:"sourceAgent"`
	SourceInstance      string `json:"sourceInstance,omitempty"`
	DestinationAgent    string `json:"destinationAgent"`
	DestinationInstance string `json:"destinationInstance"`
	Skill               string `json:"skill"`
	Verb                string `json:"verb"`
	Decision            string `json:"decision"`
	Result              string `json:"result"`
	LatencyMs           int64  `json:"latencyMs"`
	TaskId              string `json:"taskId,omitempty"`
	CreatedAt           string `json:"createdAt"`
}

// ListTrafficResponse is returned when listing gateway calls.
type ListTrafficResponse struct {
	Calls []*TrafficCallResponse
	Total int64
}

// NewTraffic creates a traffic module with the given repositories.
func NewTraffic(calls db.TrafficRepository, workspaces db.WorkspaceRepository) *Traffic {
	return &Traffic{
		TrafficRepository:   calls,
		WorkspaceRepository: workspaces,
	}
}

// toTrafficCallResponse shapes a traffic row for API responses.
func toTrafficCallResponse(call *db.TrafficCall) *TrafficCallResponse {
	return &TrafficCallResponse{
		Id:                  call.Id,
		WorkspaceId:         call.WorkspaceId,
		SourceAgent:         call.SourceAgent,
		SourceInstance:      call.SourceInstance,
		DestinationAgent:    call.DestinationAgent,
		DestinationInstance: call.DestinationInstance,
		Skill:               call.Skill,
		Verb:                call.Verb,
		Decision:            call.Decision,
		Result:              call.Result,
		LatencyMs:           int64(call.LatencyMs),
		TaskId:              lo.FromPtr(call.TaskId),
		CreatedAt:           call.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// ListTraffic returns recorded gateway calls for a workspace.
func (t *Traffic) ListTraffic(workspaceId db.Id, limit, offset int) (*ListTrafficResponse, error) {
	total, err := t.TrafficRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count traffic")
		return nil, fmt.Errorf("%w: %v", ErrFailedListTraffic, err)
	}

	calls, err := t.TrafficRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list traffic")
		return nil, fmt.Errorf("%w: %v", ErrFailedListTraffic, err)
	}

	list := make([]*TrafficCallResponse, 0, len(calls))
	for _, call := range calls {
		list = append(list, toTrafficCallResponse(call))
	}

	return &ListTrafficResponse{Calls: list, Total: total}, nil
}

// GetTraffic returns one gateway call by id.
func (t *Traffic) GetTraffic(workspaceId, trafficId db.Id) (*TrafficCallResponse, error) {
	call, err := t.TrafficRepository.GetById(trafficId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("trafficId", trafficId.String()).
			Msg("Failed to get traffic")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetTraffic, err)
	}
	if call == nil || call.WorkspaceId != workspaceId {
		return nil, ErrTrafficNotFound
	}

	return toTrafficCallResponse(call), nil
}
