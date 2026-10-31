// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrCheckNotFound      = errors.New("check not found")
	ErrInvalidCheckId     = errors.New("invalid check id")
	ErrInvalidCheckPath   = errors.New("invalid check path")
	ErrInvalidCheckTiming = errors.New("invalid check timing")
	ErrCheckNotReportable = errors.New("check not reportable")
)

const (
	// DefaultCheckInterval is how often http and tcp checks run, in seconds.
	DefaultCheckInterval = 10
	// DefaultCheckTimeout is how long an http or tcp probe may take, in seconds.
	DefaultCheckTimeout = 2
	// MaxCheckOutput caps the output stored for a check.
	MaxCheckOutput = 4096
)

// CheckDefinition is the stored definition of a check template and its instance copies.
type CheckDefinition struct {
	Path     string `json:"path,omitempty"`
	Scheme   string `json:"scheme,omitempty"`
	Port     int    `json:"port,omitempty"`
	Interval int    `json:"interval,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

// UpsertAgentCheckRequest is the body for creating or replacing an agent check template.
type UpsertAgentCheckRequest struct {
	Name     string `json:"name" validate:"omitempty,max=120" label:"Name"`
	Type     string `json:"type" validate:"required,oneof=ttl http tcp" label:"Type"`
	Path     string `json:"path" validate:"omitempty,max=255" label:"Path"`
	Scheme   string `json:"scheme" validate:"omitempty,oneof=http https" label:"Scheme"`
	Port     int    `json:"port" validate:"omitempty,min=1,max=65535" label:"Port"`
	Interval int    `json:"interval" validate:"omitempty,min=5,max=3600" label:"Interval"`
	Timeout  int    `json:"timeout" validate:"omitempty,min=1,max=60" label:"Timeout"`
	TTL      int    `json:"ttl" validate:"omitempty,min=5,max=86400" label:"TTL"`
}

// ReportCheckRequest is the body an instance sends to update one of its TTL checks.
type ReportCheckRequest struct {
	Output string `json:"output" validate:"omitempty,max=4096" label:"Output"`
}

// AgentCheckResponse is an agent check template shaped for API responses.
type AgentCheckResponse struct {
	CheckId   string `json:"checkId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Path      string `json:"path,omitempty"`
	Scheme    string `json:"scheme,omitempty"`
	Port      int    `json:"port,omitempty"`
	Interval  int    `json:"interval,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListAgentChecksResponse is returned when listing the check templates of an agent.
type ListAgentChecksResponse struct {
	Checks []*AgentCheckResponse
	Total  int64
}

// Check is the module for agent check templates and instance check reports.
type Check struct {
	AgentRepository       db.AgentRepository
	AgentCheckRepository  db.AgentCheckRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	TriggerEvent          EventTrigger
}

// NewCheck creates a check module with the given repositories.
func NewCheck(agents db.AgentRepository, agentChecks db.AgentCheckRepository, instances db.AgentInstanceRepository, checks db.HealthCheckRepository) *Check {
	return &Check{
		AgentRepository:       agents,
		AgentCheckRepository:  agentChecks,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		TriggerEvent:          TriggerEvent,
	}
}

// toAgentCheckResponse shapes a check template for API responses.
func toAgentCheckResponse(template *db.AgentCheck, definition CheckDefinition) *AgentCheckResponse {
	return &AgentCheckResponse{
		CheckId:   template.CheckId,
		Name:      template.Name,
		Type:      template.Type,
		Path:      definition.Path,
		Scheme:    definition.Scheme,
		Port:      definition.Port,
		Interval:  definition.Interval,
		Timeout:   definition.Timeout,
		TTL:       definition.TTL,
		CreatedAt: template.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: template.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListAgentChecks returns one page of the check templates of an agent.
func (c *Check) ListAgentChecks(workspaceId db.Id, agentName string, limit, offset int) (*ListAgentChecksResponse, error) {
	agent, err := c.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	total, err := c.AgentCheckRepository.CountByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to count agent checks")
		return nil, err
	}

	templates, err := c.AgentCheckRepository.ListPageByAgentId(agent.Id, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to list agent checks")
		return nil, err
	}

	list := make([]*AgentCheckResponse, 0, len(templates))
	for _, template := range templates {
		var definition CheckDefinition
		if err := json.Unmarshal([]byte(template.Definition), &definition); err != nil {
			log.Warn().
				Err(err).
				Str("checkId", template.CheckId).
				Msg("Failed to decode check definition")
		}

		list = append(list, toAgentCheckResponse(template, definition))
	}

	return &ListAgentChecksResponse{
		Checks: list,
		Total:  total,
	}, nil
}

// UpsertAgentCheck creates or replaces a check template and applies it to every instance.
func (c *Check) UpsertAgentCheck(workspaceId db.Id, agentName, checkId string, req *UpsertAgentCheckRequest, lang string) (*AgentCheckResponse, bool, error) {
	if !InstanceIdPattern.MatchString(checkId) {
		return nil, false, ErrInvalidCheckId
	}

	var definition CheckDefinition
	switch req.Type {
	case db.HealthCheckTypeTTL:
		if req.TTL == 0 {
			return nil, false, ErrInvalidCheckTiming
		}
		definition.TTL = req.TTL

	case db.HealthCheckTypeHTTP, db.HealthCheckTypeTCP:
		definition.Port = req.Port
		definition.Interval = req.Interval
		definition.Timeout = req.Timeout
		if definition.Interval == 0 {
			definition.Interval = DefaultCheckInterval
		}
		if definition.Timeout == 0 {
			definition.Timeout = DefaultCheckTimeout
		}
		if definition.Timeout >= definition.Interval {
			return nil, false, ErrInvalidCheckTiming
		}

		if req.Type == db.HealthCheckTypeHTTP {
			if !strings.HasPrefix(req.Path, "/") || strings.ContainsAny(req.Path, " \t\r\n") {
				return nil, false, ErrInvalidCheckPath
			}

			definition.Path = req.Path
			definition.Scheme = lo.CoalesceOrEmpty(req.Scheme, "http")
		}
	}

	raw, err := json.Marshal(definition)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("checkId", checkId).
			Msg("Failed to encode check definition")
		return nil, false, err
	}

	agent, err := c.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, false, err
	}
	if agent == nil {
		return nil, false, ErrAgentNotFound
	}

	template, err := c.AgentCheckRepository.GetByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("checkId", checkId).
			Msg("Failed to get agent check")
		return nil, false, err
	}

	created := template == nil
	if created {
		template = &db.AgentCheck{AgentId: agent.Id, CheckId: checkId}
	}

	template.Name = strings.TrimSpace(req.Name)
	if template.Name == "" {
		template.Name = checkId
	}

	template.Type = req.Type
	template.Definition = string(raw)

	if created {
		err = c.AgentCheckRepository.Create(template)
	} else {
		err = c.AgentCheckRepository.Update(template)
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("checkId", checkId).
			Msg("Failed to save agent check")
		return nil, false, err
	}

	instances, err := c.InstanceRepository.ListByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to list agent instances")
		return nil, false, err
	}

	for _, instance := range instances {
		err = c.SyncInstanceCheck(instance.Id, template, lang)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("agentName", agentName).
				Str("instanceId", instance.Id.String()).
				Str("checkId", checkId).
				Msg("Failed to sync instance check")
			return nil, false, err
		}
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("checkId", checkId).
		Str("type", template.Type).
		Bool("created", created).
		Int("instances", len(instances)).
		Msg("Agent check upserted")

	return toAgentCheckResponse(template, definition), created, nil
}

// SyncInstanceCheck makes one instance's copy of a check template match the template.
func (c *Check) SyncInstanceCheck(instanceId db.Id, template *db.AgentCheck, lang string) error {
	now := time.Now().UTC()
	isPull := template.Type == db.HealthCheckTypeHTTP || template.Type == db.HealthCheckTypeTCP

	var definition CheckDefinition
	if err := json.Unmarshal([]byte(template.Definition), &definition); err != nil {
		log.Warn().
			Err(err).
			Str("checkId", template.CheckId).
			Msg("Failed to decode check definition")
	}

	existing, err := c.HealthCheckRepository.GetByInstanceAndCheckId(instanceId, template.CheckId)
	if err != nil {
		log.Error().
			Err(err).
			Str("instanceId", instanceId.String()).
			Str("checkId", template.CheckId).
			Msg("Failed to get instance check")
		return err
	}

	// A type change resets the check, since its old status no longer means anything.
	if existing != nil && existing.Type != template.Type {
		err = c.HealthCheckRepository.Delete(existing.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("instanceId", instanceId.String()).
				Str("checkId", template.CheckId).
				Msg("Failed to reset instance check")
			return err
		}

		log.Info().
			Str("instanceId", instanceId.String()).
			Str("checkId", template.CheckId).
			Str("fromType", existing.Type).
			Str("toType", template.Type).
			Msg("Instance check reset after type change")

		existing = nil
	}

	stored := template.Definition
	if existing != nil {
		existing.Name = template.Name
		existing.Definition = &stored
		existing.NextRunAt = nil
		if isPull {
			existing.NextRunAt = &now
		}

		err = c.HealthCheckRepository.Update(existing)
		if err != nil {
			log.Error().
				Err(err).
				Str("instanceId", instanceId.String()).
				Str("checkId", template.CheckId).
				Msg("Failed to update instance check")
			return err
		}

		log.Debug().
			Str("instanceId", instanceId.String()).
			Str("checkId", template.CheckId).
			Msg("Instance check updated")

		return nil
	}

	check := &db.HealthCheck{
		InstanceId: instanceId,
		CheckId:    template.CheckId,
		Name:       template.Name,
		Type:       template.Type,
		Source:     db.HealthCheckSourceAgent,
		Status:     db.HealthCheckStatusUnknown,
		Definition: &stored,
	}

	if isPull {
		output := locale.T(lang, "waiting_for_first_check")
		check.Output = &output
		check.NextRunAt = &now
	} else {
		output := locale.T(lang, "waiting_for_first_report")
		expiresAt := now.Add(time.Duration(definition.TTL) * time.Second)
		check.Output = &output
		check.TTLExpiresAt = &expiresAt
	}

	err = c.HealthCheckRepository.Create(check)
	if err != nil {
		log.Error().
			Err(err).
			Str("instanceId", instanceId.String()).
			Str("checkId", template.CheckId).
			Msg("Failed to create instance check")
		return err
	}

	log.Debug().
		Str("instanceId", instanceId.String()).
		Str("checkId", template.CheckId).
		Msg("Instance check created")

	return nil
}

// DeleteAgentCheck removes an agent check template and its copies on every instance.
func (c *Check) DeleteAgentCheck(workspaceId db.Id, agentName, checkId string) error {
	agent, err := c.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return err
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	template, err := c.AgentCheckRepository.GetByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("checkId", checkId).
			Msg("Failed to get agent check")
		return err
	}
	if template == nil {
		return ErrCheckNotFound
	}

	err = c.HealthCheckRepository.DeleteByAgentAndCheckId(agent.Id, checkId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("checkId", checkId).
			Msg("Failed to delete instance checks")
		return err
	}

	err = c.AgentCheckRepository.Delete(template.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("checkId", checkId).
			Msg("Failed to delete agent check")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("checkId", checkId).
		Msg("Agent check deleted")

	return nil
}

// ReportCheck records an instance's report for one of its TTL checks and restarts the TTL.
func (c *Check) ReportCheck(workspaceId db.Id, agentName, instanceId, checkId, status string, req *ReportCheckRequest, actor EventRequest) (*HealthCheckResponse, error) {
	agent, err := c.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	instance, err := c.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to get agent instance")
		return nil, err
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	check, err := c.HealthCheckRepository.GetByInstanceAndCheckId(instance.Id, checkId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Str("checkId", checkId).
			Msg("Failed to get instance check")
		return nil, err
	}
	if check == nil {
		return nil, ErrCheckNotFound
	}
	if check.Type != db.HealthCheckTypeTTL {
		return nil, ErrCheckNotReportable
	}

	var definition CheckDefinition
	if check.Definition != nil {
		if err := json.Unmarshal([]byte(*check.Definition), &definition); err != nil {
			log.Warn().
				Err(err).
				Str("checkId", check.CheckId).
				Msg("Failed to decode check definition")
		}
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(definition.TTL) * time.Second)
	previous := lo.Ternary(check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now), db.HealthCheckStatusCritical, check.Status)
	output := req.Output
	if len(output) > MaxCheckOutput {
		output = output[:MaxCheckOutput]
	}

	err = c.HealthCheckRepository.Report(check.Id, status, output, &expiresAt)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Str("checkId", checkId).
			Msg("Failed to report instance check")
		return nil, err
	}

	response := &HealthCheckResponse{
		CheckId:      check.CheckId,
		Name:         check.Name,
		Type:         check.Type,
		Source:       check.Source,
		Status:       status,
		TTLExpiresAt: &expiresAt,
		LastRunAt:    &now,
		Output:       lo.EmptyableToPtr(output),
	}

	log.Debug().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("instanceId", instanceId).
		Str("checkId", checkId).
		Str("status", status).
		Msg("Instance check reported")

	// A first result that passes is the expected start, not a change.
	if previous != status && !(previous == db.HealthCheckStatusUnknown && status == db.HealthCheckStatusPassing) {
		actor.ResourceType = db.EventResourceTypeHealthCheck
		actor.ResourceId = check.Id
		actor.Data = map[string]string{
			"agentName":  agentName,
			"instanceId": instanceId,
			"checkId":    checkId,
			"from":       previous,
			"to":         status,
			"output":     output,
		}

		c.TriggerEvent(workspaceId, EventAgentHealthChanged, actor)
	}

	return response, nil
}
