// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrAccessKeyNotFound           = errors.New("access key not found")
	ErrFailedCreateAccessKey       = errors.New("failed create access key")
	ErrFailedListAccessKeys        = errors.New("failed list access keys")
	ErrFailedGetAccessKey          = errors.New("failed get access key")
	ErrFailedDeleteAccessKey       = errors.New("failed delete access key")
	ErrInvalidAccessKeyPermissions = errors.New("invalid access key permissions")
	ErrAccessKeyPermissionDenied   = errors.New("access key permission denied")
)

// Access is the module for workspace access key CRUD.
type Access struct {
	AccessKeyRepository db.AccessKeyRepository
	WorkspaceRepository db.WorkspaceRepository
}

// NewAccess creates an access key module with the given repositories.
func NewAccess(accessKeys db.AccessKeyRepository, workspaces db.WorkspaceRepository) *Access {
	return &Access{
		AccessKeyRepository: accessKeys,
		WorkspaceRepository: workspaces,
	}
}

// toAccessKeyResponse shapes an access key row for API responses. An empty key hides it.
func toAccessKeyResponse(item *db.AccessKey, permissions []string, key string) *AccessKeyResponse {
	return &AccessKeyResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Name:        item.Name,
		Permissions: permissions,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
		Key:         key,
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}
}

// ListAccessKeysResult is what you get back when listing workspace access keys.
type ListAccessKeysResult struct {
	Keys  []*AccessKeyResponse
	Total int64
}

// CreateAccessKeyRequest is what you pass when creating a workspace access key.
type CreateAccessKeyRequest struct {
	Name        string   `json:"name" validate:"required,max=60" label:"Name"`
	ExpiresAt   string   `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,required" label:"Permissions"`
}

// AccessKeyResponse is a workspace access key shaped for API responses.
type AccessKeyResponse struct {
	Id          db.Id    `json:"id"`
	WorkspaceId db.Id    `json:"workspaceId"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expiresAt,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Key         string   `json:"key,omitempty"`
}

// AccessKeyMeta is the JSON shape stored in access_keys.meta.
type AccessKeyMeta struct {
	Permissions []string `json:"permissions"`
}

// CreateAccessKey creates an access key
func (a *Access) CreateAccessKey(workspaceId db.Id, principal *Principal, req *CreateAccessKeyRequest) (*AccessKeyResponse, error) {
	if !ValidAccessKeyPermissions(req.Permissions) {
		return nil, ErrInvalidAccessKeyPermissions
	}

	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
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

	perm := (&Perm{WorkspaceRepository: a.WorkspaceRepository}).WithWorkspace(workspace)
	if principal.IsUser() {
		perm = perm.WithUser(principal.User)
	} else {
		perm = perm.WithAccessKey(principal.AccessKey)
	}

	for _, permission := range req.Permissions {
		allowed, err := perm.Can(permission)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("permission", permission).
				Msg("Failed to check access key permission")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateAccessKey, err)
		}
		if !allowed {
			return nil, ErrAccessKeyPermissionDenied
		}
	}

	var expiresAt *time.Time
	if lo.IsNotEmpty(req.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, ErrInvalidExpiresAt
		}

		expiresAt = new(t)
	}

	raw, err := json.Marshal(AccessKeyMeta{Permissions: req.Permissions})
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to encode access key meta")
		return nil, fmt.Errorf("%w: %v", ErrInvalidAccessKeyPermissions, err)
	}

	token, err := uuid.NewRandom()
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to generate access key")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateAccessKey, err)
	}

	item := &db.AccessKey{
		WorkspaceId: workspaceId,
		Name:        req.Name,
		Key:         token.String(),
		ExpiresAt:   expiresAt,
		Meta:        new(string(raw)),
	}

	err = a.AccessKeyRepository.Create(item)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create access key")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateAccessKey, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("accessKeyId", item.Id.String()).
		Msg("Access key created")

	return toAccessKeyResponse(item, req.Permissions, token.String()), nil
}

// ListAccessKeys returns workspace access keys (metadata only, never the secret).
func (a *Access) ListAccessKeys(workspaceId db.Id, limit, offset int) (*ListAccessKeysResult, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
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

	if _, err := a.AccessKeyRepository.DeleteExpired(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired workspace access keys")
	}

	total, err := a.AccessKeyRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count access keys")
		return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
	}

	keys, err := a.AccessKeyRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list access keys")
		return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
	}

	list := make([]*AccessKeyResponse, 0, len(keys))
	for _, item := range keys {
		var meta AccessKeyMeta
		if err := json.Unmarshal([]byte(*item.Meta), &meta); err != nil {
			log.Error().
				Err(err).
				Str("accessKeyId", item.Id.String()).
				Msg("Failed to decode access key meta")
			return nil, fmt.Errorf("%w: %v", ErrFailedListAccessKeys, err)
		}

		list = append(list, toAccessKeyResponse(item, meta.Permissions, ""))
	}

	return &ListAccessKeysResult{Keys: list, Total: total}, nil
}

// GetAccessKey returns one access key (never the secret).
func (a *Access) GetAccessKey(workspaceId, id db.Id) (*AccessKeyResponse, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
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

	if _, err := a.AccessKeyRepository.DeleteExpired(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired workspace access keys")
	}

	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", id.String()).
			Msg("Failed to get access key")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAccessKey, err)
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return nil, ErrAccessKeyNotFound
	}

	var meta AccessKeyMeta
	if err := json.Unmarshal([]byte(*item.Meta), &meta); err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", item.Id.String()).
			Msg("Failed to decode access key meta")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAccessKey, err)
	}

	return toAccessKeyResponse(item, meta.Permissions, ""), nil
}

// DeleteAccessKey deletes a workspace access key.
func (a *Access) DeleteAccessKey(workspaceId, id db.Id) error {
	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", id.String()).
			Msg("Failed to get access key")
		return fmt.Errorf("%w: %v", ErrFailedDeleteAccessKey, err)
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return ErrAccessKeyNotFound
	}

	err = a.AccessKeyRepository.Delete(item.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", item.Id.String()).
			Msg("Failed to delete access key")
		return fmt.Errorf("%w: %v", ErrFailedDeleteAccessKey, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("accessKeyId", item.Id.String()).
		Msg("Access key deleted")

	return nil
}
