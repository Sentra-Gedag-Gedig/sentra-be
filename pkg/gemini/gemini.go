package gemini

import (
	"context"
	"encoding/base64"
	"errors"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type IGemini interface {
	AnalyzeImage(ctx context.Context, base64Image string, prompt string) (string, error)
}

type geminiClient struct {
	apiKey    string
	modelName string
	client    *genai.Client
}

func NewGeminiClient() (IGemini, error) {

	apiKey := os.Getenv("GEMINI_API_KEY")

	modelName := os.Getenv("GEMINI_MODEL_NAME")
	if apiKey == "" {
		return nil, errors.New("gemini API key is required")
	}

	if modelName == "" {
		modelName = "gemini-pro-vision"
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &geminiClient{
		apiKey:    apiKey,
		modelName: modelName,
		client:    client,
	}, nil
}

func (g *geminiClient) AnalyzeImage(ctx context.Context, base64Image string, prompt string) (string, error) {
	imgData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return "", errors.New("invalid base64 image data")
	}

	model := g.client.GenerativeModel(g.modelName)

	if prompt == "" {
		prompt = "Analyze this image and provide details in JSON format."
	}

	img := genai.ImageData("image/jpeg", imgData)
	res, err := model.GenerateContent(ctx, genai.Text(prompt), img)
	if err != nil {
		return "", err
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from Gemini API")
	}

	response := res.Candidates[0].Content.Parts[0]
	text, ok := response.(genai.Text)
	if !ok {
		return "", errors.New("unexpected response format from Gemini API")
	}

	return string(text), nil
}

func (g *geminiClient) Close() {
	if g.client != nil {
		g.client.Close()
	}
}
