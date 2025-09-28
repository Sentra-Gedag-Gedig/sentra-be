package entity

import (
	"time"
)

type VoiceCommand struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	AudioFile  string                 `json:"audio_file"`
	Transcript string                 `json:"transcript"`
	Command    string                 `json:"command"`
	Response   string                 `json:"response"`
	AudioURL   string                 `json:"audio_url"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type VoiceSession struct {
	ID                  string                 `json:"id"`
	UserID              string                 `json:"user_id"`
	PendingConfirmation bool                   `json:"pending_confirmation"`
	PendingPageID       string                 `json:"pending_page_id"`
	Context             map[string]interface{} `json:"context"`
	CreatedAt           time.Time              `json:"created_at"`
	LastActivity        time.Time              `json:"last_activity"`
}

type PageMapping struct {
	PageID      string    `json:"page_id"`
	URL         string    `json:"url"`
	DisplayName string    `json:"display_name"`
	Keywords    []string  `json:"keywords"`
	Synonyms    []string  `json:"synonyms"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

