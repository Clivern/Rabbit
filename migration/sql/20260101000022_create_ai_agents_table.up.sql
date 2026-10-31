CREATE TABLE IF NOT EXISTS ai_agents (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	name VARCHAR(100) NOT NULL,
	card JSONB,
	card_checksum VARCHAR(64),
	card_path VARCHAR(255) NOT NULL DEFAULT '/.well-known/agent-card.json',
	version VARCHAR(40) NOT NULL,
	lb_strategy VARCHAR(30) NOT NULL DEFAULT 'round_robin',
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (workspace_id, name),
	CONSTRAINT ai_agents_lb_strategy_check CHECK (lb_strategy IN ('random', 'round_robin', 'least_outstanding_tasks', 'sticky'))
);
