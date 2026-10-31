// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// SendAgentMessageAction proxies a message:send from the calling agent to a named agent.
func (a *API) SendAgentMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	agentName := chi.URLParam(r, "agentName")
	caller, _ := module.AgentFrom(r.Context())

	proxy, err := a.Agents.SendMessage(
		db.Id(wid),
		caller,
		agentName,
		r.Header.Get("X-Source-Instance"),
	)

	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrUnknownSourceInstance, http.StatusUnauthorized, locale.TR(r, "unknown_source_instance")},
			{module.ErrCallDenied, http.StatusForbidden, locale.TR(r, "a2a_call_denied")},
			{module.ErrNoHealthyInstance, http.StatusServiceUnavailable, locale.TR(r, "no_healthy_instance")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", caller.Name).
			Str("destination", agentName).
			Msg("Failed to send message to agent")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_a2a_request"),
		})

		return
	}

	proxy.ServeHTTP(w, r)
}

// SendSkillMessageAction proxies a message:send from the calling agent to an agent that offers a skill.
func (a *API) SendSkillMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	skill := chi.URLParam(r, "skill")
	caller, _ := module.AgentFrom(r.Context())

	proxy, err := a.Agents.SendSkillMessage(
		db.Id(wid),
		caller,
		skill,
		r.Header.Get("X-Source-Instance"),
	)

	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrSkillNotFound, http.StatusNotFound, locale.TR(r, "skill_not_found")},
			{module.ErrUnknownSourceInstance, http.StatusUnauthorized, locale.TR(r, "unknown_source_instance")},
			{module.ErrCallDenied, http.StatusForbidden, locale.TR(r, "a2a_call_denied")},
			{module.ErrNoHealthyInstance, http.StatusServiceUnavailable, locale.TR(r, "no_healthy_instance")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", caller.Name).
			Str("skill", skill).
			Msg("Failed to send message to skill")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_a2a_request"),
		})

		return
	}

	proxy.ServeHTTP(w, r)
}

// GetTaskAction proxies a GetTask for a task the calling agent started to the instance that accepted it.
func (a *API) GetTaskAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	taskId := chi.URLParam(r, "taskId")
	caller, _ := module.AgentFrom(r.Context())

	proxy, err := a.Agents.GetTask(
		db.Id(wid),
		caller,
		taskId,
		r.Header.Get("X-Source-Instance"),
	)

	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrTaskNotFound, http.StatusNotFound, locale.TR(r, "task_not_found")},
			{module.ErrUnknownSourceInstance, http.StatusUnauthorized, locale.TR(r, "unknown_source_instance")},
			{module.ErrCallDenied, http.StatusForbidden, locale.TR(r, "a2a_call_denied")},
			{module.ErrNoHealthyInstance, http.StatusServiceUnavailable, locale.TR(r, "no_healthy_instance")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", caller.Name).
			Str("taskId", taskId).
			Msg("Failed to get agent task")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_a2a_request"),
		})

		return
	}

	proxy.ServeHTTP(w, r)
}
