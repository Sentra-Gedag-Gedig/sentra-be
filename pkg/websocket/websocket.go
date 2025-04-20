package websocketPkg

import (
	"ProjectGolang/internal/api/detection"
	"ProjectGolang/internal/entity"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"sync"
	"time"
)

type IWebsocket interface {
	ProcessFaceFrame(frame []byte) (*entity.DetectionResult, error)
	ProcessKTPFrame(frame []byte) (*entity.KTPDetectionResult, error)
	ProcessQRISFrame(frame []byte) (*entity.QRISDetectionResult, error)
	IsConnected(detectionType detection.DetectionType) bool
	Reconnect(detectionType detection.DetectionType) error
	CloseConnections()
}

type webSocketClient struct {
	faceConn     *websocket.Conn
	ktpConn      *websocket.Conn
	qrisConn     *websocket.Conn
	mu           sync.Mutex
	pingInterval time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewAIWebSocketClient() IWebsocket {
	client := &webSocketClient{
		pingInterval: 30 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 5 * time.Second,
	}

	go client.connectInBackground(detection.FaceDetection)
	go client.connectInBackground(detection.KTPDetection)
	go client.connectInBackground(detection.QRISDetection)

	return client
}

func (c *webSocketClient) connectInBackground(detectionType detection.DetectionType) {
	err := c.Reconnect(detectionType)
	if err != nil {
		log.Printf("Initial connection to %s failed: %v. Will retry on demand.",
			getDetectionTypeName(detectionType), err)
	} else {
		log.Printf("Successfully connected to %s service", getDetectionTypeName(detectionType))
	}
}

func (c *webSocketClient) IsConnected(detectionType detection.DetectionType) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch detectionType {
	case detection.FaceDetection:
		return c.faceConn != nil
	case detection.KTPDetection:
		return c.ktpConn != nil
	case detection.QRISDetection:
		return c.qrisConn != nil
	default:
		return false
	}
}

func (c *webSocketClient) Reconnect(detectionType detection.DetectionType) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if detectionType == detection.FaceDetection && c.faceConn != nil {
		c.faceConn.Close()
		c.faceConn = nil
	} else if detectionType == detection.KTPDetection && c.ktpConn != nil {
		c.ktpConn.Close()
		c.ktpConn = nil
	} else if detectionType == detection.QRISDetection && c.qrisConn != nil {
		c.qrisConn.Close()
		c.qrisConn = nil
	}

	url := getWebSocketURL(detectionType)
	if url == "" {
		return fmt.Errorf("URL for %s detection not configured", getDetectionTypeName(detectionType))
	}

	log.Printf("Connecting to %s at %s", getDetectionTypeName(detectionType), url)

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", url, err)
	}

	conn.SetPingHandler(func(appData string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(c.writeTimeout))
		if err != nil {
			log.Printf("Error sending pong: %v", err)
		}
		return nil
	})

	if detectionType == detection.FaceDetection {
		c.faceConn = conn
	} else if detectionType == detection.KTPDetection {
		c.ktpConn = conn
	} else if detectionType == detection.QRISDetection {
		c.qrisConn = conn
	}

	go c.keepAlive(detectionType)

	return nil
}

func (c *webSocketClient) CloseConnections() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.faceConn != nil {
		c.faceConn.Close()
		c.faceConn = nil
	}

	if c.ktpConn != nil {
		c.ktpConn.Close()
		c.ktpConn = nil
	}

	if c.qrisConn != nil {
		c.qrisConn.Close()
		c.qrisConn = nil
	}
}

func (c *webSocketClient) keepAlive(detectionType detection.DetectionType) {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		var conn *websocket.Conn

		if detectionType == detection.FaceDetection {
			conn = c.faceConn
		} else if detectionType == detection.KTPDetection {
			conn = c.ktpConn
		} else if detectionType == detection.QRISDetection {
			conn = c.qrisConn
		}

		if conn == nil {
			c.mu.Unlock()
			return
		}

		err := conn.WriteControl(
			websocket.PingMessage,
			[]byte{},
			time.Now().Add(c.writeTimeout),
		)

		if err != nil {
			log.Printf("Ping failed for %s, marking connection as dead: %v",
				getDetectionTypeName(detectionType), err)
			if detectionType == detection.FaceDetection {
				c.faceConn = nil
			} else if detectionType == detection.KTPDetection {
				c.ktpConn = nil
			} else if detectionType == detection.QRISDetection {
				c.qrisConn = nil
			}
			conn.Close()
			c.mu.Unlock()
			return
		}

		c.mu.Unlock()
	}
}

func (c *webSocketClient) getConnection(detectionType detection.DetectionType) (*websocket.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var conn *websocket.Conn

	if detectionType == detection.FaceDetection {
		conn = c.faceConn
	} else if detectionType == detection.KTPDetection {
		conn = c.ktpConn
	} else if detectionType == detection.QRISDetection {
		conn = c.qrisConn
	}

	if conn == nil {
		return nil, fmt.Errorf("not connected to %s service", getDetectionTypeName(detectionType))
	}

	return conn, nil
}

func (c *webSocketClient) ProcessFaceFrame(frame []byte) (*entity.DetectionResult, error) {
	conn, err := c.getConnection(detection.FaceDetection)
	if err != nil {
		if err := c.Reconnect(detection.FaceDetection); err != nil {
			return nil, fmt.Errorf("cannot connect to face detection service: %w", err)
		}
		conn, err = c.getConnection(detection.FaceDetection)
		if err != nil {
			return nil, err
		}
	}

	c.mu.Lock()

	conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	log.Printf("Sending face frame of size: %d bytes", len(frame))
	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		c.faceConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error sending face frame: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(c.readTimeout))

	c.mu.Unlock()

	_, message, err := conn.ReadMessage()
	if err != nil {
		c.mu.Lock()
		c.faceConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error reading face message: %w", err)
	}

	c.mu.Lock()
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})
	c.mu.Unlock()

	log.Printf("Received response from face AI service")

	var result entity.DetectionResult
	if err := json.Unmarshal(message, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling face response: %w", err)
	}

	log.Printf("Face Detection Result: Status=%s, Instructions=%v", result.Status, result.Instructions)
	if result.FacePosition != nil {
		log.Printf("Face Position: x=%d, y=%d", result.FacePosition.X, result.FacePosition.Y)
	}
	if result.Deviations != nil {
		log.Printf("Deviations: x=%.2f, y=%.2f", result.Deviations["x"], result.Deviations["y"])
	}

	return &result, nil
}

func (c *webSocketClient) ProcessKTPFrame(frame []byte) (*entity.KTPDetectionResult, error) {
	conn, err := c.getConnection(detection.KTPDetection)
	if err != nil {
		if err := c.Reconnect(detection.KTPDetection); err != nil {
			return nil, fmt.Errorf("cannot connect to KTP detection service: %w", err)
		}
		conn, err = c.getConnection(detection.KTPDetection)
		if err != nil {
			return nil, err
		}
	}

	base64Frame := base64.StdEncoding.EncodeToString(frame)

	c.mu.Lock()

	conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	log.Printf("Sending KTP frame of size: %d bytes", len(base64Frame))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(base64Frame)); err != nil {
		c.ktpConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error sending KTP frame: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(c.readTimeout))

	c.mu.Unlock()

	_, message, err := conn.ReadMessage()
	if err != nil {
		c.mu.Lock()
		c.ktpConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error reading KTP message: %w", err)
	}

	c.mu.Lock()
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})
	c.mu.Unlock()

	log.Printf("Received response from KTP AI service")

	var result entity.KTPDetectionResult
	if err := json.Unmarshal(message, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling KTP response: %w", err)
	}

	if result.BBox != nil && len(result.BBox) == 4 {
		result.KTPPosition = &entity.KTPPosition{
			X1: result.BBox[0],
			Y1: result.BBox[1],
			X2: result.BBox[2],
			Y2: result.BBox[3],
		}
	}

	log.Printf("KTP Detection Result: Message=%s", result.Message)
	if result.KTPPosition != nil {
		log.Printf("KTP Position: x1=%.2f, y1=%.2f, x2=%.2f, y2=%.2f",
			result.KTPPosition.X1, result.KTPPosition.Y1,
			result.KTPPosition.X2, result.KTPPosition.Y2)
	}

	return &result, nil
}

func (c *webSocketClient) ProcessQRISFrame(frame []byte) (*entity.QRISDetectionResult, error) {
	conn, err := c.getConnection(detection.QRISDetection)
	if err != nil {
		if err := c.Reconnect(detection.QRISDetection); err != nil {
			return nil, fmt.Errorf("cannot connect to QRIS detection service: %w", err)
		}
		conn, err = c.getConnection(detection.QRISDetection)
		if err != nil {
			return nil, err
		}
	}

	base64Frame := base64.StdEncoding.EncodeToString(frame)

	c.mu.Lock()

	conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

	log.Printf("Sending QRIS frame of size: %d bytes", len(base64Frame))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(base64Frame)); err != nil {
		c.qrisConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error sending QRIS frame: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(c.readTimeout))

	c.mu.Unlock()

	_, message, err := conn.ReadMessage()
	if err != nil {
		c.mu.Lock()
		c.qrisConn = nil
		conn.Close()
		c.mu.Unlock()
		return nil, fmt.Errorf("error reading QRIS message: %w", err)
	}

	c.mu.Lock()
	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})
	c.mu.Unlock()

	log.Printf("Received response from QRIS AI service")

	var result entity.QRISDetectionResult
	if err := json.Unmarshal(message, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling QRIS response: %w", err)
	}

	if result.BBox != nil && len(result.BBox) == 4 {
		result.QRISPosition = &entity.QRISPosition{
			X1: result.BBox[0],
			Y1: result.BBox[1],
			X2: result.BBox[2],
			Y2: result.BBox[3],
		}
	}

	log.Printf("QRIS Detection Result: Message=%s", result.Message)
	if result.QRISPosition != nil {
		log.Printf("QRIS Position: x1=%.2f, y1=%.2f, x2=%.2f, y2=%.2f",
			result.QRISPosition.X1, result.QRISPosition.Y1,
			result.QRISPosition.X2, result.QRISPosition.Y2)
	}

	return &result, nil
}

func getWebSocketURL(detectionType detection.DetectionType) string {
	switch detectionType {
	case detection.FaceDetection:
		url := os.Getenv("AI_FACE_DETECTION_URL")
		if url == "" {
			url = "ws://localhost:8000/api/v1/face/ws"
		}
		return url
	case detection.KTPDetection:
		url := os.Getenv("AI_KTP_DETECTION_URL")
		if url == "" {
			url = "ws://localhost:8000/api/v1/ktp/ws"
		}
		return url
	case detection.QRISDetection:
		url := os.Getenv("AI_QRIS_DETECTION_URL")
		if url == "" {
			url = "ws://localhost:8001/api/v1/qris/ws"
		}
		return url
	default:
		return ""
	}
}

func getDetectionTypeName(detectionType detection.DetectionType) string {
	switch detectionType {
	case detection.FaceDetection:
		return "Face Detection"
	case detection.KTPDetection:
		return "KTP Detection"
	case detection.QRISDetection:
		return "QRIS Detection"
	default:
		return "Unknown Detection"
	}
}
