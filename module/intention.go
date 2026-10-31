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
	ErrIntentionNotFound      = errors.New("intention not found")
	ErrInvalidIntentionAction = errors.New("invalid intention action")
	ErrFailedListIntentions   = errors.New("failed list intentions")
	ErrFailedGetIntention     = errors.New("failed get intention")
	ErrFailedCreateIntention  = errors.New("failed create intention")
	ErrFailedUpdateIntention  = errors.New("failed update intention")
	ErrFailedDeleteIntention  = errors.New("failed delete intention")
	ErrFailedCheckIntention   = errors.New("failed check intention")
)

// Intention is the module for workspace call policy.
type Intention struct {
	IntentionRepository db.IntentionRepository
	WorkspaceRepository db.WorkspaceRepository
}

// IntentionSource names the calling agent.
type IntentionSource struct {
	Agent string `json:"agent" validate:"required,max=100" label:"Source agent"`
}

// IntentionDestination names the called agent and an optional skill.
type IntentionDestination struct {
	Agent string `json:"agent" validate:"required,max=100" label:"Destination agent"`
	Skill string `json:"skill" validate:"omitempty,max=120" label:"Destination skill"`
}

// SaveIntentionRequest is the body for creating or updating an intention.
type SaveIntentionRequest struct {
	Source      IntentionSource      `json:"source" label:"Source"`
	Destination IntentionDestination `json:"destination" label:"Destination"`
	Action      string               `json:"action" validate:"required,oneof=allow deny" label:"Action"`
}

// CheckIntentionRequest is the body for an intention check.
type CheckIntentionRequest struct {
	Source      IntentionSource      `json:"source" label:"Source"`
	Destination IntentionDestination `json:"destination" label:"Destination"`
}

// IntentionResponse is an intention shaped for API responses.
type IntentionResponse struct {
	Id          db.Id  `json:"id"`
	WorkspaceId db.Id  `json:"workspaceId"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Skill       string `json:"skill,omitempty"`
	Action      string `json:"action"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ListIntentionsResponse is returned when listing intentions.
type ListIntentionsResponse struct {
	Intentions []*IntentionResponse
	Total      int64
}

// CheckIntentionResponse is the decision for one call.
type CheckIntentionResponse struct {
	Action    string             `json:"action"`
	Intention *IntentionResponse `json:"intention,omitempty"`
}

// NewIntention creates an intention module with the given repositories.
func NewIntention(intentions db.IntentionRepository, workspaces db.WorkspaceRepository) *Intention {
	return &Intention{
		IntentionRepository: intentions,
		WorkspaceRepository: workspaces,
	}
}

// toIntentionResponse shapes an intention row for API responses.
func toIntentionResponse(item *db.Intention) *IntentionResponse {
	return &IntentionResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Source:      item.SourceAgent,
		Destination: item.DestinationAgent,
		Skill:       lo.FromPtr(item.Skill),
		Action:      item.Action,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListIntentions returns paginated intentions in a workspace.
func (i *Intention) ListIntentions(workspaceId db.Id, limit, offset int) (*ListIntentionsResponse, error) {
	total, err := i.IntentionRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count intentions")
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntentions, err)
	}

	intentions, err := i.IntentionRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list intentions")
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntentions, err)
	}

	list := make([]*IntentionResponse, 0, len(intentions))
	for _, item := range intentions {
		list = append(list, toIntentionResponse(item))
	}

	return &ListIntentionsResponse{Intentions: list, Total: total}, nil
}

// GetIntention returns one intention by id.
func (i *Intention) GetIntention(workspaceId, intentionId db.Id) (*IntentionResponse, error) {
	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to get intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return nil, ErrIntentionNotFound
	}

	return toIntentionResponse(intention), nil
}

// CreateIntention inserts an allow or deny rule.
func (i *Intention) CreateIntention(workspaceId db.Id, req *SaveIntentionRequest) (*IntentionResponse, error) {
	if !lo.Contains([]string{db.IntentionActionAllow, db.IntentionActionDeny}, req.Action) {
		return nil, ErrInvalidIntentionAction
	}

	var skill *string
	if !lo.IsEmpty(req.Destination.Skill) {
		value := req.Destination.Skill
		skill = &value
	}

	intention := &db.Intention{
		WorkspaceId:      workspaceId,
		SourceAgent:      req.Source.Agent,
		DestinationAgent: req.Destination.Agent,
		Skill:            skill,
		Action:           req.Action,
	}
	err := i.IntentionRepository.Create(intention)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateIntention, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("intentionId", intention.Id.String()).
		Str("source", intention.SourceAgent).
		Str("destination", intention.DestinationAgent).
		Str("action", intention.Action).
		Msg("Intention created")

	return toIntentionResponse(intention), nil
}

// UpdateIntention replaces an existing rule.
func (i *Intention) UpdateIntention(workspaceId, intentionId db.Id, req *SaveIntentionRequest) (*IntentionResponse, error) {
	if !lo.Contains([]string{db.IntentionActionAllow, db.IntentionActionDeny}, req.Action) {
		return nil, ErrInvalidIntentionAction
	}

	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to get intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return nil, ErrIntentionNotFound
	}

	var skill *string
	if !lo.IsEmpty(req.Destination.Skill) {
		value := req.Destination.Skill
		skill = &value
	}

	intention.SourceAgent = req.Source.Agent
	intention.DestinationAgent = req.Destination.Agent
	intention.Skill = skill
	intention.Action = req.Action
	err = i.IntentionRepository.Update(intention)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to update intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}

	intention, err = i.IntentionRepository.GetById(intention.Id)
	if err != nil || intention == nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to reload intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntention, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("intentionId", intention.Id.String()).
		Str("source", intention.SourceAgent).
		Str("destination", intention.DestinationAgent).
		Str("action", intention.Action).
		Msg("Intention updated")

	return toIntentionResponse(intention), nil
}

// DeleteIntention removes a rule.
func (i *Intention) DeleteIntention(workspaceId, intentionId db.Id) error {
	intention, err := i.IntentionRepository.GetById(intentionId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to get intention")
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntention, err)
	}
	if intention == nil || intention.WorkspaceId != workspaceId {
		return ErrIntentionNotFound
	}

	err = i.IntentionRepository.Delete(intention.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("intentionId", intentionId.String()).
			Msg("Failed to delete intention")
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntention, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("intentionId", intentionId.String()).
		Msg("Intention deleted")

	return nil
}

// CheckIntention returns the matching rule, or deny when none exists.
func (i *Intention) CheckIntention(workspaceId db.Id, req *CheckIntentionRequest) (*CheckIntentionResponse, error) {
	intention, err := i.IntentionRepository.Match(
		workspaceId,
		req.Source.Agent,
		req.Destination.Agent,
		req.Destination.Skill,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("source", req.Source.Agent).
			Str("destination", req.Destination.Agent).
			Str("skill", req.Destination.Skill).
			Msg("Failed to match intention")
		return nil, fmt.Errorf("%w: %v", ErrFailedCheckIntention, err)
	}
	if intention == nil {
		log.Debug().
			Str("workspaceId", workspaceId.String()).
			Str("source", req.Source.Agent).
			Str("destination", req.Destination.Agent).
			Str("skill", req.Destination.Skill).
			Msg("No intention matched, call denied")
		return &CheckIntentionResponse{Action: db.IntentionActionDeny}, nil
	}

	log.Debug().
		Str("workspaceId", workspaceId.String()).
		Str("intentionId", intention.Id.String()).
		Str("source", req.Source.Agent).
		Str("destination", req.Destination.Agent).
		Str("skill", req.Destination.Skill).
		Str("action", intention.Action).
		Msg("Intention matched")

	return &CheckIntentionResponse{
		Action:    intention.Action,
		Intention: toIntentionResponse(intention),
	}, nil
}
