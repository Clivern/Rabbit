CREATE TABLE IF NOT EXISTS ai_traffic (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	source_agent VARCHAR(100) NOT NULL,
	source_instance VARCHAR(100) NOT NULL,
	destination_agent VARCHAR(100) NOT NULL,
	destination_instance VARCHAR(100) NOT NULL,
	skill VARCHAR(120) NOT NULL,
	verb VARCHAR(60) NOT NULL,
	decision VARCHAR(20) NOT NULL,
	result VARCHAR(40) NOT NULL,
	latency_ms INTEGER NOT NULL,
	task_id VARCHAR(100),
	meta JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_traffic_workspace_id_created_at
	ON ai_traffic (workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_traffic_workspace_id_instances
	ON ai_traffic (workspace_id, source_agent, destination_instance);
CREATE INDEX IF NOT EXISTS idx_ai_traffic_workspace_id_task_id
	ON ai_traffic (workspace_id, task_id);
