// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/util"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// DefaultIdentityTokenTTL is how long a client credentials token lives, in seconds.
const DefaultIdentityTokenTTL = 300

var (
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrIdentityAlreadyExists = errors.New("identity already exists")
	ErrInvalidIdentityConfig = errors.New("invalid identity config")
	ErrInvalidClient         = errors.New("invalid client")
	ErrInvalidAgentIdentity  = errors.New("invalid agent identity")
)

// IdentityConfig is the JSON shape stored in ai_agent_identity.config.
type IdentityConfig struct {
	Username   string `json:"username,omitempty"`
	SecretHash string `json:"secretHash"`
	TokenTTL   int    `json:"tokenTtl,omitempty"`
}

// CreateIdentityRequest is the body for creating an agent identity.
type CreateIdentityRequest struct {
	Name     string `json:"name" validate:"required,max=100" label:"Name"`
	Type     string `json:"type" validate:"required,oneof=api_key basic_auth client_credentials" label:"Type"`
	Username string `json:"username" validate:"omitempty,max=100" label:"Username"`
	TokenTTL int    `json:"tokenTtl" validate:"omitempty,min=60,max=86400" label:"Token TTL"`
	Secret   string `json:"secret" validate:"omitempty,min=8,max=512" label:"Secret"`
}

// UpdateIdentityRequest is the body for updating an agent identity.
type UpdateIdentityRequest struct {
	Name     string `json:"name" validate:"required,max=100" label:"Name"`
	IsActive bool   `json:"isActive" label:"Active"`
}

// IdentityResponse is an agent identity shaped for API responses.
type IdentityResponse struct {
	Id        db.Id  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Username  string `json:"username,omitempty"`
	TokenTTL  int    `json:"tokenTtl,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Secret    string `json:"secret,omitempty"`
}

// IssueTokenResponse is the OAuth 2.0 token response for the client credentials grant.
type IssueTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// AgentCredential is what a calling agent presents: its name with an API key or basic auth, or a bearer token.
type AgentCredential struct {
	SourceAgent string
	APIKey      string
	Username    string
	Password    string
	Token       string
}

// Identity is the module for the credentials an agent uses to call through the gateway.
type Identity struct {
	AgentRepository    db.AgentRepository
	IdentityRepository db.AgentIdentityRepository
	TokenRepository    db.AgentIdentityTokenRepository
}

// NewIdentity creates an identity module with the given repositories.
func NewIdentity(agents db.AgentRepository, identities db.AgentIdentityRepository, tokens db.AgentIdentityTokenRepository) *Identity {
	return &Identity{
		AgentRepository:    agents,
		IdentityRepository: identities,
		TokenRepository:    tokens,
	}
}

// toIdentityResponse shapes an identity row for API responses, never its secret.
func toIdentityResponse(item *db.AgentIdentity, config IdentityConfig) *IdentityResponse {
	return &IdentityResponse{
		Id:        item.Id,
		Name:      item.Name,
		Type:      item.Type,
		Username:  config.Username,
		TokenTTL:  config.TokenTTL,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListIdentities returns the identities of an agent, never their secrets.
func (i *Identity) ListIdentities(workspaceId db.Id, agentName string) ([]*IdentityResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	identities, err := i.IdentityRepository.ListByAgentId(agent.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to list agent identities")
		return nil, err
	}

	list := make([]*IdentityResponse, 0, len(identities))
	for _, identity := range identities {
		var config IdentityConfig
		if err := json.Unmarshal([]byte(identity.Config), &config); err != nil {
			log.Warn().
				Err(err).
				Str("identityId", identity.Id.String()).
				Msg("Failed to decode identity config")
		}

		list = append(list, toIdentityResponse(identity, config))
	}

	return list, nil
}

// CreateIdentity issues a new identity for an agent; a generated secret is returned once.
func (i *Identity) CreateIdentity(workspaceId db.Id, agentName string, req *CreateIdentityRequest) (*IdentityResponse, error) {
	if (req.Type == db.AgentIdentityTypeBasicAuth) == (req.Username == "") {
		return nil, ErrInvalidIdentityConfig
	}

	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	existing, err := i.IdentityRepository.GetByAgentAndName(agent.Id, req.Name)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("identityName", req.Name).
			Msg("Failed to get agent identity")
		return nil, err
	}
	if existing != nil {
		return nil, ErrIdentityAlreadyExists
	}

	secret := lo.Ternary(req.Secret == "", rand.Text(), req.Secret)

	hash, err := util.HashPassword(secret)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to hash identity secret")
		return nil, err
	}

	config := IdentityConfig{Username: req.Username, SecretHash: hash}
	if req.Type == db.AgentIdentityTypeClientCredentials {
		config.TokenTTL = lo.Ternary(req.TokenTTL == 0, DefaultIdentityTokenTTL, req.TokenTTL)
	}

	raw, err := json.Marshal(config)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to encode identity config")
		return nil, err
	}

	identity := &db.AgentIdentity{
		AgentId:  agent.Id,
		Name:     req.Name,
		Type:     req.Type,
		Config:   string(raw),
		IsActive: true,
	}

	err = i.IdentityRepository.Create(identity)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Msg("Failed to create agent identity")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("identityId", identity.Id.String()).
		Str("type", identity.Type).
		Msg("Agent identity created")

	resp := toIdentityResponse(identity, config)
	resp.Secret = lo.Ternary(req.Secret == "", secret, "")

	return resp, nil
}

// UpdateIdentity renames an agent identity or turns it on or off.
func (i *Identity) UpdateIdentity(workspaceId db.Id, agentName string, id db.Id, req *UpdateIdentityRequest) (*IdentityResponse, error) {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	identity, err := i.IdentityRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("identityId", id.String()).
			Msg("Failed to get agent identity")
		return nil, err
	}
	if identity == nil || identity.AgentId != agent.Id {
		return nil, ErrIdentityNotFound
	}

	if req.Name != identity.Name {
		existing, err := i.IdentityRepository.GetByAgentAndName(agent.Id, req.Name)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("agentName", agentName).
				Str("identityName", req.Name).
				Msg("Failed to get agent identity")
			return nil, err
		}
		if existing != nil {
			return nil, ErrIdentityAlreadyExists
		}
	}

	identity.Name = req.Name
	identity.IsActive = req.IsActive

	err = i.IdentityRepository.Update(identity)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("identityId", id.String()).
			Msg("Failed to update agent identity")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("identityId", identity.Id.String()).
		Bool("isActive", identity.IsActive).
		Msg("Agent identity updated")

	var config IdentityConfig
	if err := json.Unmarshal([]byte(identity.Config), &config); err != nil {
		log.Warn().
			Err(err).
			Str("identityId", identity.Id.String()).
			Msg("Failed to decode identity config")
	}

	resp := toIdentityResponse(identity, config)
	resp.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	return resp, nil
}

// DeleteIdentity revokes an agent identity.
func (i *Identity) DeleteIdentity(workspaceId db.Id, agentName string, id db.Id) error {
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
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

	identity, err := i.IdentityRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("identityId", id.String()).
			Msg("Failed to get agent identity")
		return err
	}
	if identity == nil || identity.AgentId != agent.Id {
		return ErrIdentityNotFound
	}

	err = i.IdentityRepository.Delete(identity.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("agentName", agentName).
			Str("identityId", identity.Id.String()).
			Msg("Failed to delete agent identity")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("agentName", agentName).
		Str("identityId", identity.Id.String()).
		Msg("Agent identity deleted")

	return nil
}

// IssueToken exchanges a client credentials identity's id and secret for a short-lived access token.
func (i *Identity) IssueToken(clientId, clientSecret string) (*IssueTokenResponse, error) {
	if uuid.Validate(clientId) != nil {
		log.Warn().
			Msg("Token request with malformed client id")
		return nil, ErrInvalidClient
	}

	identity, err := i.IdentityRepository.GetById(db.Id(clientId))
	if err != nil {
		log.Error().
			Err(err).
			Str("clientId", clientId).
			Msg("Failed to get agent identity")
		return nil, err
	}
	if identity == nil || identity.Type != db.AgentIdentityTypeClientCredentials || !identity.IsActive {
		log.Warn().
			Str("clientId", clientId).
			Msg("Token request for unknown or inactive client")
		return nil, ErrInvalidClient
	}

	var config IdentityConfig
	if err := json.Unmarshal([]byte(identity.Config), &config); err != nil {
		log.Warn().
			Err(err).
			Str("identityId", identity.Id.String()).
			Msg("Failed to decode identity config")
	}

	if !util.ComparePassword(config.SecretHash, clientSecret) {
		log.Warn().
			Str("clientId", clientId).
			Msg("Token request with invalid client secret")
		return nil, ErrInvalidClient
	}

	count, err := i.TokenRepository.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired identity tokens")
		return nil, err
	}
	if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired identity tokens deleted")
	}

	accessToken := rand.Text()
	sum := sha256.Sum256([]byte(accessToken))

	err = i.TokenRepository.Create(&db.AgentIdentityToken{
		IdentityId: identity.Id,
		TokenHash:  hex.EncodeToString(sum[:]),
		ExpiresAt:  time.Now().UTC().Add(time.Duration(config.TokenTTL) * time.Second),
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("clientId", clientId).
			Msg("Failed to create identity token")
		return nil, err
	}

	log.Info().
		Str("clientId", clientId).
		Int("expiresIn", config.TokenTTL).
		Msg("Identity token issued")

	return &IssueTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   config.TokenTTL,
	}, nil
}

// Authenticate verifies an agent credential and returns the calling agent of the workspace.
func (i *Identity) Authenticate(workspaceId db.Id, credential *AgentCredential) (*db.Agent, error) {
	// X-Source-Agent only names the agent; one of its identities must still accept the secret.
	agent, err := i.AgentRepository.GetByWorkspaceAndName(workspaceId, credential.SourceAgent)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrInvalidAgentIdentity
	}

	// Disabled identities never authenticate.
	identities, err := i.IdentityRepository.ListActiveByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}

	for _, identity := range identities {
		var config IdentityConfig
		_ = json.Unmarshal([]byte(identity.Config), &config)

		switch identity.Type {
		case db.AgentIdentityTypeAPIKey:
			// The key is stored as a bcrypt hash.
			if util.ComparePassword(config.SecretHash, credential.APIKey) {
				return agent, nil
			}
		case db.AgentIdentityTypeBasicAuth:
			// The username is stored in plain text and the password as a bcrypt hash.
			if credential.Username == config.Username && util.ComparePassword(config.SecretHash, credential.Password) {
				return agent, nil
			}
		case db.AgentIdentityTypeClientCredentials:
			// Tokens are stored as SHA-256 hashes so they can be looked up directly.
			sum := sha256.Sum256([]byte(credential.Token))

			issued, err := i.TokenRepository.GetByTokenHash(hex.EncodeToString(sum[:]))
			if err != nil {
				return nil, err
			}
			// The token must be issued to this identity and not expired.
			if issued != nil && issued.IdentityId == identity.Id && issued.ExpiresAt.After(time.Now().UTC()) {
				return agent, nil
			}
		}
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("sourceAgent", credential.SourceAgent).
		Msg("Agent failed authentication")

	return nil, ErrInvalidAgentIdentity
}
