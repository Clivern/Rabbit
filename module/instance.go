// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

var (
	ErrInstanceNotFound         = errors.New("instance not found")
	ErrInvalidInstanceId        = errors.New("invalid instance id")
	ErrFailedListInstances      = errors.New("failed list instances")
	ErrFailedGetInstance        = errors.New("failed get instance")
	ErrFailedRegisterInstance   = errors.New("failed register instance")
	ErrFailedDeregisterInstance = errors.New("failed deregister instance")
)

var InstanceIdPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Instance is the module for agent instance registration and health.
type Instance struct {
	AgentRepository       db.AgentRepository
	InstanceRepository    db.AgentInstanceRepository
	HealthCheckRepository db.HealthCheckRepository
	AgentCheckRepository  db.AgentCheckRepository
	WorkspaceRepository   db.WorkspaceRepository
	FetchCard             func(ctx context.Context, address string, port int, path string) (map[string]any, error)
	TriggerEvent          EventTrigger
}

// RegisterInstanceRequest is the body for registering or updating an instance.
type RegisterInstanceRequest struct {
	Address string          `json:"address" validate:"required,max=255" label:"Address"`
	Port    int             `json:"port" validate:"required,min=1,max=65535" label:"Port"`
	Scheme  string          `json:"scheme" validate:"omitempty,oneof=http https" label:"Scheme"`
	Path    string          `json:"path" validate:"omitempty,max=255,startswith=/" label:"Path"`
	Meta    json.RawMessage `json:"meta" label:"Meta"`
}

// HealthCheckResponse is one health check on an instance.
type HealthCheckResponse struct {
	CheckId      string     `json:"checkId"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	Output       *string    `json:"output,omitempty"`
	TTLExpiresAt *time.Time `json:"ttlExpiresAt,omitempty"`
	LastRunAt    *time.Time `json:"lastRunAt,omitempty"`
}

// InstanceResponse is an instance shaped for API responses.
type InstanceResponse struct {
	Id         db.Id                  `json:"id"`
	AgentId    db.Id                  `json:"agentId"`
	InstanceId string                 `json:"instanceId"`
	Address    string                 `json:"address"`
	Port       int                    `json:"port"`
	Scheme     string                 `json:"scheme"`
	Path       string                 `json:"path"`
	Meta       json.RawMessage        `json:"meta,omitempty"`
	Status     string                 `json:"status"`
	Checks     []*HealthCheckResponse `json:"checks"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// ListInstancesResponse is one page of an agent's instances and the total across all pages.
type ListInstancesResponse struct {
	Instances []*InstanceResponse
	Total     int64
}

// NewInstance creates an instance module with the given repositories.
func NewInstance(agents db.AgentRepository, instances db.AgentInstanceRepository, checks db.HealthCheckRepository, agentChecks db.AgentCheckRepository, workspaces db.WorkspaceRepository) *Instance {
	return &Instance{
		AgentRepository:       agents,
		InstanceRepository:    instances,
		HealthCheckRepository: checks,
		AgentCheckRepository:  agentChecks,
		WorkspaceRepository:   workspaces,
		FetchCard:             NewCard().Fetch,
		TriggerEvent:          TriggerEvent,
	}
}

// toInstanceResponse shapes an instance row for API responses.
func toInstanceResponse(instance *db.AgentInstance, status string, checks []*HealthCheckResponse) *InstanceResponse {
	return &InstanceResponse{
		Id:         instance.Id,
		AgentId:    instance.AgentId,
		InstanceId: instance.InstanceId,
		Address:    instance.Address,
		Port:       instance.Port,
		Scheme:     instance.Scheme,
		Path:       instance.Path,
		Meta:       util.JSONRawFromString(instance.Meta),
		Status:     status,
		Checks:     checks,
		CreatedAt:  instance.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  instance.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListInstances returns one page of an agent's instances.
func (i *Instance) ListInstances(workspaceId db.Id, agentName string, limit, offset int) (*ListInstancesResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	total, err := i.InstanceRepository.CountByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to count agent instances")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	instances, err := i.InstanceRepository.ListPageByAgentId(agent.Id, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to list agent instances")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	list := make([]*InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		status, checks, err := i.GetHealth(instance.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("agentName", agentName).
				Str("instanceId", instance.InstanceId).
				Msg("Failed to get instance health")
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		list = append(list, toInstanceResponse(instance, status, checks))
	}

	return &ListInstancesResponse{Instances: list, Total: total}, nil
}

// ListHealthyInstances returns one page of an agent's instances with no critical, unknown, or expired checks.
func (i *Instance) ListHealthyInstances(workspaceId db.Id, agentName string, limit, offset int) (*ListInstancesResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	now := time.Now().UTC()

	total, err := i.InstanceRepository.CountLiveByAgentId(agent.Id, now)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to count healthy agent instances")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	instances, err := i.InstanceRepository.ListLivePageByAgentId(agent.Id, now, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to list healthy agent instances")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
	}

	list := make([]*InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		status, checks, err := i.GetHealth(instance.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("agentName", agentName).
				Str("instanceId", instance.InstanceId).
				Msg("Failed to get instance health")
			return nil, fmt.Errorf("%w: %v", ErrFailedListInstances, err)
		}

		list = append(list, toInstanceResponse(instance, status, checks))
	}

	return &ListInstancesResponse{Instances: list, Total: total}, nil
}

// GetInstance returns one instance of an agent.
func (i *Instance) GetInstance(workspaceId db.Id, agentName, instanceId string) (*InstanceResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to get agent instance")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}
	if instance == nil {
		return nil, ErrInstanceNotFound
	}

	status, checks, err := i.GetHealth(instance.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instance.InstanceId).
			Msg("Failed to get instance health")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInstance, err)
	}

	return toInstanceResponse(instance, status, checks), nil
}

// RegisterInstance creates or updates an instance and syncs its checks with the agent's templates.
func (i *Instance) RegisterInstance(workspaceId db.Id, agentName, instanceId string, req *RegisterInstanceRequest, lang string, actor EventRequest) (*InstanceResponse, bool, error) {
	if !InstanceIdPattern.MatchString(instanceId) {
		return nil, false, ErrInvalidInstanceId
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}
	if agent == nil {
		return nil, false, ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to get agent instance")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	created := instance == nil
	if created {
		instance = &db.AgentInstance{AgentId: agent.Id, InstanceId: instanceId}
	}

	// A re-register to the same endpoint is routine and raises no event.
	endpoint := fmt.Sprintf("%s://%s:%d%s", instance.Scheme, instance.Address, instance.Port, instance.Path)

	instance.Address = req.Address
	instance.Port = req.Port
	instance.Scheme = lo.CoalesceOrEmpty(req.Scheme, "http")
	instance.Path = strings.TrimSuffix(req.Path, "/")

	moved := endpoint != fmt.Sprintf("%s://%s:%d%s", instance.Scheme, instance.Address, instance.Port, instance.Path)
	instance.Meta = nil
	if len(req.Meta) > 0 {
		meta := string(req.Meta)
		instance.Meta = &meta
	}

	instance.Status = db.AgentInstanceStatusPassing

	if created {
		err = i.InstanceRepository.Create(instance)
	} else {
		err = i.InstanceRepository.Update(instance)
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to save agent instance")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	err = i.SyncChecks(instance.Id, agent.Id, lang)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to sync instance checks")
		return nil, false, err
	}

	// The first instance to register supplies the card; a failed fetch leaves it to the next one.
	if agent.Card == nil {
		card, err := i.FetchCard(context.Background(), instance.Address, instance.Port, agent.CardPath)
		if err != nil {
			log.Warn().
				Err(err).
				Str("agentId", agent.Id.String()).
				Str("instanceId", instanceId).
				Msg("Failed to fetch agent card")
		} else {
			gateway := fmt.Sprintf(
				"%s/api/v1/workspaces/%s/agents/%s",
				viper.GetString("app.url"),
				workspaceId,
				agent.Name,
			)

			interfaces, _ := card["supportedInterfaces"].([]any)

			if len(interfaces) == 0 {
				interfaces = []any{map[string]any{}}
			}

			for n, iface := range interfaces {
				entry, ok := iface.(map[string]any)
				if !ok {
					entry = map[string]any{}
					interfaces[n] = entry
				}
				entry["url"] = gateway
			}

			card["supportedInterfaces"] = interfaces

			if lo.HasKey(card, "url") {
				card["url"] = gateway
			}

			cardJSON, _ := json.Marshal(card)
			sum := sha256.Sum256(cardJSON)

			_, err = i.AgentRepository.SetCardIfMissing(agent.Id, string(cardJSON), hex.EncodeToString(sum[:]))
			if err != nil {
				log.Error().
					Err(err).
					Str("agentId", agent.Id.String()).
					Msg("Failed to store agent card")
			} else {
				log.Info().
					Str("agentId", agent.Id.String()).
					Str("instanceId", instanceId).
					Msg("Agent card stored")
			}
		}
	}

	instance.UpdatedAt = time.Now().UTC()
	status, checks, err := i.GetHealth(instance.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instance.InstanceId).
			Msg("Failed to get instance health")
		return nil, false, fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	item := toInstanceResponse(instance, status, checks)

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("instanceId", instanceId).
		Str("address", instance.Address).
		Int("port", instance.Port).
		Bool("created", created).
		Msg("Agent instance registered")

	if created || moved {
		actor.ResourceType = db.EventResourceTypeAgentInstance
		actor.ResourceId = instance.Id
		actor.Data = map[string]string{
			"action":     lo.Ternary(created, "registered", "updated"),
			"agentName":  agentName,
			"instanceId": instanceId,
			"address":    instance.Address,
			"port":       fmt.Sprint(instance.Port),
			"scheme":     instance.Scheme,
			"path":       instance.Path,
		}

		i.TriggerEvent(workspaceId, EventAgentInstanceChanged, actor)
	}

	return item, created, nil
}

// SyncChecks makes an instance's copied checks match its agent's templates.
func (i *Instance) SyncChecks(instanceId, agentId db.Id, lang string) error {
	now := time.Now().UTC()

	templates, err := i.AgentCheckRepository.ListByAgentId(agentId)
	if err != nil {
		log.Error().
			Err(err).
			Str("agentId", agentId.String()).
			Str("instanceId", instanceId.String()).
			Msg("Failed to list agent checks")
		return fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	current, err := i.HealthCheckRepository.ListByInstanceId(instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("instanceId", instanceId.String()).
			Msg("Failed to list instance checks")
		return fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
	}

	// Both lists are loaded once. Check id is the key that joins a template to its copy.
	templatesById := make(map[string]*db.AgentCheck, len(templates))
	for _, template := range templates {
		templatesById[template.CheckId] = template
	}

	existingById := make(map[string]*db.HealthCheck, len(current))
	for _, check := range current {
		existingById[check.CheckId] = check
	}

	for _, template := range templates {
		existing := existingById[template.CheckId]

		// A type change resets the check, since its old status no longer means anything.
		if existing != nil && existing.Type != template.Type {
			err = i.HealthCheckRepository.Delete(existing.Id)
			if err != nil {
				log.Error().
					Err(err).
					Str("instanceId", instanceId.String()).
					Str("checkId", template.CheckId).
					Msg("Failed to reset instance check")
				return fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
			}

			log.Info().
				Str("instanceId", instanceId.String()).
				Str("checkId", template.CheckId).
				Str("fromType", existing.Type).
				Str("toType", template.Type).
				Msg("Instance check reset after type change")

			existing = nil
		}

		// http and tcp are pulled on a schedule. Anything else waits for a report within its TTL.
		isPull := template.Type == db.HealthCheckTypeHTTP || template.Type == db.HealthCheckTypeTCP
		stored := template.Definition

		if existing != nil {
			// Same type: keep status and output, and copy the template's name and definition.
			existing.Name = template.Name
			existing.Definition = &stored
			existing.NextRunAt = nil
			if isPull {
				existing.NextRunAt = &now
			}

			err = i.HealthCheckRepository.Update(existing)
		} else {
			// Missing, or just reset: start unknown until the first pull or report.
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
				var templateDefinition CheckDefinition
				if err := json.Unmarshal([]byte(template.Definition), &templateDefinition); err != nil {
					log.Warn().
						Err(err).
						Str("checkId", template.CheckId).
						Msg("Failed to decode check definition")
				}
				output := locale.T(lang, "waiting_for_first_report")
				expiresAt := now.Add(time.Duration(templateDefinition.TTL) * time.Second)
				check.Output = &output
				check.TTLExpiresAt = &expiresAt
			}

			err = i.HealthCheckRepository.Create(check)
		}
		if err != nil {
			log.Error().
				Err(err).
				Str("instanceId", instanceId.String()).
				Str("checkId", template.CheckId).
				Msg("Failed to save instance check")
			return fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
		}
	}

	// Agent copies with no remaining template are removed. Other sources are left alone.
	for _, check := range current {
		if check.Source == db.HealthCheckSourceAgent && templatesById[check.CheckId] == nil {
			err = i.HealthCheckRepository.Delete(check.Id)
			if err != nil {
				log.Error().
					Err(err).
					Str("instanceId", instanceId.String()).
					Str("checkId", check.CheckId).
					Msg("Failed to remove orphaned instance check")
				return fmt.Errorf("%w: %v", ErrFailedRegisterInstance, err)
			}

			log.Info().
				Str("instanceId", instanceId.String()).
				Str("checkId", check.CheckId).
				Msg("Orphaned instance check removed")
		}
	}

	return nil
}

// DeregisterInstance removes an instance and its health checks.
func (i *Instance) DeregisterInstance(workspaceId db.Id, agentName, instanceId string, actor EventRequest) error {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent")
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}
	if agent == nil {
		return ErrAgentNotFound
	}

	instance, err := i.InstanceRepository.GetByAgentAndInstanceId(agent.Id, instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to get agent instance")
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}
	if instance == nil {
		return ErrInstanceNotFound
	}

	err = i.InstanceRepository.Delete(instance.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("instanceId", instanceId).
			Msg("Failed to delete agent instance")
		return fmt.Errorf("%w: %v", ErrFailedDeregisterInstance, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("instanceId", instanceId).
		Msg("Agent instance deregistered")

	actor.ResourceType = db.EventResourceTypeAgentInstance
	actor.ResourceId = instance.Id
	actor.Data = map[string]string{
		"action":     "deregistered",
		"agentName":  agentName,
		"instanceId": instanceId,
	}

	i.TriggerEvent(workspaceId, EventAgentInstanceChanged, actor)

	return nil
}

// GetHealth returns the instance health and its checks, the worst status across checks.
func (i *Instance) GetHealth(instanceId db.Id) (string, []*HealthCheckResponse, error) {
	checks, err := i.HealthCheckRepository.ListByInstanceId(instanceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("instanceId", instanceId.String()).
			Msg("Failed to list instance checks")
		return "", nil, err
	}

	now := time.Now().UTC()
	status := db.AgentInstanceStatusPassing
	mapped := make([]*HealthCheckResponse, 0, len(checks))
	for _, check := range checks {
		current := lo.Ternary(
			check.Type == db.HealthCheckTypeTTL && (check.TTLExpiresAt == nil || !check.TTLExpiresAt.After(now)),
			db.HealthCheckStatusCritical,
			check.Status,
		)

		// The instance status is the worst check status. A TTL check past its TTL
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

		mapped = append(mapped, &HealthCheckResponse{
			CheckId:      check.CheckId,
			Name:         check.Name,
			Type:         check.Type,
			Source:       check.Source,
			Status:       current,
			Output:       check.Output,
			TTLExpiresAt: check.TTLExpiresAt,
			LastRunAt:    check.LastRunAt,
		})
	}

	return status, mapped, nil
}
