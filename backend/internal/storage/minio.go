package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nakamura/chatwoot-go/internal/config"
)

type MinioService struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewMinioService(cfg *config.Config) (*MinioService, error) {
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	service := &MinioService{
		client:     minioClient,
		bucketName: "nakamura-uploads",
		endpoint:   cfg.MinioEndpoint,
		useSSL:     cfg.MinioUseSSL,
	}

	if err := service.EnsureBucket(context.Background()); err != nil {
		log.Printf("Warning: Failed to ensure bucket exists: %v", err)
	}

	// Set bucket policy to public read for simplified access inside container network
	// Note: This policy allows public read access. Be careful in production.
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::` + service.bucketName + `/*"]
			}
		]
	}`
	_ = service.client.SetBucketPolicy(context.Background(), service.bucketName, policy)

	return service, nil
}

func (s *MinioService) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (s *MinioService) UploadFile(ctx context.Context, file io.Reader, fileSize int64, originalName string, contentType string) (string, error) {
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".bin"
	}
	// Generate unique filename
	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Logic to determine public URL
	publicEndpoint := s.endpoint
	// If running in docker internal network, map to localhost for external access
	if strings.Contains(s.endpoint, "minio") {
		publicEndpoint = "localhost:9000"
	}

	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}

	url := fmt.Sprintf("%s://%s/%s/%s", protocol, publicEndpoint, s.bucketName, objectName)
	return url, nil
}
