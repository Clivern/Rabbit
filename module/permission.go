// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Workspace permission constants. Use with Perm.Can for workspace-scoped checks.
const (
	CanGetWorkspace    = "CAN_GET_WORKSPACE"
	CanUpdateWorkspace = "CAN_UPDATE_WORKSPACE"
	CanDeleteWorkspace = "CAN_DELETE_WORKSPACE"

	CanListWorkspaceMembers = "CAN_LIST_WORKSPACE_MEMBERS"
	CanUpdateMemberRole     = "CAN_UPDATE_MEMBER_ROLE"
	CanRemoveMember         = "CAN_REMOVE_MEMBER"

	CanListWorkspaceInvites = "CAN_LIST_WORKSPACE_INVITES"
	CanGetWorkspaceInvite   = "CAN_GET_WORKSPACE_INVITE"
	CanInviteMember         = "CAN_INVITE_MEMBER"
	CanRemoveInvite         = "CAN_REMOVE_INVITE"

	CanGetWorkspaceBilling    = "CAN_GET_WORKSPACE_BILLING"
	CanUpdateWorkspaceBilling = "CAN_UPDATE_WORKSPACE_BILLING"

	CanListAccessKeys  = "CAN_LIST_WORKSPACE_ACCESS_KEYS"
	CanCreateAccessKey = "CAN_CREATE_WORKSPACE_ACCESS_KEY"
	CanGetAccessKey    = "CAN_GET_WORKSPACE_ACCESS_KEY"
	CanDeleteAccessKey = "CAN_DELETE_WORKSPACE_ACCESS_KEY"

	CanListAgents  = "CAN_LIST_AGENTS"
	CanGetAgent    = "CAN_GET_AGENT"
	CanUpsertAgent = "CAN_UPSERT_AGENT"
	CanDeleteAgent = "CAN_DELETE_AGENT"

	CanListInstances    = "CAN_LIST_INSTANCES"
	CanGetInstance      = "CAN_GET_INSTANCE"
	CanRegisterInstance = "CAN_REGISTER_INSTANCE"
	CanReportCheck      = "CAN_REPORT_CHECK"

	CanListIntentions  = "CAN_LIST_INTENTIONS"
	CanGetIntention    = "CAN_GET_INTENTION"
	CanCreateIntention = "CAN_CREATE_INTENTION"
	CanUpdateIntention = "CAN_UPDATE_INTENTION"
	CanDeleteIntention = "CAN_DELETE_INTENTION"
	CanCheckIntention  = "CAN_CHECK_INTENTION"

	CanListIntegrations  = "CAN_LIST_INTEGRATIONS"
	CanGetIntegration    = "CAN_GET_INTEGRATION"
	CanCreateIntegration = "CAN_CREATE_INTEGRATION"
	CanUpdateIntegration = "CAN_UPDATE_INTEGRATION"
	CanDeleteIntegration = "CAN_DELETE_INTEGRATION"

	CanListEvents = "CAN_LIST_EVENTS"
	CanGetEvent   = "CAN_GET_EVENT"

	CanListKeyValue   = "CAN_LIST_KEY_VALUE"
	CanGetKeyValue    = "CAN_GET_KEY_VALUE"
	CanPutKeyValue    = "CAN_PUT_KEY_VALUE"
	CanDeleteKeyValue = "CAN_DELETE_KEY_VALUE"

	CanListTraffic = "CAN_LIST_TRAFFIC"
	CanGetTraffic  = "CAN_GET_TRAFFIC"

	CanUseChat = "CAN_USE_CHAT"
)

var (
	ErrPermissionPrincipalMissing = errors.New("permission principal missing")
	ErrPermissionWorkspaceMissing = errors.New("permission workspace missing")
)

// AccessKeyPermissions lists permissions that may be assigned to workspace access keys.
var AccessKeyPermissions = map[string]bool{
	CanGetWorkspace:    true,
	CanUpdateWorkspace: true,
	CanDeleteWorkspace: true,

	CanListWorkspaceMembers: true,
	CanUpdateMemberRole:     true,
	CanRemoveMember:         true,

	// CanInviteMember is left out: an invite needs a user to send it
	CanListWorkspaceInvites: true,
	CanGetWorkspaceInvite:   true,
	CanRemoveInvite:         true,

	CanGetWorkspaceBilling:    true,
	CanUpdateWorkspaceBilling: true,

	CanListAccessKeys:  true,
	CanCreateAccessKey: true,
	CanGetAccessKey:    true,
	CanDeleteAccessKey: true,

	CanListAgents:  true,
	CanGetAgent:    true,
	CanUpsertAgent: true,
	CanDeleteAgent: true,

	CanListInstances:    true,
	CanGetInstance:      true,
	CanRegisterInstance: true,
	CanReportCheck:      true,

	CanListIntentions:  true,
	CanGetIntention:    true,
	CanCreateIntention: true,
	CanUpdateIntention: true,
	CanDeleteIntention: true,
	CanCheckIntention:  true,

	CanListIntegrations:  true,
	CanGetIntegration:    true,
	CanCreateIntegration: true,
	CanUpdateIntegration: true,
	CanDeleteIntegration: true,

	CanListEvents: true,
	CanGetEvent:   true,

	CanListKeyValue:   true,
	CanGetKeyValue:    true,
	CanPutKeyValue:    true,
	CanDeleteKeyValue: true,

	CanListTraffic: true,
	CanGetTraffic:  true,
}

// Perm checks workspace permissions for users and workspace access keys.
type Perm struct {
	WorkspaceRepository db.WorkspaceRepository
	user                *db.User
	accessKey           *db.AccessKey
	workspace           *db.Workspace
}

// NewPerm creates a permission checker backed by the given database connection.
func NewPerm(dbConn *sql.DB) *Perm {
	return &Perm{
		WorkspaceRepository: db.NewWorkspaceRepository(dbConn),
	}
}

// WithUser sets the user principal. Mutually exclusive with WithAccessKey.
func (p *Perm) WithUser(user *db.User) *Perm {
	p.user = user
	p.accessKey = nil

	return p
}

// WithAccessKey sets the workspace access key principal. Mutually exclusive with WithUser.
func (p *Perm) WithAccessKey(accessKey *db.AccessKey) *Perm {
	p.accessKey = accessKey
	p.user = nil

	return p
}

// WithWorkspace sets the workspace being accessed.
func (p *Perm) WithWorkspace(workspace *db.Workspace) *Perm {
	p.workspace = workspace

	return p
}

// WithWorkspaceId loads the workspace by id from the repository.
func (p *Perm) WithWorkspaceId(workspaceId db.Id) *Perm {
	workspace, _ := p.WorkspaceRepository.GetById(workspaceId)
	p.workspace = workspace

	return p
}

// Can reports whether the configured principal may perform the given permission on the workspace.
func (p *Perm) Can(permission string) (bool, error) {
	if p.workspace == nil {
		log.Error().
			Str("permission", permission).
			Msg("Permission check without a workspace")
		return false, ErrPermissionWorkspaceMissing
	}

	var allowed bool
	var err error
	switch {
	case p.user != nil:
		allowed, err = p.CanAsUser(permission)
	case p.accessKey != nil:
		allowed, err = p.CanAsAccessKey(permission)
	default:
		log.Error().
			Str("workspaceId", p.workspace.Id.String()).
			Str("permission", permission).
			Msg("Permission check without a principal")
		return false, ErrPermissionPrincipalMissing
	}

	if err == nil && !allowed {
		log.Debug().
			Str("workspaceId", p.workspace.Id.String()).
			Str("permission", permission).
			Msg("Permission denied")
	}

	return allowed, err
}

// CanAsUser checks if the user has the given permission.
func (p *Perm) CanAsUser(permission string) (bool, error) {
	membership, err := p.WorkspaceRepository.GetWorkspaceMembership(p.workspace.Id, p.user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", p.workspace.Id.String()).
			Str("userId", p.user.Id.String()).
			Msg("Failed to get workspace membership")
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case CanGetWorkspace,
		CanGetWorkspaceBilling,
		CanListAgents,
		CanGetAgent,
		CanListInstances,
		CanGetInstance,
		CanListIntentions,
		CanGetIntention,
		CanCheckIntention,
		CanListKeyValue,
		CanGetKeyValue,
		CanListTraffic,
		CanGetTraffic,
		CanUseChat:
		return true, nil
	case CanUpdateWorkspace,
		CanDeleteWorkspace,
		CanUpdateWorkspaceBilling,
		CanListWorkspaceMembers,
		CanUpdateMemberRole,
		CanRemoveMember,
		CanListWorkspaceInvites,
		CanGetWorkspaceInvite,
		CanInviteMember,
		CanRemoveInvite,
		CanListAccessKeys,
		CanCreateAccessKey,
		CanGetAccessKey,
		CanDeleteAccessKey,
		CanUpsertAgent,
		CanDeleteAgent,
		CanRegisterInstance,
		CanReportCheck,
		CanCreateIntention,
		CanUpdateIntention,
		CanDeleteIntention,
		CanListIntegrations,
		CanGetIntegration,
		CanCreateIntegration,
		CanUpdateIntegration,
		CanDeleteIntegration,
		CanListEvents,
		CanGetEvent,
		CanPutKeyValue,
		CanDeleteKeyValue:
		return membership.Role == db.UserRoleAdmin || membership.Role == db.UserRoleOwner, nil
	default:
		return false, nil
	}
}

// CanAsAccessKey checks if the access key has the given permission.
func (p *Perm) CanAsAccessKey(permission string) (bool, error) {
	if p.accessKey.WorkspaceId != p.workspace.Id {
		return false, nil
	}
	if !AccessKeyPermissions[permission] {
		return false, nil
	}

	var meta AccessKeyMeta
	err := json.Unmarshal([]byte(*p.accessKey.Meta), &meta)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", p.accessKey.Id.String()).
			Msg("Failed to decode access key meta")
		return false, err
	}

	return lo.Contains(meta.Permissions, permission), nil
}

// ValidAccessKeyPermissions reports whether every permission may be assigned to an access key.
func ValidAccessKeyPermissions(permissions []string) bool {
	if len(permissions) == 0 {
		return len(AccessKeyPermissions) == 0
	}

	return lo.EveryBy(permissions, func(permission string) bool {
		return AccessKeyPermissions[permission]
	})
}
