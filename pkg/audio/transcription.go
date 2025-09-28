
package audio

import (
	"context"
	"os"
	openai "github.com/sashabaranov/go-openai"
)

type TranscriptionService struct {
	client *openai.Client
}

func NewTranscriptionService(apiKey string) *TranscriptionService {
	client := openai.NewClient(apiKey)
	return &TranscriptionService{client: client}
}

func (t *TranscriptionService) TranscribeAudio(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filePath,
		Language: "id", // Indonesian language
	}

	resp, err := t.client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
