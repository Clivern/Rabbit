// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/clivern/sloth/api"
	"github.com/clivern/sloth/conf"
	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/middleware"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/util"

	"github.com/go-chi/chi/v5"
	cmid "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Setup creates and configures the HTTP server
func SetupServer(Static embed.FS, a *api.API) http.Handler {
	r := chi.NewRouter()

	r.Use(cmid.Recoverer)
	r.Use(cmid.Timeout(
		time.Duration(viper.GetInt("app.timeout")) * time.Second,
	))
	r.Use(middleware.ClientIP(viper.GetString("app.trusted_proxies")))
	r.Use(middleware.RequestId)
	r.Use(middleware.AppContext)
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestSizeLimit(
		int64(10 * 1024 * 1024),
	))
	r.Use(middleware.Auth())
	r.Use(middleware.ValidateParams)
	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) { // suppress browser favicon requests
		w.WriteHeader(http.StatusNoContent)
	})

	r.Group(func(r chi.Router) { // public — no auth required
		r.Get("/api/v1/public/_health", a.HealthAction)                                                              // liveness probe
		r.Get("/api/v1/public/_ready", a.ReadyAction)                                                                // readiness probe (db, deps)
		r.With(middleware.RateLimit(5, time.Minute)).Post("/api/v1/public/action/setup", a.SetupAction)              // initial app setup
		r.Get("/api/v1/public/action/setup/status", a.SetupStatusAction)                                             // setup completion status
		r.Post("/api/v1/public/action/logout", a.LogoutAction)                                                       // user logout
		r.Get("/api/v1/public/action/oauth/github", a.GitHubOAuthStartAction)                                        // start GitHub OAuth
		r.Get("/api/v1/public/action/oauth/github/callback", a.GitHubOAuthCallbackAction)                            // GitHub OAuth callback
		r.Get("/api/v1/public/action/oauth/google", a.GoogleOAuthStartAction)                                        // start Google OAuth
		r.Get("/api/v1/public/action/oauth/google/callback", a.GoogleOAuthCallbackAction)                            // Google OAuth callback
		r.With(middleware.RateLimit(10, time.Minute)).Post("/api/v1/public/oauth/token", a.IssueIdentityTokenAction) // client credentials token for an agent identity
		if conf.IsSaaS() {
			r.Post("/api/v1/public/action/stripe/webhook", a.StripeWebhookAction) // Stripe billing webhook
		}
	})
	r.Get("/api/v1/me", a.GetMeAction) // current authenticated user
	r.Group(func(r chi.Router) {       // user profile
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Get("/api/v1/action/profile", a.GetProfileAction)    // get user profile
		r.Put("/api/v1/action/profile", a.UpdateProfileAction) // update user profile
	})
	r.Group(func(r chi.Router) { // app settings — admin only
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin}}))
		r.Put("/api/v1/action/settings", a.UpdateSettingsAction) // update app settings
		r.Get("/api/v1/action/settings", a.GetSettingsAction)    // get app settings
	})
	r.Group(func(r chi.Router) { // user API keys
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Post("/api/v1/apiKeys", a.CreateUserAPIKeyAction)              // create user API key
		r.Get("/api/v1/apiKeys", a.ListUserAPIKeysAction)                // list user API keys
		r.Get("/api/v1/apiKeys/{apiKeyId}", a.GetUserAPIKeyAction)       // get user API key
		r.Delete("/api/v1/apiKeys/{apiKeyId}", a.DeleteUserAPIKeyAction) // delete user API key
	})
	r.Group(func(r chi.Router) { // workspaces
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Post("/api/v1/workspaces", a.CreateWorkspaceAction) // create workspace
		r.Get("/api/v1/workspaces", a.ListWorkspacesAction)   // list user workspaces
	})

	r.Route("/api/v1/workspaces/{workspaceId}", func(r chi.Router) { // workspace-scoped resources
		r.Use(middleware.Protect(middleware.Config{Workspace: true}))

		r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspace})).Get("/", a.GetWorkspaceAction)          // get workspace
		r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateWorkspace})).Put("/", a.UpdateWorkspaceAction)    // update workspace
		r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteWorkspace})).Delete("/", a.DeleteWorkspaceAction) // delete workspace

		r.Route("/invites", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanInviteMember})).Post("/", a.CreateInviteAction)             // create member invite
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListWorkspaceInvites})).Get("/", a.ListInvitesAction)       // list member invites
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspaceInvite})).Get("/{inviteId}", a.GetInviteAction) // get member invite
			r.With(middleware.Protect(middleware.Config{Perm: module.CanRemoveInvite})).Delete("/{inviteId}", a.DeleteInviteAction) // revoke member invite
		})

		r.Route("/members", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListWorkspaceMembers})).Get("/", a.ListWorkspaceMembersAction)                // list workspace members
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateMemberRole})).Put("/{memberUserId}", a.UpdateWorkspaceMemberRoleAction) // update member role
			r.With(middleware.Protect(middleware.Config{Perm: module.CanRemoveMember})).Delete("/{memberUserId}", a.DeleteWorkspaceMemberAction)      // remove member
		})

		r.Route("/keys", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanCreateAccessKey})).Post("/", a.CreateAccessKeyAction)          // create workspace access key
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListAccessKeys})).Get("/", a.ListAccessKeysAction)             // list workspace access keys
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetAccessKey})).Get("/{keyId}", a.GetAccessKeyAction)          // get workspace access key
			r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteAccessKey})).Delete("/{keyId}", a.DeleteAccessKeyAction) // delete workspace access key
		})

		if conf.IsSaaS() {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspaceBilling})).Get("/billing", a.GetBillingStatusAction)                   // get billing status
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetWorkspaceBilling})).Get("/billing/usage", a.GetBillingUsageAction)              // get billing usage
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateWorkspaceBilling})).Post("/billing/checkout", a.CreateBillingCheckoutAction) // start Stripe checkout
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateWorkspaceBilling})).Post("/billing/portal", a.CreateBillingPortalAction)     // open Stripe customer portal
		}

		r.Route("/agents", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListAgents})).Get("/", a.ListAgentsAction)                 // list agents
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Put("/{agentName}", a.UpsertAgentAction)    // register or replace an agent card
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetAgent})).Get("/{agentName}", a.GetAgentAction)          // get an agent
			r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteAgent})).Delete("/{agentName}", a.DeleteAgentAction) // delete an agent
			r.With(middleware.AgentAuth()).Get("/catalog/skills", a.ListCatalogSkillsAction)                                       // list the skills the calling agent may call
			r.With(middleware.AgentAuth()).Get("/catalog/skills/{skill}/agents", a.ListCatalogSkillAgentsAction)                   // list the agents the calling agent may call for a skill
			r.With(middleware.AgentAuth()).Post("/skills/{skill}/message:send", a.SendSkillMessageAction)                          // send an A2A message to an agent that offers a skill
			r.With(middleware.AgentAuth()).Get("/tasks/{taskId}", a.GetTaskAction)                                                 // get a task the calling agent started
			r.With(middleware.AgentAuth()).Post("/{agentName}/message:send", a.SendAgentMessageAction)                             // send an A2A message to a named agent

			r.Route("/{agentName}/instances", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{Perm: module.CanListInstances})).Get("/", a.ListInstancesAction)                              // list agent instances
				r.With(middleware.Protect(middleware.Config{Perm: module.CanGetInstance})).Get("/{instanceId}", a.GetInstanceAction)                      // get an agent instance
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Put("/{instanceId}", a.RegisterInstanceAction)            // register or update an instance
				r.With(middleware.Protect(middleware.Config{Perm: module.CanRegisterInstance})).Delete("/{instanceId}", a.DeregisterInstanceAction)       // deregister an instance
				r.With(middleware.Protect(middleware.Config{Perm: module.CanReportCheck})).Post("/{instanceId}/checks/{checkId}/pass", a.PassCheckAction) // report a ttl check passing
				r.With(middleware.Protect(middleware.Config{Perm: module.CanReportCheck})).Post("/{instanceId}/checks/{checkId}/warn", a.WarnCheckAction) // report a ttl check warning
				r.With(middleware.Protect(middleware.Config{Perm: module.CanReportCheck})).Post("/{instanceId}/checks/{checkId}/fail", a.FailCheckAction) // report a ttl check critical
			})

			r.Route("/{agentName}/checks", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{Perm: module.CanGetAgent})).Get("/", a.ListAgentChecksAction)                 // list health check templates
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Put("/{checkId}", a.UpsertAgentCheckAction)    // create or replace a health check template
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Delete("/{checkId}", a.DeleteAgentCheckAction) // delete a health check template
			})

			r.Route("/{agentName}/identities", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{Perm: module.CanGetAgent})).Get("/", a.ListIdentitiesAction)                   // list agent identities
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Post("/", a.CreateIdentityAction)               // issue an agent identity
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Put("/{identityId}", a.UpdateIdentityAction)    // rename or toggle an agent identity
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Delete("/{identityId}", a.DeleteIdentityAction) // revoke an agent identity
			})

			r.Route("/{agentName}/protection", func(r chi.Router) {
				r.With(middleware.Protect(middleware.Config{Perm: module.CanGetAgent})).Get("/", a.GetProtectionAction)                   // get the agent protection
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Put("/", a.SaveProtectionAction)               // set or replace the agent protection
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Put("/status", a.UpdateProtectionStatusAction) // turn the agent protection on or off
				r.With(middleware.Protect(middleware.Config{Perm: module.CanUpsertAgent})).Delete("/", a.DeleteProtectionAction)          // remove the agent protection
			})
		})

		r.Route("/intentions", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListIntentions})).Get("/", a.ListIntentionsAction)                   // list intentions
			r.With(middleware.Protect(middleware.Config{Perm: module.CanCreateIntention})).Put("/", a.CreateIntentionAction)                 // create an intention
			r.With(middleware.Protect(middleware.Config{Perm: module.CanCheckIntention})).Post("/check", a.CheckIntentionAction)             // check whether a call is allowed
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetIntention})).Get("/{intentionId}", a.GetIntentionAction)          // get an intention
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateIntention})).Put("/{intentionId}", a.UpdateIntentionAction)    // update an intention
			r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteIntention})).Delete("/{intentionId}", a.DeleteIntentionAction) // delete an intention
		})

		r.Route("/integrations", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListIntegrations})).Get("/", a.ListIntegrationsAction)                     // list event integrations
			r.With(middleware.Protect(middleware.Config{Perm: module.CanCreateIntegration})).Post("/", a.CreateIntegrationAction)                  // create an integration
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetIntegration})).Get("/{integrationId}", a.GetIntegrationAction)          // get an integration
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUpdateIntegration})).Put("/{integrationId}", a.UpdateIntegrationAction)    // update an integration
			r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteIntegration})).Delete("/{integrationId}", a.DeleteIntegrationAction) // delete an integration
		})

		r.Route("/events", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListEvents})).Get("/", a.ListEventsAction)      // list workspace events
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetEvent})).Get("/{eventId}", a.GetEventAction) // get a workspace event
		})

		r.Route("/kv", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListKeyValue})).Get("/", a.ListKeyValueAction)         // list workspace keys
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetKeyValue})).Get("/*", a.GetKeyValueAction)          // get a workspace key
			r.With(middleware.Protect(middleware.Config{Perm: module.CanPutKeyValue})).Put("/*", a.PutKeyValueAction)          // write a workspace key
			r.With(middleware.Protect(middleware.Config{Perm: module.CanDeleteKeyValue})).Delete("/*", a.DeleteKeyValueAction) // delete a workspace key
		})

		r.Route("/chats", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUseChat})).Get("/", a.ListChatsAction)                         // list the user's chats
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUseChat})).Post("/", a.CreateChatAction)                       // start a chat
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUseChat})).Get("/{chatId}", a.GetChatAction)                   // get a chat with its messages
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUseChat})).Delete("/{chatId}", a.DeleteChatAction)             // delete a chat with its messages
			r.With(middleware.Protect(middleware.Config{Perm: module.CanUseChat})).Post("/{chatId}/messages", a.SendChatMessageAction) // ask in a chat and stream the answer
		})

		r.Route("/traffic", func(r chi.Router) {
			r.With(middleware.Protect(middleware.Config{Perm: module.CanListTraffic})).Get("/", a.ListTrafficAction)          // list gateway calls
			r.With(middleware.Protect(middleware.Config{Perm: module.CanGetTraffic})).Get("/{trafficId}", a.GetTrafficAction) // get a gateway call
		})

	})

	r.With(middleware.BasicAuth(viper.GetString("app.metrics.username"), viper.GetString("app.metrics.secret"))).Get("/api/v1/public/_metrics", promhttp.Handler().ServeHTTP) // Prometheus metrics

	dist, err := fs.Sub(Static, "web/dist")
	if err != nil {
		panic(fmt.Sprintf(
			"Error while accessing dist files: %s",
			err.Error(),
		))
	}

	r.Handle("/assets/*", http.StripPrefix("/", http.FileServer(http.FS(dist)))) // static frontend assets
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {                    // SPA fallback — serve index.html
		if strings.HasPrefix(r.URL.Path, "/api/") { // an unknown API route is a JSON 404, never the dashboard
			util.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "route_not_found"),
			})
			return
		}

		indexFile, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})

	return r
}

// Run starts the HTTP server with graceful shutdown support
func RunServer(Static embed.FS) error {
	err := db.InitDB(ReadWriteDatabase(), ReadOnlyDatabase()...)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a := api.New()
	handler := SetupServer(Static, a)

	err = module.StartBus()
	if err != nil {
		return fmt.Errorf("failed to start nats bus: %w", err)
	}

	defer func() {
		err := db.CloseDB()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error closing database connection")
		}
	}()

	defer module.StopBus()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	sched := module.NewScheduler(
		db.NewHealthCheckRepository(db.GetDB()),
		db.NewAgentInstanceRepository(db.GetDB()),
		module.GetBus(),
	)

	go func() {
		defer close(done)
		sched.Run(ctx)
	}()

	defer func() {
		cancel()
		<-done
	}()

	timeout := time.Duration(viper.GetInt("app.timeout")) * time.Second

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		Handler:           handler,
		MaxHeaderBytes:    230 * 1024,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadTimeout:       timeout + 5*time.Second,
		WriteTimeout:      timeout + 5*time.Second,
	}

	serr := make(chan error, 1)

	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Str("edition", conf.Edition()).
			Msg("Starting HTTP server")

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		shutdownTimeout := 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down server")

		err := srv.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		log.Info().
			Msg("Server shutdown complete")
	}

	return nil
}
