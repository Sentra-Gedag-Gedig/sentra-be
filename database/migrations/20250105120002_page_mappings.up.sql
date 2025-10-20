CREATE TABLE IF NOT EXISTS page_mappings (
    page_id VARCHAR(100) PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    keywords JSONB NOT NULL,
    synonyms JSONB,
    category VARCHAR(100),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_page_mappings_category ON page_mappings(category);
CREATE INDEX idx_page_mappings_is_active ON page_mappings(is_active);

