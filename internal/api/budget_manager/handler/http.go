package budgetHandler

import (
	budgetService "ProjectGolang/internal/api/budget_manager/service"
	"ProjectGolang/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type BudgetHandler struct {
	log           *logrus.Logger
	validator     *validator.Validate
	middleware    middleware.Middleware
	budgetService budgetService.IBudgetService
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	middleware middleware.Middleware,
	budgetService budgetService.IBudgetService,
) *BudgetHandler {
	return &BudgetHandler{
		log:           log,
		validator:     validate,
		middleware:    middleware,
		budgetService: budgetService,
	}
}

func (h *BudgetHandler) Start(srv fiber.Router) {
	budget := srv.Group("/budget")

	budget.Post("/transactions", h.middleware.NewTokenMiddleware, h.CreateTransaction)
	budget.Get("/transactions", h.middleware.NewTokenMiddleware, h.GetTransactionsByUserID)
	budget.Get("/transactions/period", h.middleware.NewTokenMiddleware, h.GetTransactionsByPeriod)
	budget.Get("/transactions/filter", h.middleware.NewTokenMiddleware, h.GetTransactionsByTypeAndCategory)
	budget.Get("/transactions/:id", h.GetTransactionByID)
	budget.Put("/transactions", h.middleware.NewTokenMiddleware, h.UpdateTransaction)
	budget.Delete("/transactions/:id", h.middleware.NewTokenMiddleware, h.DeleteTransaction)
}
