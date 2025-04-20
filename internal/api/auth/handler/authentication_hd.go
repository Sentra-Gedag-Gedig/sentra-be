package authHandler

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"time"
)

func (h *AuthHandler) HandlePhoneNumberVerification(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.VerifyPhoneNumberRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.Auth().PhoneNumberVerification(c, req.PhoneNumber); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "phone_verification")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleVerifyOTPandPIN(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.OTPPINRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.Auth().VerifyOTPandUpdatePIN(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "verify_otp_update_pin")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleVerifyOTPandPassword(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.ResetPassword
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.Password().UpdatePassword(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_password")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleVerifyOTPandVerifyingUser(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.VerifyUserUsingOTP
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.User().UpdateUserVerifiedStatusAndPIN(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "verify_user_with_otp")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleLogin(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.LoginUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	res, err := h.authService.Auth().Login(c, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "login")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, res)
	}
}

func (h *AuthHandler) LoginTouchID(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.TouchIDLoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	res, err := h.authService.Biometric().LoginTouchID(c, req)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "login_touch_id")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, res)
	}
}

func (h *AuthHandler) EnableTouchID(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	plain, err := h.authService.Biometric().EnableTouchID(ctx.Context(), userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "enable_touch_id")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, plain)
	}
}

func (h *AuthHandler) HandleSendEmailOTP(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing email OTP request")

	var req auth.SendEmailOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.Auth().SendEmailOTP(c, req.Email); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "send_email_otp")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleVerifyEmailOTP(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing email OTP verification request")

	var req auth.VerifyEmailOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	if err := h.authService.Auth().VerifyEmailOTP(c, userData.ID, req.Email, req.Code); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "verify_email_otp")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleVerifyPhoneOTP(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing phone OTP verification request")

	var req auth.VerifyPhoneOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	if err := h.authService.Auth().VerifyPhoneOTP(c, userData.ID, req.PhoneNumber, req.Code); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "verify_phone_otp")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}
