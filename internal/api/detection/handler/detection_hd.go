package detectionHandler

import (
	"ProjectGolang/internal/api/detection"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	"ProjectGolang/pkg/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/net/context"
	"time"
)

func (h *DetectionHandler) handleWebSocket(c *websocket.Conn) {
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.BinaryMessage {
			result, err := h.detectionService.ProcessFrame(message)
			if err != nil {
				continue
			}

			if err := c.WriteJSON(result); err != nil {
				break
			}
		}
	}
}

func (h *DetectionHandler) handleKTPWebSocket(c *websocket.Conn) {
	h.log.Info("KTP detection WebSocket client connected")
	defer h.log.Info("KTP detection WebSocket client disconnected")

	c.SetPingHandler(func(data string) error {
		h.log.Debug("Received ping, sending pong")
		if err := c.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(5*time.Second)); err != nil {
			h.log.Errorf("Error sending pong: %v", err)
		}
		return nil
	})

	maxReadTimeout := 60 * time.Second

	for {
		if err := c.SetReadDeadline(time.Now().Add(maxReadTimeout)); err != nil {
			h.log.Errorf("Error setting read deadline: %v", err)
			break
		}

		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log.Errorf("KTP WebSocket error: %v", err)
			} else {
				h.log.Info("KTP WebSocket connection closed")
			}
			break
		}

		if messageType == websocket.BinaryMessage {
			h.log.Info("Received binary message for KTP detection")

			result, err := h.detectionService.ProcessKTPFrame(message)

			if err != nil {
				h.log.Errorf("Error processing KTP frame: %v", err)
				if writeErr := c.WriteJSON(map[string]string{"error": err.Error()}); writeErr != nil {
					h.log.Errorf("Error sending error response: %v", writeErr)
					break
				}
				continue
			}

			if err := c.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				h.log.Errorf("Error setting write deadline: %v", err)
				break
			}

			if err := c.WriteJSON(result); err != nil {
				h.log.Errorf("Error writing JSON response: %v", err)
				break
			}

			if err := c.SetWriteDeadline(time.Time{}); err != nil {
				h.log.Errorf("Error resetting write deadline: %v", err)
				break
			}
		} else {
			h.log.Warnf("Received unexpected message type: %d", messageType)
		}
	}
}

func (h *DetectionHandler) handleQRISWebSocket(c *websocket.Conn) {
	h.log.Info("QRIS detection WebSocket client connected")
	defer h.log.Info("QRIS detection WebSocket client disconnected")

	c.SetPingHandler(func(data string) error {
		h.log.Debug("Received ping, sending pong")
		if err := c.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(5*time.Second)); err != nil {
			h.log.Errorf("Error sending pong: %v", err)
		}
		return nil
	})

	maxReadTimeout := 60 * time.Second

	for {
		if err := c.SetReadDeadline(time.Now().Add(maxReadTimeout)); err != nil {
			h.log.Errorf("Error setting read deadline: %v", err)
			break
		}

		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log.Errorf("QRIS WebSocket error: %v", err)
			} else {
				h.log.Info("QRIS WebSocket connection closed")
			}
			break
		}

		if messageType == websocket.BinaryMessage {
			h.log.Info("Received binary message for QRIS detection")

			result, err := h.detectionService.ProcessQRISFrame(message)

			if err != nil {
				h.log.Errorf("Error processing QRIS frame: %v", err)
				if writeErr := c.WriteJSON(map[string]string{"error": err.Error()}); writeErr != nil {
					h.log.Errorf("Error sending error response: %v", writeErr)
					break
				}
				continue
			}

			if err := c.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				h.log.Errorf("Error setting write deadline: %v", err)
				break
			}

			if err := c.WriteJSON(result); err != nil {
				h.log.Errorf("Error writing JSON response: %v", err)
				break
			}

			if err := c.SetWriteDeadline(time.Time{}); err != nil {
				h.log.Errorf("Error resetting write deadline: %v", err)
				break
			}
		} else {
			h.log.Warnf("Received unexpected message type: %d", messageType)
		}
	}
}

func (h *DetectionHandler) ExtractKTP(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing KTP extraction request")

	var base64Image string
	var err error

	file, err := ctx.FormFile("image")
	if err == nil {
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
			"file_name":  file.Filename,
			"file_size":  file.Size,
		}).Debug("Processing file upload")

		if err := h.utils.ValidateImageFile(file); err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "validate_image_file")
		}

		fileContent, err := file.Open()
		if err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "open_file")
		}
		defer fileContent.Close()

		base64Image, err = h.utils.ConvertFileToBase64(fileContent)
		if err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "convert_to_base64")
		}
	} else {
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
		}).Debug("Processing JSON request")

		var req detection.OCRRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
		}

		if err := h.validator.Struct(req); err != nil {
			return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
		}

		base64Image = req.ImageBase64
	}

	ktp, err := h.detectionService.ExtractAndSaveKTP(c, base64Image)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "extract_ktp")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
			"nik":        ktp.NIK,
		}).Info("KTP extraction successful")
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, detection.OCRResponse{
			Data: *ktp,
		})
	}
}

func (h *DetectionHandler) DetectMoney(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing money detection request")

	var base64Image string
	var err error

	file, err := ctx.FormFile("image")
	if err == nil {
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
			"file_name":  file.Filename,
			"file_size":  file.Size,
		}).Debug("Processing file upload")

		if err := h.utils.ValidateImageFile(file); err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "validate_image_file")
		}

		fileContent, err := file.Open()
		if err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "open_file")
		}
		defer fileContent.Close()

		base64Image, err = h.utils.ConvertFileToBase64(fileContent)
		if err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "convert_to_base64")
		}
	} else {
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
		}).Debug("Processing JSON request")

		var req detection.MoneyDetectionRequest
		if err := ctx.BodyParser(&req); err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
		}

		if err := h.validator.Struct(req); err != nil {
			return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
		}

		base64Image = req.ImageBase64
	}

	result, err := h.detectionService.DetectMoney(c, base64Image)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "detect_money")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"path":       ctx.Path(),
			"total":      result.Total,
			"multiple":   result.Multiple,
		}).Info("Money detection successful")
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, detection.MoneyResponse{
			Data: *result,
		})
	}
}
