CREATE TABLE IF NOT EXISTS ai_assistant_chats (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title VARCHAR(120) NOT NULL,
	status VARCHAR(20) NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_assistant_chats_workspace_user_updated_at
	ON ai_assistant_chats (workspace_id, user_id, updated_at DESC);
