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

	
	voice.Use(h.middleware.NewTokenMiddleware)

	
	voice.Post("/command", h.ProcessVoiceCommand)
	voice.Post("/confirm", h.ProcessConfirmation)

	
	voice.Post("/chat", h.ProcessChatCommand)

	
	voice.Get("/history", h.GetVoiceHistory)
	voice.Get("/analytics", h.GetVoiceAnalytics)

	
	voice.Get("/suggestions", h.GetSmartSuggestions)

	
	nlp := voice.Group("/nlp")
	nlp.Post("/test", h.TestNLPProcessing)
	nlp.Get("/mappings", h.GetPageMappings)
	nlp.Post("/mappings", h.CreatePageMapping)
	nlp.Put("/mappings/:page_id", h.UpdatePageMapping)

	
	voice.Get("/audio/:filename", h.ServeAudioFile)
}