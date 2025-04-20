CREATE TABLE IF NOT EXISTS budget_transactions (
                         id VARCHAR(26) PRIMARY KEY,
                         user_id VARCHAR(26) NOT NULL,
                         title VARCHAR(255) NOT NULL,
                         description TEXT,
                         nominal DECIMAL(20, 2) NOT NULL,
                         type VARCHAR(255) NOT NULL,
                         category VARCHAR(255) NOT NULL,
                         audio_link VARCHAR(255),
                         created_at TIMESTAMP,
                         updated_at TIMESTAMP
)