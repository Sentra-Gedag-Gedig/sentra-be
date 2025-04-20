package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"mime/multipart"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type ItfS3 interface {
	UploadFile(file *multipart.FileHeader) (string, error)
	PresignUrl(fileName string) (string, error)
	DeleteFile(fileName string) error
}

type s3Client struct {
	client     *s3.S3
	session    *session.Session
	bucketName string
}

func New() (ItfS3, error) {
	sess, err := newSession()
	if err != nil {
		return nil, err
	}

	return &s3Client{
		client:     s3.New(sess),
		session:    sess,
		bucketName: os.Getenv("AWS_BUCKET_NAME"),
	}, nil
}

func (s *s3Client) UploadFile(file *multipart.FileHeader) (string, error) {
	uploader := s3manager.NewUploader(s.session)

	uniqueFileName, err := generateUniqueFileName(file.Filename)
	if err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			fmt.Println("Failed to close file")
		}
	}(src)

	uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(uniqueFileName),
		Body:   src,
	})

	if err != nil {
		return "", err
	}

	return uploadOutput.Location, nil
}

func (s *s3Client) PresignUrl(fileUrl string) (string, error) {
	key := extractKeyFromS3Url(fileUrl)

	decodedKey, err := url.QueryUnescape(key)
	if err != nil {
		return "", fmt.Errorf("failed to decode S3 key: %w", err)
	}

	_, err = s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(decodedKey),
	})
	if err != nil {
		return "", fmt.Errorf("file does not exist: %w", err)
	}

	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(decodedKey),
	})

	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}

	return urlStr, nil
}

func extractKeyFromS3Url(fileUrl string) string {
	parts := strings.Split(fileUrl, ".com/")
	if len(parts) > 1 {
		return parts[1]
	}
	return fileUrl
}

func (s *s3Client) DeleteFile(fileName string) error {
	decodedFileName, err := url.QueryUnescape(fileName)
	if err != nil {
		return fmt.Errorf("failed to decode filename: %w", err)
	}

	_, err = s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(decodedFileName),
	})

	return err
}

func newSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})

	if err != nil {
		return nil, err
	}

	return sess, nil
}

func generateUniqueFileName(fileName string) (string, error) {
	uniqueFileName := fmt.Sprintf("%s-%s", strings.ReplaceAll(time.Now().String(), " ", ""), fileName)
	return uniqueFileName, nil
}
