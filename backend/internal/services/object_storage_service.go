package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"clawreef/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorageService interface {
	PutObject(ctx context.Context, objectKey string, body []byte, contentType string) error
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
}

type objectStorageService struct {
	minioClient *minio.Client
	bucket      string
	basePath    string
	localPath   string
}

// bucketLookupMode preserves the historical path-style default for existing
// deployments. Users can explicitly select "auto" or "dns" for providers
// requiring virtual-host-style addressing (e.g. Tencent COS).
func bucketLookupMode(cfg config.ObjectStorageConfig) minio.BucketLookupType {
    switch strings.ToLower(strings.TrimSpace(cfg.BucketLookupMode)) {
    case "auto":
        return minio.BucketLookupAuto
    case "dns":
        return minio.BucketLookupDNS
    default:
        return minio.BucketLookupPath
    }
}

func NewObjectStorageService(cfg config.ObjectStorageConfig) (ObjectStorageService, error) {
	service := &objectStorageService{
		bucket:    strings.TrimSpace(cfg.Bucket),
		basePath:  strings.Trim(strings.TrimSpace(cfg.BasePath), "/"),
		localPath: strings.TrimSpace(cfg.LocalFallback),
	}
	if service.localPath == "" {
		service.localPath = ".data/object-storage"
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return service, nil
	}

	if mode := strings.ToLower(strings.TrimSpace(cfg.BucketLookupMode)); mode != "" && mode != "path" && mode != "auto" && mode != "dns" {
		return nil, fmt.Errorf("unsupported object storage bucketLookupMode %q (want path, auto, or dns)", cfg.BucketLookupMode)
	}
	if cfg.ForcePathStyle && cfg.BucketLookupMode != "" && !strings.EqualFold(strings.TrimSpace(cfg.BucketLookupMode), "path") {
		return nil, fmt.Errorf("forcePathStyle conflicts with bucketLookupMode %q", cfg.BucketLookupMode)
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: bucketLookupMode(cfg),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize object storage client: %w", err)
	}
	service.minioClient = client
	return service, nil
}

func (s *objectStorageService) PutObject(ctx context.Context, objectKey string, body []byte, contentType string) error {
	if s.minioClient == nil {
		target := filepath.Join(s.localPath, filepath.FromSlash(s.resolveObjectKey(objectKey)))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("failed to prepare local object storage directory: %w", err)
		}
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return fmt.Errorf("failed to write local object storage object: %w", err)
		}
		return nil
	}

	exists, err := s.minioClient.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check object storage bucket: %w", err)
	}
	if !exists {
		if err := s.minioClient.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create object storage bucket: %w", err)
		}
	}
	_, err = s.minioClient.PutObject(ctx, s.bucket, s.resolveObjectKey(objectKey), bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

func (s *objectStorageService) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	if s.minioClient == nil {
		target := filepath.Join(s.localPath, filepath.FromSlash(s.resolveObjectKey(objectKey)))
		content, err := os.ReadFile(target)
		if err != nil {
			return nil, fmt.Errorf("failed to read local object storage object: %w", err)
		}
		return content, nil
	}

	reader, err := s.minioClient.GetObject(ctx, s.bucket, s.resolveObjectKey(objectKey), minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to open object: %w", err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	return content, nil
}

func (s *objectStorageService) resolveObjectKey(objectKey string) string {
	objectKey = strings.Trim(strings.TrimSpace(objectKey), "/")
	if s.basePath == "" {
		return objectKey
	}
	return s.basePath + "/" + objectKey
}
