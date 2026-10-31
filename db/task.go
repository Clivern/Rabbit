// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// AgentTask is a single row in the ai_agents_tasks table: where a task an agent accepted lives.
type AgentTask struct {
	Id          Id
	WorkspaceId Id
	TaskId      string
	SourceAgent string
	AgentId     Id
	InstanceId  *Id
	Skill       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AgentTaskRepository is the interface for task routing records.
type AgentTaskRepository interface {
	Save(task *AgentTask) error
	GetBySourceAndTaskId(workspaceId Id, sourceAgent, taskId string) (*AgentTask, error)
}

type AgentTaskRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentTaskRepository returns the repository for the ai_agents_tasks table.
func NewAgentTaskRepository(db *sql.DB) AgentTaskRepository {
	return &AgentTaskRepositoryPostgres{db: db}
}

// scanAgentTask maps a row into an agent task. Column order must match the select list.
func scanAgentTask(task *AgentTask) []any {
	return []any{
		&task.Id,
		&task.WorkspaceId,
		&task.TaskId,
		&task.SourceAgent,
		&task.AgentId,
		&task.InstanceId,
		&task.Skill,
		&task.CreatedAt,
		&task.UpdatedAt,
	}
}

// Save records where a task lives, moving it when the caller reuses a task id.
func (r *AgentTaskRepositoryPostgres) Save(task *AgentTask) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	return r.db.QueryRow(
		`INSERT INTO ai_agents_tasks (id, workspace_id, task_id, source_agent, agent_id, instance_id, skill)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (workspace_id, source_agent, task_id) DO UPDATE
		SET agent_id = EXCLUDED.agent_id,
			instance_id = EXCLUDED.instance_id,
			skill = EXCLUDED.skill,
			updated_at = $8
		RETURNING id, created_at, updated_at`,
		id.String(),
		task.WorkspaceId.String(),
		task.TaskId,
		task.SourceAgent,
		task.AgentId.String(),
		task.InstanceId,
		task.Skill,
		time.Now().UTC(),
	).Scan(&task.Id, &task.CreatedAt, &task.UpdatedAt)
}

// GetBySourceAndTaskId returns the task a caller started, or nil when there is none.
func (r *AgentTaskRepositoryPostgres) GetBySourceAndTaskId(workspaceId Id, sourceAgent, taskId string) (*AgentTask, error) {
	return QueryOne(
		r.db,
		`SELECT id, workspace_id, task_id, source_agent, agent_id, instance_id, skill, created_at, updated_at
		FROM ai_agents_tasks
		WHERE workspace_id = $1 AND source_agent = $2 AND task_id = $3`,
		scanAgentTask,
		workspaceId.String(),
		sourceAgent,
		taskId,
	)
}
