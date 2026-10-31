// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrIntegrationNotFound      = errors.New("integration not found")
	ErrInvalidIntegrationConfig = errors.New("invalid integration config")
	ErrFailedListIntegrations   = errors.New("failed list integrations")
	ErrFailedGetIntegration     = errors.New("failed get integration")
	ErrFailedCreateIntegration  = errors.New("failed create integration")
	ErrFailedUpdateIntegration  = errors.New("failed update integration")
	ErrFailedDeleteIntegration  = errors.New("failed delete integration")
	IntegrationEvents           = []string{
		EventAuditLog,
		EventAgentHealthChanged,
		EventAgentInstanceChanged,
		EventCatalogChanged,
		EventGatewayCallDenied,
		EventSystemHealthDegraded,
	}
	DatadogSites = []string{
		"datadoghq.com",
		"us3.datadoghq.com",
		"us5.datadoghq.com",
		"datadoghq.eu",
		"ap1.datadoghq.com",
		"ddog-gov.com",
	}
	PagerDutySites = []string{
		"events.pagerduty.com",
		"events.eu.pagerduty.com",
	}
	PagerDutySeverities = map[string]string{
		EventAuditLog:             "info",
		EventAgentHealthChanged:   "critical",
		EventAgentInstanceChanged: "info",
		EventCatalogChanged:       "info",
		EventGatewayCallDenied:    "warning",
		EventSystemHealthDegraded: "critical",
	}
)

// Integration is the module for workspace event destinations.
type Integration struct {
	IntegrationRepository db.IntegrationRepository
}

// IntegrationConfig is the JSON shape stored in integrations.config.
type IntegrationConfig struct {
	URL             string   `json:"url,omitempty"`
	Site            string   `json:"site,omitempty"`
	Index           string   `json:"index,omitempty"`
	Events          []string `json:"events"`
	SecretEncrypted string   `json:"secretEncrypted,omitempty"`
}

// SaveIntegrationRequest is the body for creating or updating an integration.
type SaveIntegrationRequest struct {
	Type     string   `json:"type" validate:"required,oneof=webhook slack datadog splunk pagerduty" label:"Type"`
	Name     string   `json:"name" validate:"required,max=60" label:"Name"`
	URL      string   `json:"url" validate:"omitempty,http_url,max=2048" label:"URL"`
	Site     string   `json:"site" validate:"omitempty,max=100" label:"Site"`
	Index    string   `json:"index" validate:"omitempty,max=100" label:"Index"`
	Secret   string   `json:"secret" validate:"omitempty,min=8,max=512" label:"Secret"`
	Events   []string `json:"events" validate:"required,min=1,unique,dive,oneof=audit.log agent.health.changed agent.instance.changed catalog.changed gateway.call.denied system.health.degraded" label:"Events"`
	IsActive bool     `json:"isActive" label:"Active"`
}

// IntegrationResponse is an integration shaped for API responses, never with a stored secret.
type IntegrationResponse struct {
	Id          db.Id    `json:"id"`
	WorkspaceId db.Id    `json:"workspaceId"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	URL         string   `json:"url,omitempty"`
	Site        string   `json:"site,omitempty"`
	Index       string   `json:"index,omitempty"`
	Events      []string `json:"events"`
	HasSecret   bool     `json:"hasSecret"`
	IsActive    bool     `json:"isActive"`
	Secret      string   `json:"secret,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// ListIntegrationsResponse is returned when listing integrations.
type ListIntegrationsResponse struct {
	Integrations []*IntegrationResponse
	Total        int64
}

// NewIntegration creates an integration module with the given repository.
func NewIntegration(integrations db.IntegrationRepository) *Integration {
	return &Integration{
		IntegrationRepository: integrations,
	}
}

// toIntegrationResponse shapes an integration row for API responses.
func toIntegrationResponse(item *db.Integration, config IntegrationConfig) *IntegrationResponse {
	return &IntegrationResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Type:        item.Type,
		Name:        item.Name,
		URL:         config.URL,
		Site:        config.Site,
		Index:       config.Index,
		Events:      config.Events,
		HasSecret:   config.SecretEncrypted != "",
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListIntegrations returns paginated integrations in a workspace.
func (i *Integration) ListIntegrations(workspaceId db.Id, limit, offset int) (*ListIntegrationsResponse, error) {
	total, err := i.IntegrationRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count integrations")
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntegrations, err)
	}

	integrations, err := i.IntegrationRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list integrations")
		return nil, fmt.Errorf("%w: %v", ErrFailedListIntegrations, err)
	}

	list := make([]*IntegrationResponse, 0, len(integrations))
	for _, item := range integrations {
		var config IntegrationConfig
		if err := json.Unmarshal([]byte(lo.FromPtr(item.Config)), &config); err != nil {
			log.Warn().
				Err(err).
				Str("integrationId", item.Id.String()).
				Msg("Failed to decode integration config")
		}

		list = append(list, toIntegrationResponse(item, config))
	}

	return &ListIntegrationsResponse{Integrations: list, Total: total}, nil
}

// GetIntegration returns one integration by id.
func (i *Integration) GetIntegration(workspaceId, integrationId db.Id) (*IntegrationResponse, error) {
	integration, err := i.IntegrationRepository.GetById(integrationId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to get integration")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetIntegration, err)
	}
	if integration == nil || integration.WorkspaceId != workspaceId {
		return nil, ErrIntegrationNotFound
	}

	var config IntegrationConfig
	if err := json.Unmarshal([]byte(lo.FromPtr(integration.Config)), &config); err != nil {
		log.Warn().
			Err(err).
			Str("integrationId", integration.Id.String()).
			Msg("Failed to decode integration config")
	}

	return toIntegrationResponse(integration, config), nil
}

// CreateIntegration stores a destination; a generated webhook signing secret is returned once.
func (i *Integration) CreateIntegration(workspaceId db.Id, req *SaveIntegrationRequest) (*IntegrationResponse, error) {
	switch req.Type {
	case db.IntegrationTypeWebhook, db.IntegrationTypeSlack:
		if req.URL == "" {
			return nil, ErrInvalidIntegrationConfig
		}
	case db.IntegrationTypeDatadog:
		if !lo.Contains(DatadogSites, req.Site) || req.Secret == "" {
			return nil, ErrInvalidIntegrationConfig
		}
	case db.IntegrationTypeSplunk:
		if req.URL == "" || req.Secret == "" {
			return nil, ErrInvalidIntegrationConfig
		}
	case db.IntegrationTypePagerDuty:
		if !lo.Contains(PagerDutySites, req.Site) || req.Secret == "" {
			return nil, ErrInvalidIntegrationConfig
		}
	}

	config := IntegrationConfig{
		URL:             req.URL,
		Site:            req.Site,
		Index:           req.Index,
		Events:          req.Events,
		SecretEncrypted: "",
	}

	var err error
	secret := req.Secret
	if req.Type == db.IntegrationTypeWebhook && secret == "" {
		secret = "whsec_" + rand.Text()
	}
	if secret != "" {
		config.SecretEncrypted, err = encrypt.Encrypt(secret)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Msg("Failed to encrypt integration secret")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateIntegration, err)
		}
	}

	raw, err := json.Marshal(config)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to encode integration config")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateIntegration, err)
	}

	integration := &db.Integration{
		WorkspaceId: workspaceId,
		Type:        req.Type,
		Name:        req.Name,
		Config:      new(string(raw)),
		IsActive:    req.IsActive,
	}

	err = i.IntegrationRepository.Create(integration)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create integration")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateIntegration, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("integrationId", integration.Id.String()).
		Str("type", integration.Type).
		Msg("Integration created")

	resp := toIntegrationResponse(integration, config)
	resp.Secret = lo.Ternary(req.Secret == "", secret, "")

	return resp, nil
}

// UpdateIntegration replaces a destination's settings; an empty secret keeps the stored one.
func (i *Integration) UpdateIntegration(workspaceId, integrationId db.Id, req *SaveIntegrationRequest) (*IntegrationResponse, error) {
	integration, err := i.IntegrationRepository.GetById(integrationId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to get integration")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntegration, err)
	}
	if integration == nil || integration.WorkspaceId != workspaceId {
		return nil, ErrIntegrationNotFound
	}

	switch {
	case req.Type != integration.Type:
		return nil, ErrInvalidIntegrationConfig
	case req.Type == db.IntegrationTypeDatadog && !lo.Contains(DatadogSites, req.Site):
		return nil, ErrInvalidIntegrationConfig
	case req.Type == db.IntegrationTypePagerDuty && !lo.Contains(PagerDutySites, req.Site):
		return nil, ErrInvalidIntegrationConfig
	case lo.Contains([]string{db.IntegrationTypeWebhook, db.IntegrationTypeSlack, db.IntegrationTypeSplunk}, req.Type) && req.URL == "":
		return nil, ErrInvalidIntegrationConfig
	}

	var stored IntegrationConfig
	if err := json.Unmarshal([]byte(lo.FromPtr(integration.Config)), &stored); err != nil {
		log.Warn().
			Err(err).
			Str("integrationId", integration.Id.String()).
			Msg("Failed to decode integration config")
	}

	config := IntegrationConfig{
		URL:             req.URL,
		Site:            req.Site,
		Index:           req.Index,
		Events:          req.Events,
		SecretEncrypted: stored.SecretEncrypted,
	}
	if req.Secret != "" {
		config.SecretEncrypted, err = encrypt.Encrypt(req.Secret)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("integrationId", integrationId.String()).
				Msg("Failed to encrypt integration secret")
			return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntegration, err)
		}
	}

	raw, err := json.Marshal(config)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to encode integration config")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntegration, err)
	}

	integration.Name = req.Name
	integration.Config = new(string(raw))
	integration.IsActive = req.IsActive

	err = i.IntegrationRepository.Update(integration)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to update integration")
		return nil, fmt.Errorf("%w: %v", ErrFailedUpdateIntegration, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("integrationId", integration.Id.String()).
		Str("type", integration.Type).
		Bool("isActive", integration.IsActive).
		Msg("Integration updated")

	resp := toIntegrationResponse(integration, config)
	resp.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return resp, nil
}

// DeleteIntegration removes a destination.
func (i *Integration) DeleteIntegration(workspaceId, integrationId db.Id) error {
	integration, err := i.IntegrationRepository.GetById(integrationId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to get integration")
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntegration, err)
	}

	if integration == nil || integration.WorkspaceId != workspaceId {
		return ErrIntegrationNotFound
	}

	err = i.IntegrationRepository.Delete(integration.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("integrationId", integrationId.String()).
			Msg("Failed to delete integration")
		return fmt.Errorf("%w: %v", ErrFailedDeleteIntegration, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("integrationId", integrationId.String()).
		Msg("Integration deleted")

	return nil
}
