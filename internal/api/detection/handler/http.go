package detectionHandler

import (
	detectionService "ProjectGolang/internal/api/detection/service"
	"ProjectGolang/internal/middleware"
	"ProjectGolang/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

type DetectionHandler struct {
	log              *logrus.Logger
	validator        *validator.Validate
	middleware       middleware.Middleware
	detectionService detectionService.IDetectionService
	utils            utils.IUtils
}

func New(
	log *logrus.Logger,
	validator *validator.Validate,
	middleware middleware.Middleware,
	ds detectionService.IDetectionService,
	utils utils.IUtils,
) *DetectionHandler {
	return &DetectionHandler{
		detectionService: ds,
		log:              log,
		validator:        validator,
		middleware:       middleware,
		utils:            utils,
	}
}

func (h *DetectionHandler) Start(srv fiber.Router) {
	wsMiddleware := func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}

	face := srv.Group("/face")
	face.Use("/ws", wsMiddleware)
	face.Get("/ws", websocket.New(h.handleWebSocket))

	ktp := srv.Group("/ktp")
	ktp.Use("/ws", wsMiddleware)
	ktp.Get("/ws", websocket.New(h.handleKTPWebSocket))
	ktp.Post("/extract", h.ExtractKTP)

	qris := srv.Group("/qris")
	qris.Use("/ws", wsMiddleware)
	qris.Get("/ws", websocket.New(h.handleQRISWebSocket))

	srv.Post("/money", h.DetectMoney)

}
