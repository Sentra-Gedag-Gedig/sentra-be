package voiceService

import (
	"ProjectGolang/internal/api/voice"
	voiceRepository "ProjectGolang/internal/api/voice/repository"
	"ProjectGolang/pkg/nlp"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/utils"
	"context"
	"github.com/sirupsen/logrus"
)

type IVoiceService interface {
	ProcessVoiceCommand(ctx context.Context, userID string, req voice.ProcessVoiceRequest) (*voice.VoiceResponse, error)
	ProcessConfirmation(ctx context.Context, userID string, req voice.ConfirmationRequest) (*voice.VoiceResponse, error)
	GetVoiceHistory(ctx context.Context, userID string, page, limit int) ([]voice.VoiceCommandHistory, int, error)
	GetVoiceAnalytics(ctx context.Context, userID string) (*voice.VoiceAnalytics, error)
	GetSmartSuggestions(ctx context.Context, userID string) (*voice.SuggestionsResponse, error)
	TestNLPProcessing(ctx context.Context, req voice.NLPTestRequest) (*voice.NLPTestResponse, error)
	GetPageMappings(ctx context.Context) ([]voice.PageMapping, error)
	CreatePageMapping(ctx context.Context, mapping voice.PageMapping) error
	UpdatePageMapping(ctx context.Context, pageID string, mapping voice.PageMapping) error
	ServeAudioFile(ctx context.Context, filename string) ([]byte, error)
}

type voiceService struct {
	log          *logrus.Logger
	voiceRepo    voiceRepository.Repository
	s3Client     s3.ItfS3
	utils        utils.IUtils
	nlpProcessor nlp.INLPProcessor
	config       *VoiceConfig
}

type VoiceConfig struct {
	ElevenLabsAPIKey  string   `json:"eleven_labs_api_key"`
	ElevenLabsVoiceID string   `json:"eleven_labs_voice_id"`
	OpenAIAPIKey      string   `json:"openai_api_key"`
	UploadPath        string   `json:"upload_path"`
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
) IVoiceService {
	return &voiceService{
		log:          log,
		voiceRepo:    voiceRepo,
		s3Client:     s3Client,
		utils:        utils,
		nlpProcessor: nlpProcessor,
		config:       config,
	}
}
