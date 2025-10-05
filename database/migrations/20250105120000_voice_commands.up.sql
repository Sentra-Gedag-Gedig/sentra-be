CREATE TABLE IF NOT EXISTS voice_commands (
    id VARCHAR(26) PRIMARY KEY,
    user_id VARCHAR(26) NOT NULL,
    audio_file VARCHAR(500),
    transcript TEXT NOT NULL,
    command VARCHAR(100) NOT NULL,
    response TEXT NOT NULL,
    audio_url VARCHAR(500),
    confidence DECIMAL(5, 4),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_voice_commands_user_id ON voice_commands(user_id);
CREATE INDEX idx_voice_commands_created_at ON voice_commands(created_at DESC);
CREATE INDEX idx_voice_commands_command ON voice_commands(command);