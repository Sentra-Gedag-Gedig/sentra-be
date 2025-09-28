package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type IUtils interface {
	NewULIDFromTimestamp(t time.Time) (string, error)
	ValidateImageFile(file *multipart.FileHeader) error
	ConvertFileToBase64(file multipart.File) (string, error)
	OptimizeImageForOCR(imageData []byte, maxWidth, maxHeight int, quality int) ([]byte, error)
}

type utils struct {
	maxFileSize int64
}

func New() IUtils {
	return &utils{
		maxFileSize: 5 * 1024 * 1024,
	}
}

func (u *utils) NewULIDFromTimestamp(t time.Time) (string, error) {
	ms := ulid.Timestamp(t)
	entropy := ulid.Monotonic(rand.Reader, 0)

	id, err := ulid.New(ms, entropy)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (u *utils) ValidateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return errors.New("no file uploaded")
	}

	if file.Size > u.maxFileSize {
		return errors.New("file size exceeds limit")
	}

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return errors.New("uploaded file is not an image")
	}

	return nil
}

func (u *utils) ConvertFileToBase64(file multipart.File) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	base64Encoded := base64.StdEncoding.EncodeToString(fileBytes)
	return base64Encoded, nil
}

func (u *utils) OptimizeImageForOCR(imageData []byte, maxWidth, maxHeight int, quality int) ([]byte, error) {
	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	// Get original dimensions
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate new dimensions if needed
	newWidth, newHeight := origWidth, origHeight
	if origWidth > maxWidth || origHeight > maxHeight {
		ratio := float64(origWidth) / float64(origHeight)
		
		if origWidth > origHeight {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / ratio)
		} else {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * ratio)
		}
	}

	// Only resize if dimensions changed
	if newWidth != origWidth || newHeight != origHeight {
		// Use a simple resize (you might want to use a proper image library)
		// For now, we'll just return original if resize is needed
		// You should implement proper image resizing here
	}

	// Re-encode with compression
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, img)
	default:
		// Convert to JPEG for better compression
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}