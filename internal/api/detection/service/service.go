package detectionService

import (
	"ProjectGolang/internal/api/detection"
	"ProjectGolang/internal/entity"
	"ProjectGolang/pkg/gemini"
	websocketPkg "ProjectGolang/pkg/websocket"
	"golang.org/x/net/context"
)

type IDetectionService interface {
	ProcessFrame(frame []byte) (*entity.DetectionResult, error)
	ProcessKTPFrame(frame []byte) (*entity.KTPDetectionResult, error)
	ProcessQRISFrame(frame []byte) (*entity.QRISDetectionResult, error)
	ExtractAndSaveKTP(ctx context.Context, base64Image string) (*detection.KTP, error)
	DetectMoney(ctx context.Context, base64Image string) (*detection.MoneyDetectionResponse, error)
}

type detectionService struct {
	websocketPkg websocketPkg.IWebsocket
	gemini       gemini.IGemini
}

func NewDetectionService(
	websocket websocketPkg.IWebsocket,
	gemini gemini.IGemini,
) IDetectionService {
	return &detectionService{
		websocketPkg: websocket,
		gemini:       gemini,
	}
}
