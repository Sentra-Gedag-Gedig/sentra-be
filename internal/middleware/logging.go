package middleware

import (
	"ProjectGolang/pkg/log"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func LoggerConfig() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID, ok := c.Locals("X-Request-ID").(string)
		if !ok || requestID == "" {
			requestID = "unknown"
		}

		c.Locals("request_id", requestID)

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		if err != nil && status == fiber.StatusInternalServerError {
			return err
		}

		logFields := log.Fields{
			"request_id":    requestID,
			"method":        c.Method(),
			"path":          c.Path(),
			"status":        status,
			"latency_ms":    latency.Milliseconds(),
			"ip":            c.IP(),
			"host":          c.Hostname(),
			"user_agent":    c.Get("User-Agent"),
			"referer":       c.Get("Referer"),
			"response_size": len(c.Response().Body()),
		}

		if c.Request().Body() != nil && len(c.Request().Body()) > 0 {
			sanitizedBody := sanitizeRequestBody(c.Path(), string(c.Request().Body()))
			logFields["request_body"] = sanitizedBody
		}

		if status >= 500 {
			log.Error(logFields, "Server error")
		} else if status >= 400 {
			log.Warn(logFields, "Client error")
		} else {
			log.Info(logFields, "Success")
		}

		return err
	}
}

func sanitizeRequestBody(path string, body string) string {
	var jsonBody map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonBody); err != nil {
		return "[non-JSON body]"
	}

	sensitiveFields := []string{
		"password", "token", "secret", "key", "auth",
		"credential", "authorization", "pin", "security_question",
		"security_answer", "credit_card", "card_number", "cvv",
		"ssn", "social_security", "passport", "license",
	}

	if strings.Contains(path, "/users") || strings.Contains(path, "/auth") {
		sensitiveFields = append(sensitiveFields, "password_confirmation", "old_password", "new_password")
	}

	for _, field := range sensitiveFields {
		if _, exists := jsonBody[field]; exists {
			jsonBody[field] = "[SECRET]"
		}
	}

	sanitized, err := json.Marshal(jsonBody)
	if err != nil {
		return "[sanitization-failed]"
	}

	return string(sanitized)
}

type loggingMiddleware struct {
	logger *logrus.Logger
}

func newLoggingMiddleware(logger *logrus.Logger) *loggingMiddleware {
	return &loggingMiddleware{
		logger: logger,
	}
}

func (m *middleware) NewLoggingMiddleware(ctx *fiber.Ctx) error {
	return ctx.Next()
}
