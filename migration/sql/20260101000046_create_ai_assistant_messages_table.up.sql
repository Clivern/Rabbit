CREATE TABLE IF NOT EXISTS ai_assistant_messages (
	id UUID PRIMARY KEY,
	chat_id UUID NOT NULL REFERENCES ai_assistant_chats(id) ON DELETE CASCADE,
	kind VARCHAR(20) NOT NULL,
	content TEXT NOT NULL,
	meta JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_assistant_messages_chat_id_created_at
	ON ai_assistant_messages (chat_id, created_at);
