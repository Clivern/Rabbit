// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrWorkspaceUserNotFound   = errors.New("workspace user not found")
	ErrWorkspaceOwnerProtected = errors.New("workspace owner cannot be changed or removed")
	ErrWorkspaceAccessDenied   = errors.New("workspace access denied")
	ErrWorkspaceCheckAccess    = errors.New("failed check workspace access")
)

// Workspace is the module for workspace CRUD and members.
type Workspace struct {
	WorkspaceRepository     db.WorkspaceRepository
	WorkspaceUserRepository db.WorkspaceUserRepository
	SubscriptionRepository  db.SubscriptionRepository
	UserRepository          db.UserRepository
}

// NewWorkspace creates a workspace module with the given repositories.
func NewWorkspace(
	workspaces db.WorkspaceRepository,
	workspaceUsers db.WorkspaceUserRepository,
	subscriptions db.SubscriptionRepository,
	users db.UserRepository,
) *Workspace {
	return &Workspace{
		WorkspaceRepository:     workspaces,
		WorkspaceUserRepository: workspaceUsers,
		SubscriptionRepository:  subscriptions,
		UserRepository:          users,
	}
}

// toWorkspaceResponse shapes a workspace row for API responses.
func toWorkspaceResponse(item *db.Workspace, count int64, role string) *WorkspaceResponse {
	return &WorkspaceResponse{
		Id:           item.Id,
		Name:         item.Name,
		Handle:       item.Handle,
		MembersCount: count,
		Role:         role,
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// workspaceRole returns the member role, or empty when there is no membership.
func workspaceRole(membership *db.WorkspaceUser) string {
	return lo.TernaryF(
		membership != nil,
		func() string { return membership.Role },
		func() string { return "" },
	)
}

// CreateWorkspaceRequest is the body for creating a workspace.
type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=3,max=60" label:"Name"`
}

// UpdateWorkspaceRequest is the body for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=3,max=60" label:"Name"`
}

// WorkspaceResponse is a workspace shaped for API responses.
type WorkspaceResponse struct {
	Id           db.Id  `json:"id"`
	Name         string `json:"name"`
	Handle       string `json:"handle"`
	MembersCount int64  `json:"count"`
	Role         string `json:"role"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ListWorkspacesResponse is returned when listing workspace responses.
type ListWorkspacesResponse struct {
	Workspaces []*WorkspaceResponse
	Total      int64
}

// CreateWorkspace creates a workspace and adds the user as owner.
func (w *Workspace) CreateWorkspace(req *CreateWorkspaceRequest, user *db.User) (*WorkspaceResponse, error) {
	var handle string
	for {
		h, err := util.RandomHandle(4, 8)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Msg("Failed to generate workspace handle")
			return nil, err
		}

		existing, err := w.WorkspaceRepository.GetByHandle(h)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("handle", h).
				Msg("Failed to get workspace by handle")
			return nil, err
		}
		if existing == nil {
			handle = h
			break
		}
	}

	workspace := &db.Workspace{
		Name:   req.Name,
		Handle: handle,
	}

	err := w.WorkspaceRepository.Create(workspace)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create workspace")
		return nil, err
	}

	err = w.WorkspaceUserRepository.Create(&db.WorkspaceUser{
		WorkspaceId: workspace.Id,
		UserId:      user.Id,
		Role:        db.UserRoleOwner,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("workspaceId", workspace.Id.String()).
			Msg("Failed to add workspace owner")
		return nil, err
	}

	err = w.SubscriptionRepository.Create(&db.Subscription{
		WorkspaceId: workspace.Id,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("workspaceId", workspace.Id.String()).
			Msg("Failed to create workspace subscription")
		return nil, err
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspace.Id.String()).
		Str("handle", workspace.Handle).
		Msg("Workspace created")

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspace.Id.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	return toWorkspaceResponse(workspace, count, db.UserRoleOwner), nil
}

// ListWorkspaces retrieves workspace responses visible to the user.
func (w *Workspace) ListWorkspaces(user *db.User, limit, offset int) (*ListWorkspacesResponse, error) {
	workspaces, err := w.WorkspaceRepository.List(limit, offset, user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list workspaces")
		return nil, err
	}

	total, err := w.WorkspaceRepository.Count(user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to count workspaces")
		return nil, err
	}

	resp := make([]*WorkspaceResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		membership, err := w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", workspace.Id.String()).
				Msg("Failed to get workspace membership")
			return nil, err
		}

		count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspace.Id.String()).
				Msg("Failed to count workspace members")
			return nil, err
		}

		resp = append(resp, toWorkspaceResponse(workspace, count, workspaceRole(membership)))
	}

	return &ListWorkspacesResponse{Workspaces: resp, Total: total}, nil
}

// GetWorkspace retrieves a workspace after checking the user's access.
func (w *Workspace) GetWorkspace(workspaceId db.Id, user *db.User) (*WorkspaceResponse, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
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

	// An access key has no user, so it gets no role
	var membership *db.WorkspaceUser
	if user != nil {
		membership, err = w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", workspace.Id.String()).
				Msg("Failed to get workspace membership")
			return nil, err
		}
	}

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspace.Id.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	return toWorkspaceResponse(workspace, count, workspaceRole(membership)), nil
}

// UpdateWorkspace updates a workspace after checking the user's access.
func (w *Workspace) UpdateWorkspace(workspaceId db.Id, req *UpdateWorkspaceRequest, user *db.User) (*WorkspaceResponse, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
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

	workspace.Name = req.Name
	err = w.WorkspaceRepository.Update(workspace)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to update workspace")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Msg("Workspace updated")

	// An access key has no user, so it gets no role
	var membership *db.WorkspaceUser
	if user != nil {
		membership, err = w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", workspace.Id.String()).
				Msg("Failed to get workspace membership")
			return nil, err
		}
	}

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspace.Id.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	return toWorkspaceResponse(workspace, count, workspaceRole(membership)), nil
}

// DeleteWorkspace deletes a workspace.
func (w *Workspace) DeleteWorkspace(workspaceId db.Id) error {
	err := w.WorkspaceRepository.Delete(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to delete workspace")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Msg("Workspace deleted")

	return nil
}
