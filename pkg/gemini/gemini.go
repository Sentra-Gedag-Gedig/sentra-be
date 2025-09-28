package gemini

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type IGemini interface {
	AnalyzeImage(ctx context.Context, base64Image string, prompt string) (string, error)
	AnalyzeBinaryImage(ctx context.Context, binaryData []byte, prompt string) (string, error)
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
		modelName = "gemini-2.5-flash"
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

// detectMIMEType detects the MIME type from image data using magic bytes first
func detectMIMEType(data []byte) string {
	if len(data) < 4 {
		fmt.Printf("DEBUG detectMIMEType: Data too small (%d bytes), returning jpeg fallback\n", len(data))
		return "image/jpeg" // Safe fallback
	}

	fmt.Printf("DEBUG detectMIMEType: First 4 bytes: %X\n", data[:4])

	// Primary detection using magic bytes (more reliable)
	// JPEG magic bytes: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		fmt.Printf("DEBUG detectMIMEType: JPEG magic bytes detected, returning image/jpeg\n")
		return "image/jpeg"
	}
	
	// PNG magic bytes: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		fmt.Printf("DEBUG detectMIMEType: PNG magic bytes detected, returning image/png\n")
		return "image/png"
	}
	
	// GIF magic bytes: "GIF"
	if len(data) >= 3 && string(data[0:3]) == "GIF" {
		fmt.Printf("DEBUG detectMIMEType: GIF magic bytes detected, returning image/gif\n")
		return "image/gif"
	}
	
	// WebP magic bytes: "RIFF" ... "WEBP"
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		fmt.Printf("DEBUG detectMIMEType: WebP magic bytes detected, returning image/webp\n")
		return "image/webp"
	}

	// Fallback to http.DetectContentType but clean up the result
	rawMimeType := http.DetectContentType(data)
	fmt.Printf("DEBUG detectMIMEType: http.DetectContentType returned: '%s'\n", rawMimeType)
	
	mimeType := strings.ToLower(rawMimeType)
	fmt.Printf("DEBUG detectMIMEType: After toLowerCase: '%s'\n", mimeType)
	
	// Fix any duplicate prefixes
	if strings.HasPrefix(mimeType, "image/image/") {
		mimeType = strings.Replace(mimeType, "image/image/", "image/", 1)
		fmt.Printf("DEBUG detectMIMEType: Fixed duplicate prefix, now: '%s'\n", mimeType)
	}
	
	// Normalize known types
	switch {
	case strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg"):
		fmt.Printf("DEBUG detectMIMEType: Normalized to image/jpeg\n")
		return "image/jpeg"
	case strings.Contains(mimeType, "png"):
		fmt.Printf("DEBUG detectMIMEType: Normalized to image/png\n")
		return "image/png"
	case strings.Contains(mimeType, "gif"):
		fmt.Printf("DEBUG detectMIMEType: Normalized to image/gif\n")
		return "image/gif"
	case strings.Contains(mimeType, "webp"):
		fmt.Printf("DEBUG detectMIMEType: Normalized to image/webp\n")
		return "image/webp"
	default:
		fmt.Printf("DEBUG detectMIMEType: Unknown type '%s', falling back to image/jpeg\n", mimeType)
		// Safe fallback
		return "image/jpeg"
	}
}

// parseBase64Image handles both plain base64 and data URL formats
func parseBase64Image(base64Image string) ([]byte, string, error) {
	var imgData []byte
	var mimeType string
	var err error

	fmt.Printf("DEBUG parseBase64Image: Input length: %d\n", len(base64Image))
	fmt.Printf("DEBUG parseBase64Image: First 50 chars: '%.50s'\n", base64Image)

	// Check if it's a data URL (data:image/jpeg;base64,...)
	if strings.HasPrefix(base64Image, "data:") {
		fmt.Printf("DEBUG parseBase64Image: Processing as data URL\n")
		parts := strings.Split(base64Image, ",")
		if len(parts) != 2 {
			return nil, "", errors.New("invalid data URL format")
		}

		// Extract MIME type from data URL
		header := parts[0] // "data:image/jpeg;base64"
		fmt.Printf("DEBUG parseBase64Image: Data URL header: '%s'\n", header)
		headerParts := strings.Split(header, ";")
		if len(headerParts) >= 1 {
			mimeType = strings.TrimPrefix(headerParts[0], "data:")
			fmt.Printf("DEBUG parseBase64Image: Extracted MIME from data URL: '%s'\n", mimeType)
		}

		// Decode base64 part
		imgData, err = base64.StdEncoding.DecodeString(parts[1])
		fmt.Printf("DEBUG parseBase64Image: Data URL decode result - size: %d, error: %v\n", len(imgData), err)
	} else {
		fmt.Printf("DEBUG parseBase64Image: Processing as plain base64\n")
		// Plain base64 string
		imgData, err = base64.StdEncoding.DecodeString(base64Image)
		if err != nil {
			return nil, "", errors.New("invalid base64 image data")
		}
		
		fmt.Printf("DEBUG parseBase64Image: Plain base64 decode result - size: %d\n", len(imgData))
		
		// Auto-detect MIME type
		mimeType = detectMIMEType(imgData)
		fmt.Printf("DEBUG parseBase64Image: Auto-detected MIME type: '%s'\n", mimeType)
	}

	if err != nil {
		return nil, "", err
	}

	fmt.Printf("DEBUG parseBase64Image: Final result - MIME: '%s', size: %d\n", mimeType, len(imgData))
	return imgData, mimeType, nil
}

// AnalyzeImage handles base64 encoded images
func (g *geminiClient) AnalyzeImage(ctx context.Context, base64Image string, prompt string) (string, error) {
	imgData, mimeType, err := parseBase64Image(base64Image)
	if err != nil {
		return "", err
	}

	return g.analyzeImageData(ctx, imgData, mimeType, prompt)
}

// AnalyzeBinaryImage handles binary image data directly
func (g *geminiClient) AnalyzeBinaryImage(ctx context.Context, binaryData []byte, prompt string) (string, error) {
	mimeType := detectMIMEType(binaryData)
	return g.analyzeImageData(ctx, binaryData, mimeType, prompt)
}

// analyzeImageData is the core method that handles the actual API call
func (g *geminiClient) analyzeImageData(ctx context.Context, imgData []byte, mimeType string, prompt string) (string, error) {
	model := g.client.GenerativeModel(g.modelName)

	if prompt == "" {
		prompt = "Analyze this image and provide details in JSON format."
	}

	// Debug logging - remove this after fixing
	fmt.Printf("DEBUG: Image size: %d bytes\n", len(imgData))
	fmt.Printf("DEBUG: First 4 bytes: %X\n", imgData[:min(4, len(imgData))])
	fmt.Printf("DEBUG: Detected MIME type: %s\n", mimeType)
	fmt.Printf("DEBUG: Using model: %s\n", g.modelName)

	// WORKAROUND: Force JPEG MIME type to prevent duplicate prefix issue
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		mimeType = "image/jpeg"
		fmt.Printf("DEBUG: Forced MIME type to: %s\n", mimeType)
	}

	// Try different approaches to avoid the duplicate prefix issue
	var img genai.Part
	
	// Method 1: Try with explicit MIME type
	img = genai.ImageData(mimeType, imgData)
	
	res, err := model.GenerateContent(ctx, genai.Text(prompt), img)
	if err != nil {
		// If still failing with duplicate MIME type, try alternative approaches
		if strings.Contains(err.Error(), "image/image/") {
			fmt.Printf("DEBUG: Attempting workaround for duplicate MIME type issue\n")
			
			// Method 2: Try with just "jpeg" instead of "image/jpeg"
			img = genai.ImageData("jpeg", imgData)
			res, err = model.GenerateContent(ctx, genai.Text(prompt), img)
			
			if err != nil && strings.Contains(err.Error(), "Unsupported MIME type") {
				// Method 3: Try creating a blob without MIME type specification
				fmt.Printf("DEBUG: Trying blob approach without MIME type\n")
				img = genai.Blob{MIMEType: "image/jpeg", Data: imgData}
				res, err = model.GenerateContent(ctx, genai.Text(prompt), img)
			}
		}
		
		if err != nil {
			fmt.Printf("DEBUG: All methods failed. Final error: %v\n", err)
			return "", err
		}
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

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *geminiClient) Close() {
	if g.client != nil {
		g.client.Close()
	}
}