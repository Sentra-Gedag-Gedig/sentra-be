package detectionHandler

import (
	"ProjectGolang/internal/api/detection"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	"ProjectGolang/pkg/log"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/net/context"
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
	defer func() {
		h.log.Info("KTP detection WebSocket client disconnected")
		c.Close()
	}()

	
	maxReadTimeout := 60 * time.Second
	writeTimeout := 10 * time.Second
	
	
	c.SetCloseHandler(func(code int, text string) error {
		switch code {
		case websocket.CloseNormalClosure:
			h.log.Info("KTP WebSocket closed normally")
		case websocket.CloseGoingAway:
			h.log.Info("KTP WebSocket client going away (page refresh/navigation)")
		case websocket.CloseNoStatusReceived:
			h.log.Info("KTP WebSocket closed without status (likely client-initiated)")
		case websocket.CloseAbnormalClosure:
			h.log.Warn("KTP WebSocket closed abnormally")
		default:
			h.log.Infof("KTP WebSocket closed with code: %d, reason: %s", code, text)
		}
		return nil
	})

	
	c.SetPingHandler(func(data string) error {
		h.log.Debug("Received ping from KTP client, sending pong")
		pongDeadline := time.Now().Add(5 * time.Second)
		if err := c.WriteControl(websocket.PongMessage, []byte(data), pongDeadline); err != nil {
			h.log.Errorf("Error sending pong to KTP client: %v", err)
			return err
		}
		return nil
	})

	
	c.SetPongHandler(func(data string) error {
		h.log.Debug("Received pong from KTP client")
		return nil
	})

	
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()
	
	done := make(chan bool)
	defer close(done)

	go func() {
		for {
			select {
			case <-pingTicker.C:
				pingDeadline := time.Now().Add(writeTimeout)
				if err := c.WriteControl(websocket.PingMessage, nil, pingDeadline); err != nil {
					h.log.Errorf("Error sending ping to KTP client: %v", err)
					return
				}
				h.log.Debug("Sent ping to KTP client")
			case <-done:
				return
			}
		}
	}()

	
	for {
		
		if err := c.SetReadDeadline(time.Now().Add(maxReadTimeout)); err != nil {
			h.log.Errorf("Error setting read deadline for KTP WebSocket: %v", err)
			break
		}

		
		messageType, message, err := c.ReadMessage()
		if err != nil {
			
			if websocket.IsCloseError(err, 
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure) {
				
				
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					h.log.Info("KTP WebSocket connection closed by client")
				} else {
					h.log.Warn("KTP WebSocket connection closed unexpectedly by client")
				}
				break
			}
			
			
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				h.log.Warn("KTP WebSocket read timeout - client may be inactive")
			} else {
				h.log.Errorf("KTP WebSocket read error: %v", err)
			}
			break
		}

		
		if err := c.SetReadDeadline(time.Time{}); err != nil {
			h.log.Errorf("Error resetting read deadline for KTP WebSocket: %v", err)
		}

		
		if messageType == websocket.BinaryMessage {
			h.log.Infof("Received KTP binary frame of size: %d bytes", len(message))

			
			result, err := h.detectionService.ProcessKTPFrame(message)
			if err != nil {
				h.log.Errorf("Error processing KTP frame: %v", err)
				
				
				errorResponse := map[string]interface{}{
					"error": true,
					"message": "Failed to process KTP frame",
					"details": err.Error(),
				}
				
				if writeErr := h.writeJSONWithTimeout(c, errorResponse, writeTimeout); writeErr != nil {
					h.log.Errorf("Error sending KTP error response: %v", writeErr)
					break
				}
				continue
			}

			
			h.log.Infof("KTP frame processed successfully - Message: %s", result.Message)
			
			if err := h.writeJSONWithTimeout(c, result, writeTimeout); err != nil {
				h.log.Errorf("Error sending KTP success response: %v", err)
				break
			}
			
			h.log.Debug("KTP detection result sent to client successfully")

		} else if messageType == websocket.TextMessage {
			h.log.Debugf("Received KTP text message: %s", string(message))
			
			
		} else {
			h.log.Warnf("Received unexpected KTP message type: %d", messageType)
		}
	}
	
	
	done <- true
	h.log.Debug("KTP WebSocket handler cleanup completed")
}


func (h *DetectionHandler) writeJSONWithTimeout(c *websocket.Conn, data interface{}, timeout time.Duration) error {
	
	if err := c.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	
	if err := c.WriteJSON(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	
	if err := c.SetWriteDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to reset write deadline: %w", err)
	}

	return nil
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
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
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
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
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
