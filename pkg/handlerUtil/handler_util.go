package handlerUtil

import (
	"ProjectGolang/internal/api/auth"
	"ProjectGolang/internal/api/budget_manager"
	"ProjectGolang/internal/api/detection"
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	"ProjectGolang/pkg/log"
	"ProjectGolang/pkg/response"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type ErrorHandler struct {
	logger *logrus.Logger
}

func New(logger *logrus.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

func (h *ErrorHandler) Handle(c *fiber.Ctx, requestID string, err error, path string, operation string) error {
	var respErr *response.Error
	if errors.As(err, &respErr) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"code":       respErr.Code,
			"path":       path,
			"operation":  operation,
		}).Warn("Operation failed with error response")
		return c.Status(respErr.Code).JSON(fiber.Map{"error": err.Error()})
	}

	// Auth domain errors
	if errors.Is(err, auth.ErrPhoneNumberAlreadyExists) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Phone number already exists")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Phone number already exists",
			"code":    "PHONE_NUMBER_ALREADY_EXISTS",
		})
	}

	if errors.Is(err, auth.ErrUserNotFound) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("User not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
			"code":    "USER_NOT_FOUND",
		})
	}

	if errors.Is(err, auth.ErrInvalidEmailOrPassword) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid email or password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid email or password",
			"code":    "INVALID_CREDENTIALS",
		})
	}

	if errors.Is(err, auth.ErrorTokenExpired) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("OTP has expired")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "OTP has expired",
			"code":    "EXPIRED_OTP",
		})
	}

	if errors.Is(err, auth.ErrorInvalidToken) || errors.Is(err, auth.ErrInvalidOTP) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid OTP provided")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid OTP provided",
			"code":    "INVALID_OTP",
		})
	}

	if errors.Is(err, auth.ErrPasswordSame) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("New password is the same as old password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "New password cannot be the same as old password",
			"code":    "PASSWORD_SAME",
		})
	}

	if errors.Is(err, auth.ErrInvalidPhoneNumber) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid phone number")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid phone number",
			"code":    "INVALID_PHONE",
		})
	}

	if errors.Is(err, auth.ErrInvalidFileType) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid file type")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file type. Only images are allowed.",
		})
	}

	if errors.Is(err, auth.ErrFileTooLarge) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("File too large")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File too large. Maximum size is 5MB.",
		})
	}

	if errors.Is(err, auth.ErrFailedToUploadFile) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Failed to upload file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload file",
		})
	}

	if errors.Is(err, auth.ErrEmailAlreadyInUse) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Email already in use")
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already in use by another user",
		})
	}

	// Budget manager domain errors
	if errors.Is(err, budget_manager.ErrTransactionNotFound) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Transaction not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	if errors.Is(err, budget_manager.ErrInvalidCategory) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid category")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category for the transaction type",
		})
	}

	if errors.Is(err, budget_manager.ErrTransactionNotOwned) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Transaction does not belong to user")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Transaction does not belong to user",
		})
	}

	if errors.Is(err, budget_manager.ErrInvalidTransactionType) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid transaction type")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction type",
		})
	}

	if errors.Is(err, budget_manager.ErrInvalidAudioFile) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid audio file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audio file type",
		})
	}

	if errors.Is(err, budget_manager.ErrFailedToUploadAudio) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Failed to upload audio file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload audio file",
		})
	}

	// SentraPay domain errors
	if errors.Is(err, sentrapay.ErrInvalidBank) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid bank selection")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid bank selection",
		})
	}

	if errors.Is(err, sentrapay.ErrInvalidAmount) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid amount")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid amount",
		})
	}

	if errors.Is(err, sentrapay.ErrCreateVirtualAccount) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Failed to create virtual account")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create virtual account",
		})
	}

	if errors.Is(err, sentrapay.ErrTransactionNotFound) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Transaction not found")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	if errors.Is(err, sentrapay.ErrInvalidCallback) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Warn("Invalid callback data")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid callback data",
		})
	}

	// Detection domain errors
	if errors.Is(err, detection.ErrInternalServerError) {
		h.logger.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"path":       path,
			"operation":  operation,
		}).Error("Internal server error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	h.logger.WithFields(log.Fields{
		"request_id": requestID,
		"error":      err.Error(),
		"path":       path,
		"operation":  operation,
	}).Error("Unexpected error")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "An unexpected error occurred",
	})
}

func (h *ErrorHandler) HandleValidationError(c *fiber.Ctx, requestID string, err error, path string) error {
	h.logger.WithFields(log.Fields{
		"request_id": requestID,
		"error":      err.Error(),
		"path":       path,
	}).Warn("Validation failed")

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Validation failed: " + err.Error(),
		"code":  "VALIDATION_ERROR",
	})
}

func (h *ErrorHandler) HandleRequestTimeout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusRequestTimeout).JSON(utils.StatusMessage(fiber.StatusRequestTimeout))
}

func (h *ErrorHandler) HandleUnauthorized(c *fiber.Ctx, requestID string, message string) error {
	h.logger.WithFields(log.Fields{
		"request_id": requestID,
		"path":       c.Path(),
		"message":    message,
	}).Warn("Unauthorized access")

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": message,
		"code":  "UNAUTHORIZED",
	})
}

func (h *ErrorHandler) HandleSuccess(c *fiber.Ctx, statusCode int, data interface{}) error {
	if data == nil {
		return c.SendStatus(statusCode)
	}
	return c.Status(statusCode).JSON(data)
}
