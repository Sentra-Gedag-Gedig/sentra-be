package authHandler

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	handlerutil "ProjectGolang/pkg/handlerUtil"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"time"
)

func (h *AuthHandler) HandleResetPassword(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerutil.New(h.log)

	var req auth.ResetPassword
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.Password().UpdatePassword(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "reset_password")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}
