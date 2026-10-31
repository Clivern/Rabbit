// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/pkg/encrypt"
	"github.com/clivern/sloth/pkg/lb"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	// AgentsTimeout bounds how long an instance may take to start answering.
	AgentsTimeout = 60 * time.Second
	// MaxSkillCandidates caps the agents considered when a message is addressed by skill.
	MaxSkillCandidates = 1000

	VerbSendMessage = "SendMessage"
	VerbGetTask     = "GetTask"
)

var (
	ErrCallDenied            = errors.New("call denied")
	ErrNoHealthyInstance     = errors.New("no healthy instance")
	ErrTaskNotFound          = errors.New("task not found")
	ErrUnknownSourceInstance = errors.New("unknown source instance")
)

// AgentsTrafficMeta is the JSON stored in ai_traffic.meta for a forwarded call.
type AgentsTrafficMeta struct {
	AgentId    db.Id `json:"agentId"`
	InstanceId db.Id `json:"instanceId"`
}

// SendReply is the part of a message:send reply that names a task; a direct message has none.
type SendReply struct {
	Task struct {
		Id string `json:"id"`
	} `json:"task"`
}

// Agents is the module that proxies messages between agents and the tasks they start.
type Agents struct {
	AgentRepository      db.AgentRepository
	InstanceRepository   db.AgentInstanceRepository
	IntentionRepository  db.IntentionRepository
	ProtectionRepository db.AgentProtectionRepository
	TrafficRepository    db.TrafficRepository
	TaskRepository       db.AgentTaskRepository
	SkillRepository      db.AgentSkillRepository
	Picker               *lb.Picker
	transport            http.RoundTripper
}

// NewAgents creates an agents module with the given repositories.
func NewAgents(
	agents db.AgentRepository,
	instances db.AgentInstanceRepository,
	intentions db.IntentionRepository,
	protections db.AgentProtectionRepository,
	traffic db.TrafficRepository,
	tasks db.AgentTaskRepository,
	skills db.AgentSkillRepository,
) *Agents {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = AgentsTimeout

	return &Agents{
		AgentRepository:      agents,
		InstanceRepository:   instances,
		IntentionRepository:  intentions,
		ProtectionRepository: protections,
		TrafficRepository:    traffic,
		TaskRepository:       tasks,
		SkillRepository:      skills,
		Picker:               lb.New(),
		transport:            transport,
	}
}

// SendMessage returns a proxy that sends caller's message to a named agent; only intentions without a skill apply.
func (a *Agents) SendMessage(workspaceId db.Id, caller *db.Agent, agentName string, sourceInstanceId string) (*httputil.ReverseProxy, error) {
	resolved, err := a.ResolveSourceInstance(caller, sourceInstanceId)
	if err != nil {
		return nil, err
	}

	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, agentName)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	call := &db.TrafficCall{
		WorkspaceId:      workspaceId,
		SourceAgent:      caller.Name,
		SourceInstance:   resolved,
		DestinationAgent: agent.Name,
		Verb:             VerbSendMessage,
	}

	allowed, err := a.IsAllowed(workspaceId, caller.Name, agent.Name, "")
	if err != nil {
		return nil, err
	}
	if !allowed {
		call.Decision = db.TrafficDecisionDeny
		call.Result = db.TrafficResultDenied
		_ = a.TrafficRepository.Create(call)
		return nil, ErrCallDenied
	}

	instances, err := a.InstanceRepository.ListLiveByAgentId(agent.Id, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		call.Decision = db.TrafficDecisionAllow
		call.Result = db.TrafficResultNoInstance
		_ = a.TrafficRepository.Create(call)
		return nil, ErrNoHealthyInstance
	}

	// Take turns across the agent's live instances.
	// TODO: support other lb strategies (random, least outstanding tasks, sticky), configurable per agent, instead of always round robin.
	picked, err := a.Picker.Pick(lo.Map(instances, func(instance *db.AgentInstance, _ int) string {
		return instance.Id.String()
	}), lb.Options{Strategy: lb.RoundRobin, Key: agent.Id.String()})
	if err != nil {
		return nil, err
	}

	return a.Forward(call, agent, instances[picked], "/message:send")
}

// SendSkillMessage returns a proxy that sends caller's message to a random live agent it may call for a skill.
func (a *Agents) SendSkillMessage(workspaceId db.Id, caller *db.Agent, skill string, sourceInstanceId string) (*httputil.ReverseProxy, error) {
	resolved, err := a.ResolveSourceInstance(caller, sourceInstanceId)
	if err != nil {
		return nil, err
	}

	offered, err := a.SkillRepository.GetSkill(workspaceId, skill)
	if err != nil {
		return nil, err
	}
	if offered == nil {
		return nil, ErrSkillNotFound
	}

	call := &db.TrafficCall{
		WorkspaceId:    workspaceId,
		SourceAgent:    caller.Name,
		SourceInstance: resolved,
		Skill:          skill,
		Verb:           VerbSendMessage,
	}

	// The same query as the catalog, so the caller can reach exactly the agents it was shown.
	callable, _, err := a.SkillRepository.ListCallableAgents(workspaceId, skill, &db.ListCallableAgentsFilter{
		Caller: caller.Name,
		Limit:  MaxSkillCandidates,
	})
	if err != nil {
		return nil, err
	}
	if len(callable) == 0 {
		call.Decision = db.TrafficDecisionDeny
		call.Result = db.TrafficResultDenied
		_ = a.TrafficRepository.Create(call)
		return nil, ErrCallDenied
	}

	live := lo.Filter(callable, func(agent *db.CallableAgent, _ int) bool { return agent.PassingInstances > 0 })
	if len(live) == 0 {
		call.Decision = db.TrafficDecisionAllow
		call.Result = db.TrafficResultNoInstance
		_ = a.TrafficRepository.Create(call)
		return nil, ErrNoHealthyInstance
	}

	agent, err := a.AgentRepository.GetByWorkspaceAndName(workspaceId, lo.Sample(live).Name)
	if err != nil {
		return nil, err
	}

	call.DestinationAgent = agent.Name

	instances, err := a.InstanceRepository.ListLiveByAgentId(agent.Id, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		call.Decision = db.TrafficDecisionAllow
		call.Result = db.TrafficResultNoInstance
		_ = a.TrafficRepository.Create(call)
		return nil, ErrNoHealthyInstance
	}

	// Take turns across the agent's live instances.
	// TODO: support other lb strategies (random, least outstanding tasks, sticky), configurable per agent, instead of always round robin.
	picked, err := a.Picker.Pick(lo.Map(instances, func(instance *db.AgentInstance, _ int) string {
		return instance.Id.String()
	}), lb.Options{Strategy: lb.RoundRobin, Key: agent.Id.String()})
	if err != nil {
		return nil, err
	}

	return a.Forward(call, agent, instances[picked], "/message:send")
}

// GetTask returns a proxy that fetches a task caller started from the instance that accepted it.
func (a *Agents) GetTask(workspaceId db.Id, caller *db.Agent, taskId string, sourceInstanceId string) (*httputil.ReverseProxy, error) {
	resolved, err := a.ResolveSourceInstance(caller, sourceInstanceId)
	if err != nil {
		return nil, err
	}

	// Tasks are looked up by their caller, so another agent's task is not found.
	task, err := a.TaskRepository.GetBySourceAndTaskId(workspaceId, caller.Name, taskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	agent, err := a.AgentRepository.GetById(task.AgentId)
	if err != nil {
		return nil, err
	}

	call := &db.TrafficCall{
		WorkspaceId:      workspaceId,
		SourceAgent:      caller.Name,
		SourceInstance:   resolved,
		DestinationAgent: agent.Name,
		Skill:            task.Skill,
		Verb:             VerbGetTask,
	}

	// Revoking the intention also stops the caller from reading tasks it already started.
	allowed, err := a.IsAllowed(workspaceId, caller.Name, agent.Name, task.Skill)
	if err != nil {
		return nil, err
	}
	if !allowed {
		call.Decision = db.TrafficDecisionDeny
		call.Result = db.TrafficResultDenied
		_ = a.TrafficRepository.Create(call)
		return nil, ErrCallDenied
	}

	// The task lives only on the instance that accepted it.
	if task.InstanceId == nil {
		call.Decision = db.TrafficDecisionAllow
		call.Result = db.TrafficResultNoInstance
		_ = a.TrafficRepository.Create(call)
		return nil, ErrNoHealthyInstance
	}

	instance, err := a.InstanceRepository.GetById(*task.InstanceId)
	if err != nil {
		return nil, err
	}

	return a.Forward(call, agent, instance, fmt.Sprintf("/tasks/%s", taskId))
}

// Forward returns a reverse proxy to one instance at path that records the call and, for message:send, where a returned task lives.
func (a *Agents) Forward(call *db.TrafficCall, agent *db.Agent, instance *db.AgentInstance, path string) (*httputil.ReverseProxy, error) {
	// Protected agents only accept calls that carry the credential Sloth holds for them.
	protection, err := a.ProtectionRepository.GetByAgentId(agent.Id)
	if err != nil {
		return nil, err
	}

	var config ProtectionConfig
	var secret string
	if protection != nil && protection.IsActive {
		_ = json.Unmarshal([]byte(protection.Config), &config)

		secret, err = encrypt.Decrypt(config.SecretEncrypted)
		if err != nil {
			return nil, err
		}
	}

	meta, _ := json.Marshal(AgentsTrafficMeta{
		AgentId:    agent.Id,
		InstanceId: instance.Id,
	})
	call.DestinationInstance = instance.InstanceId
	call.Meta = lo.ToPtr(string(meta))
	call.Decision = db.TrafficDecisionAllow

	start := time.Now()

	return &httputil.ReverseProxy{
		Transport: a.transport,
		Rewrite: func(pr *httputil.ProxyRequest) {
			// The instance registered how to reach it: scheme://address:port/path.
			pr.Out.URL.Scheme = instance.Scheme
			pr.Out.URL.Host = net.JoinHostPort(instance.Address, strconv.Itoa(instance.Port))
			pr.Out.URL.Path = instance.Path + path
			pr.Out.URL.RawPath = ""
			pr.Out.Host = pr.Out.URL.Host
			pr.SetXForwarded()

			// The caller's credentials prove who it is to Sloth and never reach the destination.
			pr.Out.Header.Del("X-Source-Agent")
			pr.Out.Header.Del("X-Source-Instance")
			pr.Out.Header.Del("X-API-Key")
			pr.Out.Header.Del("Authorization")

			if secret != "" {
				switch protection.Type {
				case db.AgentProtectionTypeAPIKey:
					pr.Out.Header.Set(config.Header, secret)
				case db.AgentProtectionTypeBasicAuth:
					pr.Out.SetBasicAuth(config.Username, secret)
				}
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			success := resp.StatusCode >= 200 && resp.StatusCode < 300

			// Only message:send can start a task, so only it reads the reply; GetTask streams through.
			if success && call.Verb == VerbSendMessage {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				resp.Body = io.NopCloser(bytes.NewReader(body))

				var reply SendReply
				_ = json.Unmarshal(body, &reply)

				// The reply still reaches the caller, so a lost record only stops it from polling the task.
				if reply.Task.Id != "" {
					err = a.TaskRepository.Save(&db.AgentTask{
						WorkspaceId: call.WorkspaceId,
						TaskId:      reply.Task.Id,
						SourceAgent: call.SourceAgent,
						AgentId:     agent.Id,
						InstanceId:  &instance.Id,
						Skill:       call.Skill,
					})
					if err != nil {
						log.Error().
							Err(err).
							Str("workspaceId", call.WorkspaceId.String()).
							Str("taskId", reply.Task.Id).
							Msg("Failed to record agent task")
					}
				}
			}

			call.LatencyMs = int(time.Since(start).Milliseconds())
			call.Result = lo.Ternary(success, db.TrafficResultCompleted, db.TrafficResultFailed)
			_ = a.TrafficRepository.Create(call)

			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			var timeout interface{ Timeout() bool }
			call.LatencyMs = int(time.Since(start).Milliseconds())
			call.Result = lo.Ternary(
				errors.As(err, &timeout) && timeout.Timeout(),
				db.TrafficResultTimeout,
				db.TrafficResultFailed,
			)
			_ = a.TrafficRepository.Create(call)

			util.WriteJSON(w, http.StatusBadGateway, map[string]any{
				"errorMessage": locale.TR(r, "agent_unreachable"),
			})
		},
	}, nil
}

// IsAllowed reports whether an intention lets source call destination for skill; no intention means deny.
func (a *Agents) IsAllowed(workspaceId db.Id, source, destination, skill string) (bool, error) {
	intention, err := a.IntentionRepository.Match(
		workspaceId,
		source,
		destination,
		skill,
	)
	if err != nil {
		return false, err
	}

	return intention != nil && intention.Action == db.IntentionActionAllow, nil
}

// ResolveSourceInstance returns the calling instance id when the caller named one of its own instances.
func (a *Agents) ResolveSourceInstance(caller *db.Agent, instanceId string) (string, error) {
	if instanceId == "" {
		return "", nil
	}

	instance, err := a.InstanceRepository.GetByAgentAndInstanceId(caller.Id, instanceId)
	if err != nil {
		return "", err
	}
	if instance == nil {
		return "", ErrUnknownSourceInstance
	}

	return instance.InstanceId, nil
}
