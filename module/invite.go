// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/sloth/conf"
	"github.com/clivern/sloth/db"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const (
	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRejected = "rejected"
	InviteStatusExpired  = "expired"
)

var (
	ErrInviteNotFound         = errors.New("invite not found")
	ErrPendingInviteExists    = errors.New("pending invite already exists")
	ErrUserAlreadyInWorkspace = errors.New("user already in workspace")
	ErrFailedCreateInvite     = errors.New("failed create invite")
	ErrFailedListInvites      = errors.New("failed list invites")
	ErrFailedGetInvite        = errors.New("failed get invite")
	ErrFailedDeleteInvite     = errors.New("failed delete invite")
	ErrFailedAttachInvites    = errors.New("failed attach invites")
)

// InviteMailer sends invite emails.
type InviteMailer interface {
	SendInviteEmail(to, inviteLink, pname string) error
}

// Invite is the module for workspace invite CRUD.
type Invite struct {
	UserInviteRepository    db.UserInviteRepository
	UserRepository          db.UserRepository
	ConfigRepository        db.ConfigRepository
	WorkspaceRepository     db.WorkspaceRepository
	WorkspaceUserRepository db.WorkspaceUserRepository
	Mailer                  InviteMailer
}

// toInviteResponse shapes an invite row for API responses. An empty token hides it.
func toInviteResponse(inv *db.UserInvite, token string) *InviteResponse {
	return &InviteResponse{
		Id:            inv.Id,
		Email:         inv.Email,
		Role:          inv.Role,
		Token:         token,
		Status:        inv.Status,
		InviterUserId: inv.InviterUserId,
		WorkspaceId:   inv.WorkspaceId,
		ExpiresAt:     inv.ExpiresAt.UTC().Format(time.RFC3339),
		CreatedAt:     inv.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     inv.UpdatedAt.UTC().Format(time.RFC3339),
		AcceptedAt: lo.TernaryF(
			inv.AcceptedAt != nil,
			func() *string { return new(inv.AcceptedAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}
}

// NewInvite creates an invite module with the given repositories.
func NewInvite(invites db.UserInviteRepository, users db.UserRepository, config db.ConfigRepository, workspaces db.WorkspaceRepository, workspaceUsers db.WorkspaceUserRepository, mailer InviteMailer) *Invite {
	return &Invite{
		UserInviteRepository:    invites,
		UserRepository:          users,
		ConfigRepository:        config,
		WorkspaceRepository:     workspaces,
		WorkspaceUserRepository: workspaceUsers,
		Mailer:                  mailer,
	}
}

// CreateInviteRequest is what you pass when creating an invite.
type CreateInviteRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=255" label:"Email"`
	Role  string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
}

// InviteResponse is an invite shaped for API responses.
type InviteResponse struct {
	Id            db.Id   `json:"id"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	Token         string  `json:"token,omitempty"`
	Status        string  `json:"status"`
	InviterUserId db.Id   `json:"inviterUserId"`
	WorkspaceId   db.Id   `json:"workspaceId"`
	WorkspaceName string  `json:"workspaceName,omitempty"`
	ExpiresAt     string  `json:"expiresAt"`
	AcceptedAt    *string `json:"acceptedAt,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
}

// ListInvitesResponse is returned when listing invites.
type ListInvitesResponse struct {
	Invites []*InviteResponse
	Total   int64
}

// CreateInvite creates a workspace invite and sends the login email.
func (i *Invite) CreateInvite(workspaceId db.Id, req *CreateInviteRequest, inviter *db.User) (*InviteResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
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

	existingUser, err := i.UserRepository.GetByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("email", req.Email).
			Msg("Failed to get user by email")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}
	if existingUser != nil {
		membership, err := i.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, existingUser.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("userId", existingUser.Id.String()).
				Msg("Failed to get workspace membership")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
		}
		if membership != nil {
			return nil, ErrUserAlreadyInWorkspace
		}
	}

	count, err := i.UserInviteRepository.CountPendingByEmailInWorkspace(workspaceId, req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("email", req.Email).
			Msg("Failed to count pending invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}
	if count > 0 {
		return nil, ErrPendingInviteExists
	}

	token, err := uuid.NewRandom()
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to generate invite token")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}

	expiresAt := time.Now().UTC().Add(conf.InviteExpiry)
	invite := &db.UserInvite{
		Email:         req.Email,
		Role:          req.Role,
		Token:         token.String(),
		Status:        InviteStatusPending,
		InviterUserId: inviter.Id,
		WorkspaceId:   workspaceId,
		ExpiresAt:     expiresAt,
	}

	err = i.UserInviteRepository.Create(invite)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("email", req.Email).
			Msg("Failed to create invite")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("inviteId", invite.Id.String()).
		Str("userId", inviter.Id.String()).
		Str("email", req.Email).
		Str("role", req.Role).
		Msg("Workspace invite created")

	purl := viper.GetString("app.url")
	pname := viper.GetString("app.name")

	err = i.Mailer.SendInviteEmail(req.Email, fmt.Sprintf("%s/login", purl), pname)
	if err != nil {
		log.Warn().
			Err(err).
			Str("userId", inviter.Id.String()).
			Str("email", req.Email).
			Msg("Failed to send invite email; invite was created")
	}

	if existingUser != nil {
		err = i.AttachPending(existingUser)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", existingUser.Id.String()).
				Str("email", req.Email).
				Msg("Failed to attach workspace invite")
		} else {
			invite.Status = InviteStatusAccepted
		}
	}

	return toInviteResponse(invite, invite.Token), nil
}

// ListInvites returns paginated invites for a workspace.
func (i *Invite) ListInvites(workspaceId db.Id, limit, offset int) (*ListInvitesResponse, error) {
	count, err := i.UserInviteRepository.MarkExpiredAsExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to mark expired invites")
	} else if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired invites marked")
	}

	total, err := i.UserInviteRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInvites, err)
	}

	invites, err := i.UserInviteRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInvites, err)
	}

	list := make([]*InviteResponse, 0, len(invites))
	for _, inv := range invites {
		list = append(list, toInviteResponse(inv, inv.Token))
	}

	return &ListInvitesResponse{Invites: list, Total: total}, nil
}

// GetInvite returns one invite by Id within a workspace.
func (i *Invite) GetInvite(workspaceId, inviteId db.Id) (*InviteResponse, error) {
	count, err := i.UserInviteRepository.MarkExpiredAsExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to mark expired invites")
	} else if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired invites marked")
	}

	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
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

	invite, err := i.UserInviteRepository.GetById(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to get invite")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInvite, err)
	}
	if invite == nil || invite.WorkspaceId != workspaceId {
		return nil, ErrInviteNotFound
	}

	return toInviteResponse(invite, ""), nil
}

// DeleteInvite deletes an invite by Id within a workspace.
func (i *Invite) DeleteInvite(workspaceId, inviteId db.Id) error {
	workspace, err := i.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace")
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	invite, err := i.UserInviteRepository.GetById(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to get invite")
		return fmt.Errorf("%w: %v", ErrFailedDeleteInvite, err)
	}
	if invite == nil || invite.WorkspaceId != workspaceId {
		return ErrInviteNotFound
	}

	err = i.UserInviteRepository.Delete(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to delete invite")
		return fmt.Errorf("%w: %v", ErrFailedDeleteInvite, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("inviteId", inviteId.String()).
		Msg("Workspace invite deleted")

	return nil
}

// AttachPending adds the user to every pending workspace invite for their email.
func (i *Invite) AttachPending(user *db.User) error {
	_, err := i.UserInviteRepository.MarkExpiredAsExpired()
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to mark expired invites")
		return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
	}

	invites, err := i.UserInviteRepository.ListPendingByEmail(user.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list pending invites")
		return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
	}

	now := time.Now().UTC()
	for _, invite := range invites {
		membership, err := i.WorkspaceUserRepository.GetByWorkspaceAndUser(invite.WorkspaceId, user.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", invite.WorkspaceId.String()).
				Msg("Failed to get workspace membership")
			return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
		}
		if membership == nil {
			err = i.WorkspaceUserRepository.Create(&db.WorkspaceUser{
				WorkspaceId: invite.WorkspaceId,
				UserId:      user.Id,
				Role:        invite.Role,
			})
			if err != nil {
				log.Error().
					Err(err).
					Str("userId", user.Id.String()).
					Str("workspaceId", invite.WorkspaceId.String()).
					Msg("Failed to add workspace member")
				return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
			}
		}

		err = i.UserInviteRepository.UpdateStatus(invite.Id, InviteStatusAccepted, &now)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("inviteId", invite.Id.String()).
				Msg("Failed to accept invite")
			return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
		}

		log.Info().
			Str("inviteId", invite.Id.String()).
			Str("workspaceId", invite.WorkspaceId.String()).
			Str("userId", user.Id.String()).
			Msg("Workspace invite attached")
	}

	return nil
}
