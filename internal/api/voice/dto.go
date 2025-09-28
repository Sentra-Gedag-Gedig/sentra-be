package voice

import (
	"mime/multipart"
	"time"
)

type ProcessVoiceRequest struct {
	AudioFile *multipart.FileHeader `json:"audio_file" validate:"required"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

type VoiceResponse struct {
	Text         string                 `json:"text"`
	AudioURL     string                 `json:"audio_url,omitempty"`
	Action       string                 `json:"action"`
	Target       string                 `json:"target,omitempty"`
	Success      bool                   `json:"success"`
	Confidence   float64                `json:"confidence,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	SessionState *SessionState          `json:"session_state,omitempty"`
}

type SessionState struct {
	PendingConfirmation bool                   `json:"pending_confirmation"`
	PendingPageID       string                 `json:"pending_page_id,omitempty"`
	Context             map[string]interface{} `json:"context,omitempty"`
}

type VoiceCommandHistory struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	AudioFile   string                 `json:"audio_file"`
	Transcript  string                 `json:"transcript"`
	Command     string                 `json:"command"`
	Response    string                 `json:"response"`
	AudioURL    string                 `json:"audio_url"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type VoiceAnalytics struct {
	TotalCommands     int                    `json:"total_commands"`
	SuccessRate       float64                `json:"success_rate"`
	MostUsedCommands  map[string]int         `json:"most_used_commands"`
	ConfidenceStats   map[string]float64     `json:"confidence_stats"`
	UsageByTime       map[string]int         `json:"usage_by_time"`
	CategoryUsage     map[string]int         `json:"category_usage"`
}

type SmartSuggestion struct {
	Text        string `json:"text"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
}

type SuggestionsResponse struct {
	Suggestions []SmartSuggestion      `json:"suggestions"`
	Context     map[string]interface{} `json:"context"`
}

type NLPTestRequest struct {
	Text string `json:"text" validate:"required,min=1,max=500"`
}

type NLPTestResponse struct {
	Input      string           `json:"input"`
	Intent     string           `json:"intent"`
	Page       string           `json:"page"`
	Confidence float64          `json:"confidence"`
	Matches    []MatchResult    `json:"matches"`
	Processing ProcessingDetail `json:"processing"`
}

type MatchResult struct {
	Keyword string  `json:"keyword"`
	Score   float64 `json:"score"`
	Type    string  `json:"type"` // exact, synonym, fuzzy
}

type ProcessingDetail struct {
	CleanedText    string   `json:"cleaned_text"`
	ExtractedTokens []string `json:"extracted_tokens"`
	ProcessingTime string   `json:"processing_time"`
}

type PageMapping struct {
	PageID      string   `json:"page_id"`
	URL         string   `json:"url"`
	DisplayName string   `json:"display_name"`
	Keywords    []string `json:"keywords"`
	Synonyms    []string `json:"synonyms"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	IsActive    bool     `json:"is_active"`
}

type ConfirmationRequest struct {
	AudioFile     *multipart.FileHeader `json:"audio_file" validate:"required"`
	PendingPageID string                `json:"pending_page_id" validate:"required"`
	SessionID     string                `json:"session_id" validate:"required"`
}

 
