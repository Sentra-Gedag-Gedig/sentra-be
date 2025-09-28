package voiceHandler

import (
	"ProjectGolang/internal/api/voice"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
)

func (h *VoiceHandler) ProcessVoiceCommand(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing voice command request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	// Get audio file
	audioFile, err := ctx.FormFile("audio")
	if err != nil {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("audio file is required"), ctx.Path())
	}

	// Get optional context
	//contextStr := ctx.FormValue("context", "{}")
	var contextData map[string]interface{}
	// Parse context if provided (optional)

	req := voice.ProcessVoiceRequest{
		AudioFile: audioFile,
		Context:   contextData,
	}

	// Process voice command
	response, err := h.voiceService.ProcessVoiceCommand(c, userData.ID, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "process_voice_command")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *VoiceHandler) ProcessConfirmation(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing voice confirmation request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	// Get audio file
	audioFile, err := ctx.FormFile("audio")
	if err != nil {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("audio file is required"), ctx.Path())
	}

	// Get required fields
	pendingPageID := ctx.FormValue("pending_page_id")
	sessionID := ctx.FormValue("session_id")

	if pendingPageID == "" || sessionID == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("pending_page_id and session_id are required"), ctx.Path())
	}

	req := voice.ConfirmationRequest{
		AudioFile:     audioFile,
		PendingPageID: pendingPageID,
		SessionID:     sessionID,
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	// Process confirmation
	response, err := h.voiceService.ProcessConfirmation(c, userData.ID, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "process_confirmation")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *VoiceHandler) GetVoiceHistory(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get voice history request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	history, total, err := h.voiceService.GetVoiceHistory(c, userData.ID, page, limit)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_voice_history")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"history": history,
			"total":   total,
			"page":    page,
			"limit":   limit,
		})
	}
}

func (h *VoiceHandler) GetVoiceAnalytics(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get voice analytics request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	analytics, err := h.voiceService.GetVoiceAnalytics(c, userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_voice_analytics")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, analytics)
	}
}

func (h *VoiceHandler) GetSmartSuggestions(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get smart suggestions request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	suggestions, err := h.voiceService.GetSmartSuggestions(c, userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_smart_suggestions")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, suggestions)
	}
}

func (h *VoiceHandler) TestNLPProcessing(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing test NLP request")

	var req voice.NLPTestRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	result, err := h.voiceService.TestNLPProcessing(c, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "test_nlp_processing")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}

func (h *VoiceHandler) GetPageMappings(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get page mappings request")

	mappings, err := h.voiceService.GetPageMappings(c)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_page_mappings")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"mappings": mappings,
		})
	}
}

func (h *VoiceHandler) CreatePageMapping(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing create page mapping request")

	var req voice.PageMapping
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.voiceService.CreatePageMapping(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_page_mapping")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusCreated, fiber.Map{
			"message": "Page mapping created successfully",
		})
	}
}

func (h *VoiceHandler) UpdatePageMapping(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing update page mapping request")

	pageID := ctx.Params("page_id")
	if pageID == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("page_id is required"), ctx.Path())
	}

	var req voice.PageMapping
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.voiceService.UpdatePageMapping(c, pageID, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_page_mapping")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"message": "Page mapping updated successfully",
		})
	}
}

func (h *VoiceHandler) ServeAudioFile(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	filename := ctx.Params("filename")
	if filename == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("filename is required"), ctx.Path())
	}

	// Security check for path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("invalid filename"), ctx.Path())
	}

	audioData, err := h.voiceService.ServeAudioFile(c, filename)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "serve_audio_file")
	}

	ctx.Set("Content-Type", "audio/mpeg")
	ctx.Set("Cache-Control", "public, max-age=3600")

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return ctx.Send(audioData)
	}
}