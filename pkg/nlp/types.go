package nlp

type IntentResult struct {
	Intent          string        `json:"intent"`
	Page            string        `json:"page"`
	PageURL         string        `json:"page_url"`
	PageDisplayName string        `json:"page_display_name"`
	PageDescription string        `json:"page_description"`
	Confidence      float64       `json:"confidence"`
	Matches         []MatchResult `json:"matches"`
	ProcessingTime  string        `json:"processing_time"`
}

type MatchResult struct {
	Keyword string  `json:"keyword"`
	Score   float64 `json:"score"`
	Type    string  `json:"type"`
}

type INLPProcessor interface {
	ProcessCommand(text string) (*IntentResult, error)
	GetPageMapping(pageID string) (PageMappingData, bool)
	GetAllMappings() map[string]PageMappingData
	AddPageMapping(pageID string, mapping PageMappingData) error
	GenerateResponseText(result *IntentResult) string
}

type PageMappingData struct {
	PageID      string   `json:"page_id"`
	URL         string   `json:"url"`
	DisplayName string   `json:"display_name"`
	Keywords    []string `json:"keywords"`
	Synonyms    []string `json:"synonyms"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
}