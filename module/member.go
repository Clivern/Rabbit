// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
)

// WorkspaceMemberResponse is a workspace member shaped for API responses.
type WorkspaceMemberResponse struct {
	Id          db.Id  `json:"id"`
	WorkspaceId db.Id  `json:"workspaceId"`
	UserId      db.Id  `json:"userId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// UpdateWorkspaceMemberRoleRequest is the body for updating a workspace member role.
type UpdateWorkspaceMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
}

// ListWorkspaceMembersResult is what you get back when listing workspace members.
type ListWorkspaceMembersResult struct {
	Members []*WorkspaceMemberResponse
	Total   int64
}

// toWorkspaceMemberResponse shapes a membership and its user for API responses.
func toWorkspaceMemberResponse(member *db.WorkspaceUser, user *db.User) *WorkspaceMemberResponse {
	return &WorkspaceMemberResponse{
		Id:          member.Id,
		WorkspaceId: member.WorkspaceId,
		UserId:      member.UserId,
		Name:        user.Name,
		Email:       user.Email,
		Role:        member.Role,
		CreatedAt:   member.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   member.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListWorkspaceMembers returns members in a workspace with pagination.
func (w *Workspace) ListWorkspaceMembers(workspaceId db.Id, limit, offset int) (*ListWorkspaceMembersResult, error) {
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

	list, err := w.WorkspaceUserRepository.ListByWorkspaceId(
		workspaceId,
		limit,
		offset,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list workspace members")
		return nil, err
	}

	total, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	members := make([]*WorkspaceMemberResponse, 0, len(list))
	for _, member := range list {
		user, err := w.UserRepository.GetById(member.UserId)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("userId", member.UserId.String()).
				Msg("Failed to get member user")
			return nil, err
		}
		if user == nil {
			log.Warn().
				Str("workspaceId", workspaceId.String()).
				Str("userId", member.UserId.String()).
				Msg("Workspace member has no user")
			continue
		}

		members = append(members, toWorkspaceMemberResponse(member, user))
	}

	return &ListWorkspaceMembersResult{Members: members, Total: total}, nil
}

// GetWorkspaceMember returns the membership for a user in a workspace, or nil if not a member. TODO: remove if not used in future.
func (w *Workspace) GetWorkspaceMember(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	member, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get workspace membership")
		return nil, err
	}

	return member, nil
}

// UpdateWorkspaceMemberRole updates the role of a member in a workspace.
func (w *Workspace) UpdateWorkspaceMemberRole(workspaceId, userId db.Id, role string) (*WorkspaceMemberResponse, error) {
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

	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get workspace membership")
		return nil, err
	}
	if m == nil {
		return nil, ErrWorkspaceUserNotFound
	}
	if m.Role == db.UserRoleOwner {
		return nil, ErrWorkspaceOwnerProtected
	}

	m.Role = role

	err = w.WorkspaceUserRepository.Update(m)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Str("role", role).
			Msg("Failed to update member role")
		return nil, err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("userId", userId.String()).
		Str("role", role).
		Msg("Workspace member role updated")

	user, err := w.UserRepository.GetById(m.UserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get member user")
		return nil, err
	}
	if user == nil {
		return nil, ErrWorkspaceUserNotFound
	}

	return toWorkspaceMemberResponse(m, user), nil
}

// DeleteWorkspaceMember removes a member from a workspace.
func (w *Workspace) DeleteWorkspaceMember(workspaceId, userId db.Id) error {
	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get workspace membership")
		return err
	}
	if m == nil {
		return ErrWorkspaceUserNotFound
	}
	if m.Role == db.UserRoleOwner {
		return ErrWorkspaceOwnerProtected
	}

	err = w.WorkspaceUserRepository.Delete(m.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to remove workspace member")
		return err
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("userId", userId.String()).
		Msg("Workspace member removed")

	return nil
}
