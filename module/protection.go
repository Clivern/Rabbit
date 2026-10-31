// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrProtectionNotFound      = errors.New("protection not found")
	ErrInvalidProtectionConfig = errors.New("invalid protection config")
)

// ProtectionConfig is the JSON shape stored in ai_agent_protection.config.
type ProtectionConfig struct {
	Header          string `json:"header,omitempty"`
	Username        string `json:"username,omitempty"`
	SecretEncrypted string `json:"secretEncrypted,omitempty"`
}

// SaveProtectionRequest is the body for setting or replacing an agent protection.
type SaveProtectionRequest struct {
	Type     string `json:"type" validate:"required,oneof=api_key basic_auth" label:"Type"`
	Username string `json:"username" validate:"omitempty,max=100" label:"Username"`
	Header   string `json:"header" validate:"omitempty,max=100" label:"Header"`
	Secret   string `json:"secret" validate:"omitempty,min=8,max=512" label:"Secret"`
}

// UpdateProtectionStatusRequest is the body for turning an agent protection on or off.
type UpdateProtectionStatusRequest struct {
	IsActive bool `json:"isActive" label:"Active"`
}

// ProtectionResponse is an agent protection shaped for API responses.
type ProtectionResponse struct {
	Id        db.Id  `json:"id"`
	Type      string `json:"type"`
	Username  string `json:"username,omitempty"`
	Header    string `json:"header,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Secret    string `json:"secret,omitempty"`
}

// Protection is the module for the credential the gateway uses to reach an agent.
type Protection struct {
	AgentRepository      db.AgentRepository
	ProtectionRepository db.AgentProtectionRepository
}

// NewProtection creates a protection module with the given repositories.
func NewProtection(agents db.AgentRepository, protections db.AgentProtectionRepository) *Protection {
	return &Protection{
		AgentRepository:      agents,
		ProtectionRepository: protections,
	}
}

// toProtectionResponse shapes a protection row for API responses, never its secret.
func toProtectionResponse(item *db.AgentProtection, config ProtectionConfig) *ProtectionResponse {
	return &ProtectionResponse{
		Id:        item.Id,
		Type:      item.Type,
		Username:  config.Username,
		Header:    config.Header,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// GetProtection returns the protection of an agent, never its secret.
func (p *Protection) GetProtection(workspaceId db.Id, agentName string) (*ProtectionResponse, error) {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent protection")
		return nil, err
	}
	if protection == nil {
		return nil, ErrProtectionNotFound
	}

	var config ProtectionConfig
	if err := json.Unmarshal([]byte(protection.Config), &config); err != nil {
		log.Warn().
			Err(err).
			Str("protectionId", protection.Id.String()).
			Msg("Failed to decode protection config")
	}

	return toProtectionResponse(protection, config), nil
}

// SaveProtection sets or replaces the protection of an agent; a generated secret is returned once.
func (p *Protection) SaveProtection(workspaceId db.Id, agentName string, req *SaveProtectionRequest) (*ProtectionResponse, error) {
	if (req.Type == db.AgentProtectionTypeBasicAuth) == (req.Username == "") {
		return nil, ErrInvalidProtectionConfig
	}

	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	secret := lo.Ternary(req.Secret == "", rand.Text(), req.Secret)
	config := ProtectionConfig{Username: req.Username}

	switch req.Type {
	case db.AgentProtectionTypeAPIKey:
		config.Header = lo.Ternary(req.Header == "", "X-API-Key", req.Header)
		config.SecretEncrypted, err = encrypt.Encrypt(secret)
	case db.AgentProtectionTypeBasicAuth:
		config.SecretEncrypted, err = encrypt.Encrypt(secret)
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to encrypt protection secret")
		return nil, err
	}

	raw, err := json.Marshal(config)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to encode protection config")
		return nil, err
	}

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent protection")
		return nil, err
	}

	if protection == nil {
		protection = &db.AgentProtection{AgentId: agent.Id, Type: req.Type, Config: string(raw), IsActive: true}
		err = p.ProtectionRepository.Create(protection)
	} else {
		protection.Type = req.Type
		protection.Config = string(raw)
		protection.IsActive = true
		err = p.ProtectionRepository.Update(protection)
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to save agent protection")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("protectionId", protection.Id.String()).
		Str("type", protection.Type).
		Msg("Agent protection saved")

	resp := toProtectionResponse(protection, config)
	resp.Secret = lo.Ternary(req.Secret == "", secret, "")

	return resp, nil
}

// UpdateProtectionStatus turns the protection of an agent on or off.
func (p *Protection) UpdateProtectionStatus(workspaceId db.Id, agentName string, req *UpdateProtectionStatusRequest) (*ProtectionResponse, error) {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent protection")
		return nil, err
	}
	if protection == nil {
		return nil, ErrProtectionNotFound
	}

	protection.IsActive = req.IsActive

	err = p.ProtectionRepository.Update(protection)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to update agent protection status")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("protectionId", protection.Id.String()).
		Bool("isActive", protection.IsActive).
		Msg("Agent protection status updated")

	var config ProtectionConfig
	if err := json.Unmarshal([]byte(protection.Config), &config); err != nil {
		log.Warn().
			Err(err).
			Str("protectionId", protection.Id.String()).
			Msg("Failed to decode protection config")
	}

	return toProtectionResponse(protection, config), nil
}

// DeleteProtection removes the protection of an agent.
func (p *Protection) DeleteProtection(workspaceId db.Id, agentName string) error {
	agent, err := p.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	protection, err := p.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to get agent protection")
		return err
	}
	if protection == nil {
		return ErrProtectionNotFound
	}

	err = p.ProtectionRepository.Delete(protection.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to delete agent protection")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("protectionId", protection.Id.String()).
		Msg("Agent protection deleted")

	return nil
}
