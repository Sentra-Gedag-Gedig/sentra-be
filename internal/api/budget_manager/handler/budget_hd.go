package budgetHandler

import (
	"ProjectGolang/internal/api/budget_manager"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"errors"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"time"
)

func (h *BudgetHandler) CreateTransaction(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing create transaction request")

	var req budget_manager.CreateTransactionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	req.UserID = userData.ID

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	audioFile, _ := ctx.FormFile("audio")

	if err := h.budgetService.CreateTransaction(c, req, audioFile); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_transaction")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusCreated, fiber.Map{
			"message": "Transaction created successfully",
		})
	}
}

func (h *BudgetHandler) GetTransactionByID(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get transaction by ID request")

	id := ctx.Params("id")
	if id == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("transaction ID is required"), ctx.Path())
	}

	transaction, err := h.budgetService.GetTransactionByID(c, id)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_transaction")
	}

	response := budget_manager.TransactionResponse{
		ID:          transaction.ID,
		UserID:      transaction.UserID,
		Title:       transaction.Title,
		Description: transaction.Description,
		Nominal:     transaction.Nominal,
		Type:        transaction.Type,
		Category:    transaction.Category,
		AudioLink:   transaction.AudioLink,
		CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   transaction.UpdatedAt.Format(time.RFC3339),
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *BudgetHandler) GetTransactionsByUserID(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get transactions by user ID request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	transactions, err := h.budgetService.GetTransactionsByUserID(c, userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_transactions")
	}

	var (
		transactionResponses []budget_manager.TransactionResponse
		totalIncome          float64
		totalExpense         float64
	)

	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, budget_manager.TransactionResponse{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			Title:       transaction.Title,
			Description: transaction.Description,
			Nominal:     transaction.Nominal,
			Type:        transaction.Type,
			Category:    transaction.Category,
			AudioLink:   transaction.AudioLink,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   transaction.UpdatedAt.Format(time.RFC3339),
		})

		if transaction.Type == "income" {
			totalIncome += transaction.Nominal
		} else if transaction.Type == "expense" {
			totalExpense += transaction.Nominal
		}
	}

	response := budget_manager.TransactionListResponse{
		Transactions: transactionResponses,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Balance:      totalIncome - totalExpense,
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *BudgetHandler) GetTransactionsByPeriod(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get transactions by period request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	period := ctx.Query("period", "all")

	if period != "all" && period != "week" && period != "month" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("invalid period parameter"), ctx.Path())
	}

	transactions, err := h.budgetService.GetTransactionsByPeriod(c, userData.ID, period)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_transactions_by_period")
	}

	var (
		transactionResponses []budget_manager.TransactionResponse
		totalIncome          float64
		totalExpense         float64
	)

	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, budget_manager.TransactionResponse{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			Title:       transaction.Title,
			Description: transaction.Description,
			Nominal:     transaction.Nominal,
			Type:        transaction.Type,
			Category:    transaction.Category,
			AudioLink:   transaction.AudioLink,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   transaction.UpdatedAt.Format(time.RFC3339),
		})

		if transaction.Type == "income" {
			totalIncome += transaction.Nominal
		} else if transaction.Type == "expense" {
			totalExpense += transaction.Nominal
		}
	}

	response := budget_manager.TransactionListResponse{
		Transactions: transactionResponses,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Balance:      totalIncome - totalExpense,
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *BudgetHandler) UpdateTransaction(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing update transaction request")

	var req budget_manager.UpdateTransactionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	req.UserID = userData.ID

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	audioFile, _ := ctx.FormFile("audio")

	if err := h.budgetService.UpdateTransaction(c, req, audioFile); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_transaction")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"message": "Transaction updated successfully",
		})
	}
}

func (h *BudgetHandler) DeleteTransaction(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing delete transaction request")

	id := ctx.Params("id")
	if id == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("transaction ID is required"), ctx.Path())
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	if err := h.budgetService.DeleteTransaction(c, id, userData.ID); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "delete_transaction")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"message": "Transaction deleted successfully",
		})
	}
}

func (h *BudgetHandler) GetTransactionsByTypeAndCategory(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get transactions by type and category request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	transactionType := ctx.Query("type")
	category := ctx.Query("category")

	if transactionType == "" || category == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("transaction type and category are required"), ctx.Path())
	}

	transactions, err := h.budgetService.GetTransactionsByTypeAndCategory(c, userData.ID, transactionType, category)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_transactions_by_type_and_category")
	}

	var transactionResponses []budget_manager.TransactionResponse
	var total float64

	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, budget_manager.TransactionResponse{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			Title:       transaction.Title,
			Description: transaction.Description,
			Nominal:     transaction.Nominal,
			Type:        transaction.Type,
			Category:    transaction.Category,
			AudioLink:   transaction.AudioLink,
			CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   transaction.UpdatedAt.Format(time.RFC3339),
		})

		total += transaction.Nominal
	}

	response := struct {
		Transactions []budget_manager.TransactionResponse `json:"transactions"`
		Total        float64                              `json:"total"`
		Type         string                               `json:"type"`
		Category     string                               `json:"category"`
	}{
		Transactions: transactionResponses,
		Total:        total,
		Type:         transactionType,
		Category:     category,
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}
