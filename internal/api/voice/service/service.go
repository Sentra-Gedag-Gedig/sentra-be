package voiceService

import (
	budgetService "ProjectGolang/internal/api/budget_manager/service"
	chatGPT "ProjectGolang/pkg/openai"
	"ProjectGolang/internal/api/voice"
	voiceRepository "ProjectGolang/internal/api/voice/repository"
	"ProjectGolang/pkg/nlp"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/utils"
	"context"

	"github.com/sirupsen/logrus"
)

type IVoiceService interface {
	
	ProcessChatCommand(ctx context.Context, userID string, req voice.ProcessVoiceRequest) (*voice.VoiceResponse, error)
	
	
	ProcessVoiceCommand(ctx context.Context, userID string, req voice.ProcessVoiceRequest) (*voice.VoiceResponse, error)
	ProcessConfirmation(ctx context.Context, userID string, req voice.ConfirmationRequest) (*voice.VoiceResponse, error)
	
	
	GetVoiceHistory(ctx context.Context, userID string, page, limit int) ([]voice.VoiceCommandHistory, int, error)
	GetVoiceAnalytics(ctx context.Context, userID string) (*voice.VoiceAnalytics, error)
	
	
	GetSmartSuggestions(ctx context.Context, userID string) (*voice.SuggestionsResponse, error)
	
	
	GetPageMappings(ctx context.Context) ([]voice.PageMapping, error)
	CreatePageMapping(ctx context.Context, mapping voice.PageMapping) error
	UpdatePageMapping(ctx context.Context, pageID string, mapping voice.PageMapping) error
	
	
	TestNLPProcessing(ctx context.Context, req voice.NLPTestRequest) (*voice.NLPTestResponse, error)
	ServeAudioFile(ctx context.Context, filename string) ([]byte, error)
}

type voiceService struct {
	log          *logrus.Logger
	voiceRepo    voiceRepository.Repository
	s3Client     s3.ItfS3
	utils        utils.IUtils
	nlpProcessor nlp.INLPProcessor
	config       *VoiceConfig
	budgetService budgetService.IBudgetService
	chatGPT chatGPT.IChatGPT
}

type VoiceConfig struct {
	ElevenLabsAPIKey  string   `json:"eleven_labs_api_key"`
	ElevenLabsVoiceID string   `json:"eleven_labs_voice_id"`
	OpenAIAPIKey      string   `json:"openai_api_key"`
	MaxFileSize       int64    `json:"max_file_size"`
	AllowedFormats    []string `json:"allowed_formats"`
	SessionTimeout    int64    `json:"session_timeout_hours"`
	RateLimitPerHour  int      `json:"rate_limit_per_hour"`
	EnableAnalytics   bool     `json:"enable_analytics"`
	AudioCacheTimeout int64    `json:"audio_cache_timeout_minutes"`
}

func NewVoiceService(
	log *logrus.Logger,
	voiceRepo voiceRepository.Repository,
	s3Client s3.ItfS3,
	utils utils.IUtils,
	nlpProcessor nlp.INLPProcessor,
	config *VoiceConfig,
	budgetService budgetService.IBudgetService,
	chatGPT chatGPT.IChatGPT,
) IVoiceService {
	return &voiceService{
		log:          log,
		voiceRepo:    voiceRepo,
		s3Client:     s3Client,
		utils:        utils,
		nlpProcessor: nlpProcessor,
		config:       config,
		budgetService: budgetService,
		chatGPT: chatGPT,
	}
}
