// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/lb"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrInvalidAgentName     = errors.New("invalid agent name")
	ErrInvalidCardPath      = errors.New("invalid agent card path")
	ErrInvalidLbStrategy    = errors.New("invalid load balancing strategy")
	ErrAgentVersionRequired = errors.New("agent version required")
	ErrFailedListAgents     = errors.New("failed list agents")
	ErrFailedGetAgent       = errors.New("failed get agent")
	ErrFailedUpsertAgent    = errors.New("failed upsert agent")
	ErrFailedDeleteAgent    = errors.New("failed delete agent")
)

// AgentNamePattern is the regex pattern for agent names.
var AgentNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Agent is the module for the workspace agent catalog.
type Agent struct {
	AgentRepository       db.AgentRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
}

// UpsertAgentRequest is the body for registering or updating an agent.
type UpsertAgentRequest struct {
	Version    string `json:"version" validate:"required,max=40" label:"Version"`
	CardPath   string `json:"cardPath" validate:"omitempty,max=255,startswith=/" label:"Card Path"`
	LbStrategy string `json:"lbStrategy" validate:"omitempty,oneof=random round_robin least_outstanding_tasks sticky" label:"Load Balancing Strategy"`
}

// AgentResponse is an agent shaped for API responses.
type AgentResponse struct {
	Id           db.Id           `json:"id"`
	WorkspaceId  db.Id           `json:"workspaceId"`
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Card         json.RawMessage `json:"card"`
	CardChecksum string          `json:"cardChecksum"`
	CardPath     string          `json:"cardPath"`
	LbStrategy   string          `json:"lbStrategy"`
	Health       string          `json:"health"`
	StatusCounts map[string]int  `json:"statusCounts"`
	CreatedAt    string          `json:"createdAt"`
	UpdatedAt    string          `json:"updatedAt"`
}

// ListAgentsResponse is returned when listing agents.
type ListAgentsResponse struct {
	Agents []*AgentResponse
	Total  int64
}

// NewAgent creates an agent module with the given repositories.
func NewAgent(
	agents db.AgentRepository,
	instances db.AgentInstanceRepository,
	checks db.HealthCheckRepository,
	workspaces db.WorkspaceRepository,
) *Agent {
	return &Agent{
		AgentRepository:       agents,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		WorkspaceRepository:   workspaces,
	}
}

// toAgentResponse shapes an agent row for API responses.
func toAgentResponse(item *db.Agent, health string, counts map[string]int) *AgentResponse {
	return &AgentResponse{
		Id:           item.Id,
		WorkspaceId:  item.WorkspaceId,
		Name:         item.Name,
		Version:      item.Version,
		Card:         util.JSONRawFromString(item.Card),
		CardChecksum: lo.FromPtr(item.CardChecksum),
		CardPath:     item.CardPath,
		LbStrategy:   lo.CoalesceOrEmpty(item.LbStrategy, string(lb.RoundRobin)),
		Health:       health,
		StatusCounts: counts,
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// isValidLbStrategy reports whether s names a strategy the picker knows.
func isValidLbStrategy(s string) bool {
	switch lb.Strategy(s) {
	case lb.Random, lb.RoundRobin, lb.LeastOutstandingTasks, lb.Sticky:
		return true
	}

	return false
}

// ListAgents returns paginated agents in a workspace.
func (a *Agent) ListAgents(workspaceId db.Id, limit, offset int) (*ListAgentsResponse, error) {
	total, err := a.AgentRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count agents")
		return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
	}

	agents, err := a.AgentRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list agents")
		return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
	}

	list := make([]*AgentResponse, 0, len(agents))
	for _, item := range agents {
		health, counts, err := a.GetHealthStatus(item.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("agentId", item.Id.String()).
				Msg("Failed to get agent health")
			return nil, fmt.Errorf("%w: %v", ErrFailedListAgents, err)
		}

		list = append(list, toAgentResponse(item, health, counts))
	}

	return &ListAgentsResponse{Agents: list, Total: total}, nil
}

// GetAgent returns one agent by name.
func (a *Agent) GetAgent(workspaceId db.Id, name string) (*AgentResponse, error) {
	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", name).
			Msg("Failed to get agent")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	health, counts, err := a.GetHealthStatus(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentId", agent.Id.String()).
			Msg("Failed to get agent health")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAgent, err)
	}

	return toAgentResponse(agent, health, counts), nil
}

// UpsertAgent creates or updates an agent's version, card path and load balancing strategy.
func (a *Agent) UpsertAgent(workspaceId db.Id, name string, req *UpsertAgentRequest) (*AgentResponse, bool, error) {
	if !AgentNamePattern.MatchString(name) {
		return nil, false, ErrInvalidAgentName
	}

	if req == nil || req.Version == "" || len(req.Version) > 40 {
		return nil, false, ErrAgentVersionRequired
	}

	cardPath := lo.CoalesceOrEmpty(req.CardPath, DefaultCardPath)
	if !strings.HasPrefix(cardPath, "/") || len(cardPath) > 255 {
		return nil, false, ErrInvalidCardPath
	}

	if req.LbStrategy != "" && !isValidLbStrategy(req.LbStrategy) {
		return nil, false, ErrInvalidLbStrategy
	}

	existing, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", name).
			Msg("Failed to get agent")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	created := existing == nil
	if created {
		existing = &db.Agent{WorkspaceId: workspaceId, Name: name}
	}

	// A new version or card path makes the stored card stale, so the next instance fetches it again.
	if existing.Version != req.Version || existing.CardPath != cardPath {
		existing.Card = nil
		existing.CardChecksum = nil
	}

	existing.Name = name
	existing.CardPath = cardPath
	existing.Version = req.Version
	existing.LbStrategy = lo.CoalesceOrEmpty(
		req.LbStrategy,
		existing.LbStrategy,
		string(lb.RoundRobin),
	)

	if created {
		err = a.AgentRepository.Create(existing)
	} else {
		err = a.AgentRepository.Update(existing)
		if err == nil {
			existing, err = a.AgentRepository.GetById(existing.Id)
		}
	}
	if err != nil || existing == nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", name).
			Msg("Failed to upsert agent")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	health, counts, err := a.GetHealthStatus(existing.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentId", existing.Id.String()).
			Msg("Failed to get agent health")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedUpsertAgent, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentId", existing.Id.String()).
		Str("agentName", name).
		Str("version", existing.Version).
		Bool("created", created).
		Msg("Agent upserted")

	return toAgentResponse(existing, health, counts), created, nil
}

// DeleteAgent removes an agent by name.
func (a *Agent) DeleteAgent(workspaceId db.Id, name string) error {
	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, name)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", name).
			Msg("Failed to get agent")
		return fmt.Errorf("%w: %v", ErrFailedDeleteAgent, err)
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	err = a.AgentRepository.Delete(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentId", agent.Id.String()).
			Msg("Failed to delete agent")
		return fmt.Errorf("%w: %v", ErrFailedDeleteAgent, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentId", agent.Id.String()).
		Str("agentName", name).
		Msg("Agent deleted")

	return nil
}

// GetHealthStatus returns the agent health and how many of its instances are in each status.
func (a *Agent) GetHealthStatus(agentId db.Id) (string, map[string]int, error) {
	instances, err := a.InstanceRepository.ListByAgentId(agentId)
	if err != nil {
		log.Error().
			Err(err).
			Str("agentId", agentId.String()).
			Msg("Failed to list agent instances")
		return "", nil, err
	}

	now := time.Now().UTC()
	counts := map[string]int{
		db.AgentInstanceStatusPassing:  0,
		db.AgentInstanceStatusWarning:  0,
		db.AgentInstanceStatusUnknown:  0,
		db.AgentInstanceStatusCritical: 0,
	}

	for _, instance := range instances {
		checks, err := a.HealthCheckRepository.ListByInstanceId(instance.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("agentId", agentId.String()).
				Str("instanceId", instance.Id.String()).
				Msg("Failed to list instance health checks")
			return "", nil, err
		}

		// The instance status is the worst check status. A TTL check past its TTL
		status := db.AgentInstanceStatusPassing
		for _, check := range checks {
			current := lo.Ternary(
				check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)),
				db.HealthCheckStatusCritical,
				check.Status,
			)

			switch current {
			case db.HealthCheckStatusPassing:
			case db.HealthCheckStatusWarning:
				if status == db.AgentInstanceStatusPassing {
					status = db.AgentInstanceStatusWarning
				}
			case db.HealthCheckStatusUnknown:
				if status != db.AgentInstanceStatusCritical {
					status = db.AgentInstanceStatusUnknown
				}
			default:
				status = db.AgentInstanceStatusCritical
			}
		}

		counts[status]++
	}

	// The agent is as healthy as its worst instance. No instances means unknown.
	health := db.AgentInstanceStatusUnknown
	for _, status := range []string{
		db.AgentInstanceStatusCritical,
		db.AgentInstanceStatusUnknown,
		db.AgentInstanceStatusWarning,
		db.AgentInstanceStatusPassing,
	} {
		if counts[status] > 0 {
			health = status
			break
		}
	}

	return health, counts, nil
}
