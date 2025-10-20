CREATE TABLE IF NOT EXISTS voice_sessions (
    id VARCHAR(26) PRIMARY KEY,
    user_id VARCHAR(26) NOT NULL,
    pending_confirmation BOOLEAN DEFAULT false,
    pending_page_id VARCHAR(100),
    context JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    last_activity TIMESTAMP NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_voice_sessions_user_id ON voice_sessions(user_id);
CREATE INDEX idx_voice_sessions_last_activity ON voice_sessions(last_activity DESC);