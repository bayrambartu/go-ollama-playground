
CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_conversation
        FOREIGN KEY (conversation_id)
        REFERENCES conversations(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_message_role
        CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id
ON messages(conversation_id, id);
