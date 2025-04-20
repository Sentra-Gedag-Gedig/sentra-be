package sentrapayHandler

import (
	sentrapayService "ProjectGolang/internal/api/sentra_pay/service"
	"ProjectGolang/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type SentraPayHandler struct {
	log              *logrus.Logger
	validator        *validator.Validate
	middleware       middleware.Middleware
	sentraPayService sentrapayService.ISentraPayService
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	middleware middleware.Middleware,
	sps sentrapayService.ISentraPayService,
) *SentraPayHandler {
	return &SentraPayHandler{
		log:              log,
		validator:        validate,
		middleware:       middleware,
		sentraPayService: sps,
	}
}

func (h *SentraPayHandler) Start(srv fiber.Router) {
	wallet := srv.Group("/wallet")

	wallet.Post("/topup", h.middleware.NewTokenMiddleware, h.CreateTopUp)
	wallet.Get("/balance", h.middleware.NewTokenMiddleware, h.GetWalletBalance)
	wallet.Get("/transactions", h.middleware.NewTokenMiddleware, h.GetTransactionHistory)
	wallet.Get("/transactions/status/:reference_no", h.middleware.NewTokenMiddleware, h.CheckTransactionStatus)

	wallet.Post("/callback", h.PaymentCallback)
}
