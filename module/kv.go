// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrKeyValueNotFound     = errors.New("kv not found")
	ErrInvalidKeyValueKey   = errors.New("invalid kv key")
	ErrFailedListKeyValue   = errors.New("failed list kv")
	ErrFailedGetKeyValue    = errors.New("failed get kv")
	ErrFailedPutKeyValue    = errors.New("failed put kv")
	ErrFailedDeleteKeyValue = errors.New("failed delete kv")
	KeyValueKeyPattern      = regexp.MustCompile(`^[A-Za-z0-9_./-]{1,200}$`)
)

// KeyValue is the module for workspace key/value config.
type KeyValue struct {
	KeyValueRepository  db.WorkspaceKeyValueRepository
	WorkspaceRepository db.WorkspaceRepository
}

// PutKeyValueRequest is the body for writing a workspace key.
type PutKeyValueRequest struct {
	Value     string `json:"value" validate:"required" label:"Value"`
	ExpiresAt string `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
}

// KeyValueResponse is a workspace key shaped for API responses.
type KeyValueResponse struct {
	Id          db.Id   `json:"id"`
	WorkspaceId db.Id   `json:"workspaceId"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// ListKeyValueResponse is returned when listing workspace keys.
type ListKeyValueResponse struct {
	Items []*KeyValueResponse
	Total int64
}

// NewKeyValue creates a workspace KeyValue module with the given repositories.
func NewKeyValue(items db.WorkspaceKeyValueRepository, workspaces db.WorkspaceRepository) *KeyValue {
	return &KeyValue{
		KeyValueRepository:  items,
		WorkspaceRepository: workspaces,
	}
}

// toKeyValueResponse shapes a workspace key for API responses.
func toKeyValueResponse(item *db.WorkspaceKeyValue) *KeyValueResponse {
	return &KeyValueResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Key:         item.Key,
		Value:       item.Value,
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListKeyValue returns non-expired keys that start with prefix.
func (k *KeyValue) ListKeyValue(workspaceId db.Id, prefix string) (*ListKeyValueResponse, error) {
	items, err := k.KeyValueRepository.ListByPrefix(workspaceId, strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(prefix))
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("prefix", prefix).
			Msg("Failed to list workspace keys")
		return nil, fmt.Errorf("%w: %v", ErrFailedListKeyValue, err)
	}

	list := make([]*KeyValueResponse, 0, len(items))
	for _, item := range items {
		list = append(list, toKeyValueResponse(item))
	}

	return &ListKeyValueResponse{Items: list, Total: int64(len(list))}, nil
}

// GetKeyValue returns one non-expired key.
func (k *KeyValue) GetKeyValue(workspaceId db.Id, key string) (*KeyValueResponse, error) {
	item, err := k.KeyValueRepository.Get(workspaceId, key)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("key", key).
			Msg("Failed to get workspace key")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetKeyValue, err)
	}
	if item == nil {
		return nil, ErrKeyValueNotFound
	}

	return toKeyValueResponse(item), nil
}

// PutKeyValue writes a workspace key.
func (k *KeyValue) PutKeyValue(workspaceId db.Id, key string, req *PutKeyValueRequest) (*KeyValueResponse, error) {
	valid := KeyValueKeyPattern.MatchString(key)
	if valid {
		for _, part := range strings.Split(key, "/") {
			if part == "" || part == "." || part == ".." {
				valid = false
				break
			}
		}
	}
	if !valid {
		return nil, ErrInvalidKeyValueKey
	}

	workspace, err := k.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace")
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	var expiresAt *time.Time
	if lo.IsNotEmpty(req.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, ErrInvalidExpiresAt
		}

		expiresAt = new(t)
	}

	item := &db.WorkspaceKeyValue{
		WorkspaceId: workspaceId,
		Key:         key,
		Value:       req.Value,
		ExpiresAt:   expiresAt,
	}
	err = k.KeyValueRepository.Upsert(item)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("key", key).
			Msg("Failed to write workspace key")
		return nil, fmt.Errorf("%w: %v", ErrFailedPutKeyValue, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("key", key).
		Msg("Workspace key written")

	return toKeyValueResponse(item), nil
}

// DeleteKeyValue removes a workspace key.
func (k *KeyValue) DeleteKeyValue(workspaceId db.Id, key string) error {
	item, err := k.KeyValueRepository.Get(workspaceId, key)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("key", key).
			Msg("Failed to get workspace key")
		return fmt.Errorf("%w: %v", ErrFailedDeleteKeyValue, err)
	}
	if item == nil {
		return ErrKeyValueNotFound
	}

	err = k.KeyValueRepository.Delete(workspaceId, key)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("key", key).
			Msg("Failed to delete workspace key")
		return fmt.Errorf("%w: %v", ErrFailedDeleteKeyValue, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("key", key).
		Msg("Workspace key deleted")

	return nil
}
