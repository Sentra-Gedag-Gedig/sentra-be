package voiceHandler

import (
	voiceService "ProjectGolang/internal/api/voice/service"
	"ProjectGolang/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type VoiceHandler struct {
	log          *logrus.Logger
	validator    *validator.Validate
	middleware   middleware.Middleware
	voiceService voiceService.IVoiceService
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	middleware middleware.Middleware,
	vs voiceService.IVoiceService,
) *VoiceHandler {
	return &VoiceHandler{
		log:          log,
		validator:    validate,
		middleware:   middleware,
		voiceService: vs,
	}
}

func (h *VoiceHandler) Start(srv fiber.Router) {
	voice := srv.Group("/voice")

	// All voice endpoints require authentication
	voice.Use(h.middleware.NewTokenMiddleware)

	// Voice command processing
	voice.Post("/command", h.ProcessVoiceCommand)
	voice.Post("/confirm", h.ProcessConfirmation)

	// Voice history and analytics
	voice.Get("/history", h.GetVoiceHistory)
	voice.Get("/analytics", h.GetVoiceAnalytics)

	// Smart suggestions and help
	voice.Get("/suggestions", h.GetSmartSuggestions)

	// NLP testing and page mappings (admin endpoints)
	nlp := voice.Group("/nlp")
	nlp.Post("/test", h.TestNLPProcessing)
	nlp.Get("/mappings", h.GetPageMappings)
	nlp.Post("/mappings", h.CreatePageMapping)
	nlp.Put("/mappings/:page_id", h.UpdatePageMapping)

	// Audio file serving
	voice.Get("/audio/:filename", h.ServeAudioFile)
}