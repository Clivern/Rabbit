// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrInvalidExpiresAt = errors.New("invalid expires at format")
)

// APIKey is the module for user API key CRUD.
type APIKey struct {
	APIKeyRepository db.APIKeyRepository
}

// NewAPIKey creates an API key module with the given repository.
func NewAPIKey(apiKeys db.APIKeyRepository) *APIKey {
	return &APIKey{APIKeyRepository: apiKeys}
}

// toAPIKeyResponse shapes an API key row for API responses. An empty key hides it.
func toAPIKeyResponse(item *db.APIKey, key string) *APIKeyResponse {
	return &APIKeyResponse{
		Id:        item.Id,
		Name:      item.Name,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
		Key:       key,
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}
}

// ListAPIKeysResult is what you get back when listing API keys.
type ListAPIKeysResult struct {
	APIKeys []*APIKeyResponse
	Total   int64
}

// CreateAPIKeyRequest is what you pass when creating an API key.
type CreateAPIKeyRequest struct {
	Name      string `json:"name" validate:"required,max=255" label:"Name"`
	ExpiresAt string `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
}

// APIKeyResponse is an API key shaped for API responses.
type APIKeyResponse struct {
	Id        db.Id   `json:"id"`
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expiresAt"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	Key       string  `json:"key,omitempty"`
}

// CreateAPIKey creates an API key for the user; the raw key is returned once.
func (a *APIKey) CreateAPIKey(req *CreateAPIKeyRequest, user *db.User) (*APIKeyResponse, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to generate API key")
		return nil, err
	}

	var expiresAt *time.Time
	if lo.IsNotEmpty(req.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, ErrInvalidExpiresAt
		}

		expiresAt = new(t)
	}

	apiKey := &db.APIKey{
		UserId:    user.Id,
		Name:      req.Name,
		Key:       id.String(),
		ExpiresAt: expiresAt,
	}

	err = a.APIKeyRepository.Create(apiKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create API key")
		return nil, err
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("apiKeyId", apiKey.Id.String()).
		Msg("API key created")

	return toAPIKeyResponse(apiKey, apiKey.Key), nil
}

// ListAPIKeys returns the user's API keys (metadata only, never the secret).
func (a *APIKey) ListAPIKeys(user *db.User, limit, offset int) (*ListAPIKeysResult, error) {
	count, err := a.APIKeyRepository.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired API keys")
	} else if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired API keys deleted")
	}

	total, err := a.APIKeyRepository.CountByUserId(user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to count API keys")
		return nil, err
	}

	keys, err := a.APIKeyRepository.ListByUserId(user.Id, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list API keys")
		return nil, err
	}

	list := make([]*APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		list = append(list, toAPIKeyResponse(k, ""))
	}

	return &ListAPIKeysResult{APIKeys: list, Total: total}, nil
}

// GetAPIKey returns a single API key owned by the user (never the secret).
func (a *APIKey) GetAPIKey(id db.Id, user *db.User) (*APIKeyResponse, error) {
	count, err := a.APIKeyRepository.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired API keys")
	} else if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired API keys deleted")
	}

	k, err := a.APIKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("apiKeyId", id.String()).
			Msg("Failed to get API key")
		return nil, err
	}
	if k == nil || k.UserId != user.Id {
		return nil, ErrAPIKeyNotFound
	}

	return toAPIKeyResponse(k, ""), nil
}

// DeleteAPIKey deletes a single API key owned by the user.
func (a *APIKey) DeleteAPIKey(id db.Id, user *db.User) error {
	k, err := a.APIKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("apiKeyId", id.String()).
			Msg("Failed to get API key")
		return err
	}
	if k == nil || k.UserId != user.Id {
		return ErrAPIKeyNotFound
	}

	err = a.APIKeyRepository.Delete(k.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("apiKeyId", k.Id.String()).
			Msg("Failed to delete API key")
		return err
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("apiKeyId", k.Id.String()).
		Msg("API key deleted")

	return nil
}
