// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"
	"strings"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seenRequest is what the fake agent instance received.
type seenRequest struct {
	Method         string
	Path           string
	Query          string
	Version        string
	SourceAgent    string
	SourceInstance string
	APIKey         string
	AgentKey       string
	Body           string
}

func TestUnitAgents(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}
	message := `{"message":{"role":"ROLE_USER","parts":[{"data":{"user_id":"u_123"}}]}}`

	// newInstance starts a fake agent instance that answers with status and reply and records each request.
	newInstance := func(t *testing.T, status int, reply string) (*agentsFixture, *[]seenRequest) {
		var seen []seenRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			seen = append(seen, seenRequest{
				Method:         r.Method,
				Path:           r.URL.Path,
				Query:          r.URL.RawQuery,
				Version:        r.Header.Get("A2A-Version"),
				SourceAgent:    r.Header.Get("X-Source-Agent"),
				SourceInstance: r.Header.Get("X-Source-Instance"),
				APIKey:         r.Header.Get("X-API-Key"),
				AgentKey:       r.Header.Get("X-Agent-Key"),
				Body:           string(body),
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			w.Write([]byte(reply))
		}))
		t.Cleanup(server.Close)

		host, port, _ := net.SplitHostPort(server.Listener.Addr().String())
		portNumber, _ := strconv.Atoi(port)

		f := newAgentsFixture(host, portNumber)

		return f, &seen
	}

	// serve plays the caller's request through a proxy the way the handler does.
	serve := func(proxy *httputil.ReverseProxy, method, target, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, target, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("A2A-Version", "1.0")
		request.Header.Set("X-Source-Agent", "support-assistant")
		request.Header.Set("X-Source-Instance", "assistant-1")
		request.Header.Set("X-API-Key", "caller-secret")

		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, request)

		return w
	}

	sendPath := "/api/v1/workspaces/ws-1/agents/user-directory/message:send"

	t.Run("send to an agent proxies to the advertised path and records the task", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{"task":{"id":"task-7f3a","status":{"state":"TASK_STATE_WORKING"}}}`)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		w := serve(proxy, http.MethodPost, sendPath, message)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"task":{"id":"task-7f3a","status":{"state":"TASK_STATE_WORKING"}}}`, w.Body.String())

		require.Len(t, *seen, 1)
		assert.Equal(t, http.MethodPost, (*seen)[0].Method)
		assert.Equal(t, "/a2a/v1/message:send", (*seen)[0].Path)
		assert.Equal(t, "1.0", (*seen)[0].Version)
		assert.JSONEq(t, message, (*seen)[0].Body)

		require.Len(t, f.tasks.saved, 1)
		assert.Equal(t, "task-7f3a", f.tasks.saved[0].TaskId)
		assert.Equal(t, "support-assistant", f.tasks.saved[0].SourceAgent)
		assert.Equal(t, db.Id("instance-1"), *f.tasks.saved[0].InstanceId)

		require.Len(t, f.traffic.list, 1)
		assert.Equal(t, db.TrafficResultCompleted, f.traffic.list[0].Result)
		assert.Equal(t, VerbSendMessage, f.traffic.list[0].Verb)
		assert.Equal(t, "user-directory", f.traffic.list[0].DestinationAgent)
		assert.Equal(t, "directory-1", f.traffic.list[0].DestinationInstance)
	})

	t.Run("send to an agent takes turns across its live instances", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		first := f.instances.live[0]
		second := &db.AgentInstance{Id: "instance-2", AgentId: "agent-1", InstanceId: "directory-2", Address: first.Address, Port: first.Port, Scheme: "http", Path: "/a2a/v1"}
		f.instances.live = []*db.AgentInstance{first, second}

		for range 4 {
			proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
			require.NoError(t, err)
			serve(proxy, http.MethodPost, sendPath, message)
		}

		require.Len(t, f.traffic.list, 4)
		assert.Equal(t, "directory-1", f.traffic.list[0].DestinationInstance)
		assert.Equal(t, "directory-2", f.traffic.list[1].DestinationInstance)
		assert.Equal(t, "directory-1", f.traffic.list[2].DestinationInstance)
		assert.Equal(t, "directory-2", f.traffic.list[3].DestinationInstance)
	})

	t.Run("the caller's credentials never reach the destination", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{}`)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		serve(proxy, http.MethodPost, sendPath, message)
		assert.Empty(t, (*seen)[0].SourceAgent)
		assert.Empty(t, (*seen)[0].APIKey)
	})

	t.Run("a direct message reply records no task", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{"message":{"messageId":"m-1","role":"ROLE_AGENT","parts":[]}}`)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		serve(proxy, http.MethodPost, sendPath, message)
		assert.Empty(t, f.tasks.saved)
	})

	t.Run("an instance error is relayed as is and recorded as failed", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusBadRequest, `{"error":"bad message"}`)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		w := serve(proxy, http.MethodPost, sendPath, message)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"bad message"}`, w.Body.String())
		assert.Equal(t, db.TrafficResultFailed, f.traffic.list[0].Result)
		assert.Empty(t, f.tasks.saved)
	})

	t.Run("send to a protected agent adds its credential", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{}`)
		secret, err := encrypt.Encrypt("agent-secret")
		require.NoError(t, err)
		config, _ := json.Marshal(ProtectionConfig{Header: "X-Agent-Key", SecretEncrypted: secret})
		f.protections.protection = &db.AgentProtection{Type: db.AgentProtectionTypeAPIKey, Config: string(config), IsActive: true}

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		serve(proxy, http.MethodPost, sendPath, message)
		assert.Equal(t, "agent-secret", (*seen)[0].AgentKey)
	})

	t.Run("send to an agent is denied without an allow intention", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.intentions.match = nil

		_, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		assert.ErrorIs(t, err, ErrCallDenied)
		require.Len(t, f.traffic.list, 1)
		assert.Equal(t, db.TrafficDecisionDeny, f.traffic.list[0].Decision)
		assert.Equal(t, "user-directory", f.traffic.list[0].DestinationAgent)
		assert.Empty(t, f.traffic.list[0].DestinationInstance)
	})

	t.Run("send to an unknown agent is not found", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.agents.agent = nil

		_, err := f.module.SendMessage("ws-1", caller, "ghost", "")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("send to an agent without a live instance fails", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.instances.live = nil

		_, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
		assert.Equal(t, db.TrafficResultNoInstance, f.traffic.list[0].Result)
	})

	t.Run("an instance that cannot be reached answers 502", func(t *testing.T) {
		f := newAgentsFixture("127.0.0.1", 1)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "")
		require.NoError(t, err)

		w := serve(proxy, http.MethodPost, sendPath, message)
		assert.Equal(t, http.StatusBadGateway, w.Code)
		assert.Equal(t, db.TrafficResultFailed, f.traffic.list[0].Result)
	})

	t.Run("send to a skill nobody offers is not found", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)

		_, err := f.module.SendSkillMessage("ws-1", caller, "ghost.skill", "")
		assert.ErrorIs(t, err, ErrSkillNotFound)
	})

	t.Run("send to a skill the caller may not call is denied", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{}

		_, err := f.module.SendSkillMessage("ws-1", caller, "user.profile.get", "")
		assert.ErrorIs(t, err, ErrCallDenied)
	})

	t.Run("send to a skill whose agents have no live instance fails", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{{Name: "user-directory", Instances: 1}}

		_, err := f.module.SendSkillMessage("ws-1", caller, "user.profile.get", "")
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
	})

	t.Run("send to a skill proxies to a live agent and records the skill", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{"task":{"id":"task-1"}}`)
		f.skills.skill = &db.AgentSkill{SkillId: "user.profile.get"}
		f.skills.agents = []*db.CallableAgent{{Name: "user-directory", Instances: 1, PassingInstances: 1}}

		proxy, err := f.module.SendSkillMessage("ws-1", caller, "user.profile.get", "")
		require.NoError(t, err)

		serve(proxy, http.MethodPost, "/api/v1/workspaces/ws-1/agents/skills/user.profile.get/message:send", message)
		require.Len(t, *seen, 1)
		assert.Equal(t, "support-assistant", f.skills.agentFilter.Caller)
		assert.Equal(t, "user.profile.get", f.tasks.saved[0].Skill)
		assert.Equal(t, "user.profile.get", f.traffic.list[0].Skill)
	})

	t.Run("get task proxies to the instance that accepted it with the query", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{"id":"task-7f3a","status":{"state":"TASK_STATE_COMPLETED"}}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1", InstanceId: &f.instances.instance.Id}

		proxy, err := f.module.GetTask("ws-1", caller, "task-7f3a", "")
		require.NoError(t, err)

		w := serve(proxy, http.MethodGet, "/api/v1/workspaces/ws-1/agents/tasks/task-7f3a?historyLength=2", "")
		assert.Contains(t, w.Body.String(), "TASK_STATE_COMPLETED")
		require.Len(t, *seen, 1)
		assert.Equal(t, http.MethodGet, (*seen)[0].Method)
		assert.Equal(t, "/a2a/v1/tasks/task-7f3a", (*seen)[0].Path)
		assert.Equal(t, "historyLength=2", (*seen)[0].Query)
		assert.Equal(t, VerbGetTask, f.traffic.list[0].Verb)
	})

	t.Run("get task of another caller or an unknown task is not found", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)

		_, err := f.module.GetTask("ws-1", caller, "task-7f3a", "")
		assert.ErrorIs(t, err, ErrTaskNotFound)
	})

	t.Run("get task is denied once the intention is revoked", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1", InstanceId: &f.instances.instance.Id}
		f.intentions.match = &db.Intention{Action: db.IntentionActionDeny}

		_, err := f.module.GetTask("ws-1", caller, "task-7f3a", "")
		assert.ErrorIs(t, err, ErrCallDenied)
	})

	t.Run("get task whose instance is gone fails", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.tasks.task = &db.AgentTask{TaskId: "task-7f3a", SourceAgent: "support-assistant", AgentId: "agent-1"}

		_, err := f.module.GetTask("ws-1", caller, "task-7f3a", "")
		assert.ErrorIs(t, err, ErrNoHealthyInstance)
	})

	t.Run("send records the calling instance and strips its header", func(t *testing.T) {
		f, seen := newInstance(t, http.StatusOK, `{}`)

		proxy, err := f.module.SendMessage("ws-1", caller, "user-directory", "directory-1")
		require.NoError(t, err)

		w := serve(proxy, http.MethodPost, sendPath, message)
		assert.Equal(t, http.StatusOK, w.Code)

		require.Len(t, f.traffic.list, 1)
		assert.Equal(t, "support-assistant", f.traffic.list[0].SourceAgent)
		assert.Equal(t, "directory-1", f.traffic.list[0].SourceInstance)

		require.Len(t, *seen, 1)
		assert.Empty(t, (*seen)[0].SourceAgent)
		assert.Empty(t, (*seen)[0].SourceInstance)
	})

	t.Run("send rejects a source instance the caller does not own", func(t *testing.T) {
		f, _ := newInstance(t, http.StatusOK, `{}`)
		f.instances.instance = nil

		_, err := f.module.SendMessage("ws-1", caller, "user-directory", "ghost-9")
		assert.ErrorIs(t, err, ErrUnknownSourceInstance)
		assert.Empty(t, f.traffic.list)
	})
}
