package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
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
