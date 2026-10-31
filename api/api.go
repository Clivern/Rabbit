// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/github"
	"github.com/clivern/sloth/pkg/google"
	"github.com/clivern/sloth/pkg/resend"

	"github.com/spf13/viper"
)

// Instance is the process-wide API wired at server startup.
var Instance *API

// API holds shared application modules for HTTP handlers.
type API struct {
	Auth        *module.Auth
	Profile     *module.Profile
	Setup       *module.Setup
	Settings    *module.Settings
	APIKey      *module.APIKey
	Me          *module.Me
	Workspace   *module.Workspace
	Invite      *module.Invite
	Access      *module.Access
	Billing     *module.Billing
	Stats       *module.Stats
	Agent       *module.Agent
	Instance    *module.Instance
	Check       *module.Check
	Identity    *module.Identity
	Protection  *module.Protection
	Intention   *module.Intention
	Catalog     *module.Catalog
	Agents      *module.Agents
	Integration *module.Integration
	Event       *module.Event
	KeyValue    *module.KeyValue
	Traffic     *module.Traffic
	Chat        *module.Chat

	GitHubOAuth *github.OAuth
	GoogleOAuth *google.OAuth

	WorkspaceUsers db.WorkspaceUserRepository
	Usage          db.UsageRepository
}

// New wires repositories and modules once for the process.
func New() *API {
	conn := db.GetDB()

	users := db.NewUserRepository(conn)
	sessions := db.NewSessionRepository(conn)
	configs := db.NewConfigRepository(conn)
	workspaces := db.NewWorkspaceRepository(conn)
	workspaceUsers := db.NewWorkspaceUserRepository(conn)
	invites := db.NewUserInviteRepository(conn)
	apiKeys := db.NewAPIKeyRepository(conn)
	accessKeys := db.NewAccessKeyRepository(conn)
	subscriptions := db.NewSubscriptionRepository(conn)
	purchases := db.NewTokenPurchaseRepository(conn)
	stats := db.NewWorkspaceStatsRepository(conn)
	agents := db.NewAgentRepository(conn)
	skills := db.NewAgentSkillRepository(conn)
	instances := db.NewAgentInstanceRepository(conn)
	healthChecks := db.NewHealthCheckRepository(conn)
	agentChecks := db.NewAgentCheckRepository(conn)
	identities := db.NewAgentIdentityRepository(conn)
	identityTokens := db.NewAgentIdentityTokenRepository(conn)
	protections := db.NewAgentProtectionRepository(conn)
	intentions := db.NewIntentionRepository(conn)
	integrations := db.NewIntegrationRepository(conn)
	events := db.NewEventRepository(conn)
	kv := db.NewWorkspaceKeyValueRepository(conn)
	traffic := db.NewTrafficRepository(conn)
	usage := db.NewUsageRepository(conn)

	a := &API{
		Auth:        module.NewAuth(users, sessions, configs),
		Profile:     module.NewProfile(users),
		Setup:       module.NewSetup(configs, users),
		Settings:    module.NewSettings(configs),
		APIKey:      module.NewAPIKey(apiKeys),
		Me:          module.NewMe(apiKeys, users, accessKeys),
		Workspace:   module.NewWorkspace(workspaces, workspaceUsers, subscriptions, users),
		Invite:      module.NewInvite(invites, users, configs, workspaces, workspaceUsers, resend.NewMailer()),
		Access:      module.NewAccess(accessKeys, workspaces),
		Billing:     module.NewBilling(workspaces, subscriptions, purchases, module.Usage{}),
		Stats:       module.NewStats(workspaces, stats),
		Agent:       module.NewAgent(agents, instances, healthChecks, workspaces),
		Instance:    module.NewInstance(agents, instances, healthChecks, agentChecks, workspaces),
		Check:       module.NewCheck(agents, agentChecks, instances, healthChecks),
		Identity:    module.NewIdentity(agents, identities, identityTokens),
		Protection:  module.NewProtection(agents, protections),
		Intention:   module.NewIntention(intentions, workspaces),
		Catalog:     module.NewCatalog(skills),
		Agents:      module.NewAgents(agents, instances, intentions, protections, traffic, db.NewAgentTaskRepository(conn), skills),
		Integration: module.NewIntegration(integrations),
		Event:       module.NewEvent(events),
		KeyValue:    module.NewKeyValue(kv, workspaces),
		Traffic:     module.NewTraffic(traffic, workspaces),
		Chat: module.NewChat(
			db.NewChatRepository(conn),
			workspaces,
			module.NewChatTools(
				module.NewWorkspace(workspaces, workspaceUsers, subscriptions, users),
				module.NewAgent(agents, instances, healthChecks, workspaces),
				module.NewInstance(agents, instances, healthChecks, agentChecks, workspaces),
				module.NewIntention(intentions, workspaces),
				module.NewTraffic(traffic, workspaces),
				module.NewKeyValue(kv, workspaces),
				module.NewEvent(events),
				module.NewIntegration(integrations),
			),
		),
		GitHubOAuth: github.NewOAuth(github.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.github.client_id"),
			ClientSecret: viper.GetString("app.oauth.github.client_secret"),
			RedirectURL:  viper.GetString("app.oauth.github.redirect_url"),
			Scopes:       []string{"read:user", "user:email"},
			AllowSignup:  true,
		}),
		GoogleOAuth: google.NewOAuth(google.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.google.client_id"),
			ClientSecret: viper.GetString("app.oauth.google.client_secret"),
			RedirectURL:  viper.GetString("app.oauth.google.redirect_url"),
			Scopes:       []string{"openid", "email", "profile"},
		}),
		WorkspaceUsers: workspaceUsers,
		Usage:          usage,
	}

	Instance = a

	return a
}
