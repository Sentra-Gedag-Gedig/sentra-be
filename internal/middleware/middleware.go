package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	NewRateLimiter(ctx *fiber.Ctx) error
	NewTokenMiddleware(ctx *fiber.Ctx) error
	NewRequestIDMiddleware() fiber.Handler
	GetRequestID(ctx *fiber.Ctx) string
}

type middleware struct {
	token               *tokenMiddleware
	rateLimitter        *rateLimiter
	loggingMiddleware   *loggingMiddleware
	requestIDMiddleware fiber.Handler
	log                 *logrus.Logger
}

func New(logger *logrus.Logger) Middleware {
	rateLimit := newRateLimiter(50, 100)
	token := newTokenMiddleware()
	logging := newLoggingMiddleware(logger)
	requestID := NewRequestIDMiddleware()

	return &middleware{
		token:               token,
		rateLimitter:        rateLimit,
		loggingMiddleware:   logging,
		requestIDMiddleware: requestID,
		log:                 logger,
	}
}

func (m *middleware) GetRequestID(ctx *fiber.Ctx) string {
	requestID, ok := ctx.Locals(RequestIDKey).(string)
	if !ok || requestID == "" {
		return "unknown"
	}
	return requestID
}

func (m *middleware) NewRequestIDMiddleware() fiber.Handler {
	return m.requestIDMiddleware
}
