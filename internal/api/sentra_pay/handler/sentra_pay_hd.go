package sentrapayHandler

import (
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"errors"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

func (h *SentraPayHandler) CreateTopUp(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing top-up request")

	var req sentrapay.TopUpRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	response, err := h.sentraPayService.CreateTopUpTransaction(c, userData.ID, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_topup_transaction")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusCreated, response)
	}
}

func (h *SentraPayHandler) PaymentCallback(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing payment callback")

	var req sentrapay.PaymentCallbackRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	channelID := ctx.Get("CHANNEL-ID")
	xExternalID := ctx.Get("X-EXTERNAL-ID")
	xTimestamp := ctx.Get("X-TIMESTAMP")
	xPartnerID := ctx.Get("X-PARTNER-ID")

	h.log.WithFields(log.Fields{
		"request_id":    requestID,
		"channelID":     channelID,
		"xExternalID":   xExternalID,
		"xTimestamp":    xTimestamp,
		"xPartnerID":    xPartnerID,
		"trxId":         req.TrxId,
		"paidAmount":    req.PaidAmount.Value,
		"paymentMethod": req.AdditionalInfo.Channel,
	}).Info("Received payment callback")

	if err := h.sentraPayService.ProcessPaymentCallback(c, req, channelID, xExternalID, xTimestamp, xPartnerID); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "process_payment_callback")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"responseCode":    "2002500",
			"responseMessage": "success",
			"virtualAccountData": map[string]interface{}{
				"partnerServiceId":   req.PartnerServiceId,
				"customerNo":         req.CustomerNo,
				"virtualAccountNo":   req.VirtualAccountNo,
				"virtualAccountName": req.VirtualAccountName,
				"paymentRequestId":   req.PaymentRequestId,
				"trxId":              req.TrxId,
				"trxDateTime":        req.TrxDateTime,
				"additionalInfo": map[string]interface{}{
					"channel": req.AdditionalInfo.Channel,
				},
			},
		})
	}
}

func (h *SentraPayHandler) GetWalletBalance(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get wallet balance request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	balance, err := h.sentraPayService.GetWalletBalance(c, userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_wallet_balance")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, balance)
	}
}

func (h *SentraPayHandler) GetTransactionHistory(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get transaction history request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"user_id":    userData.ID,
		"page":       page,
		"limit":      limit,
	}).Debug("Fetching transaction history")

	history, err := h.sentraPayService.GetTransactionHistory(c, userData.ID, page, limit)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_transaction_history")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, history)
	}
}

func (h *SentraPayHandler) CheckTransactionStatus(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing check transaction status request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	referenceNo := ctx.Params("reference_no")
	if referenceNo == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("Reference number is required"), ctx.Path())
	}

	h.log.WithFields(log.Fields{
		"request_id":   requestID,
		"user_id":      userData.ID,
		"reference_no": referenceNo,
	}).Debug("Checking transaction status")

	status, err := h.sentraPayService.CheckTransactionStatus(c, referenceNo)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "check_transaction_status")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"reference_no": referenceNo,
			"status":       status,
			"user_id":      userData.ID,
		})
	}
}
