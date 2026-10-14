package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"clawreef/internal/config"
	"github.com/minio/minio-go/v7"
)

func TestBucketLookupMode(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.ObjectStorageConfig
		want minio.BucketLookupType
	}{
		{"legacy default", config.ObjectStorageConfig{}, minio.BucketLookupPath},
		{"legacy forced path", config.ObjectStorageConfig{ForcePathStyle: true}, minio.BucketLookupPath},
		{"explicit path", config.ObjectStorageConfig{BucketLookupMode: "path"}, minio.BucketLookupPath},
		{"explicit auto", config.ObjectStorageConfig{BucketLookupMode: "auto"}, minio.BucketLookupAuto},
		{"explicit dns", config.ObjectStorageConfig{BucketLookupMode: "dns"}, minio.BucketLookupDNS},
		{"trim and case", config.ObjectStorageConfig{BucketLookupMode: " AUTO "}, minio.BucketLookupAuto},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bucketLookupMode(tc.cfg); got != tc.want {
				t.Fatalf("lookup mode = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestObjectStorageBucketLookupValidation(t *testing.T) {
	cases := []struct {
		name      string
		cfg       config.ObjectStorageConfig
		wantError bool
	}{
		{"invalid", config.ObjectStorageConfig{Endpoint: "localhost:9000", BucketLookupMode: "bad"}, true},
		{"invalid even with local fallback", config.ObjectStorageConfig{BucketLookupMode: "bad"}, true},
		{"conflict even with local fallback", config.ObjectStorageConfig{ForcePathStyle: true, BucketLookupMode: "dns"}, true},
		{"local fallback without mode", config.ObjectStorageConfig{}, false},
		{"conflicting", config.ObjectStorageConfig{Endpoint: "localhost:9000", ForcePathStyle: true, BucketLookupMode: "dns"}, true},
		{"conflicting auto", config.ObjectStorageConfig{Endpoint: "localhost:9000", ForcePathStyle: true, BucketLookupMode: "auto"}, true},
		{"explicit path consistent", config.ObjectStorageConfig{Endpoint: "localhost:9000", ForcePathStyle: true, BucketLookupMode: "path"}, false},
		{"valid auto", config.ObjectStorageConfig{Endpoint: "localhost:9000", BucketLookupMode: "auto"}, false},
		{"valid dns", config.ObjectStorageConfig{Endpoint: "localhost:9000", BucketLookupMode: "dns"}, false},
		{"valid default", config.ObjectStorageConfig{Endpoint: "localhost:9000"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewObjectStorageService(tc.cfg)
			if (err != nil) != tc.wantError {
				t.Fatalf("NewObjectStorageService error = %v; wantError = %v", err, tc.wantError)
			}
		})
	}
}

// Verify explicit modes with exact address expectations. Auto is intentionally
// implementation-defined by the MinIO SDK and must not be treated as a
// contractual path-style default for non-AWS endpoints.
func TestObjectStorageSDKPresignedAddressing(t *testing.T) {
	cases := []struct {
		name     string
		mode     string
		wantHost string
		wantPath string
	}{
		{"legacy path default", "", "s3.example.test", "/fixture-bucket/receipt.json"},
		{"explicit path", "path", "s3.example.test", "/fixture-bucket/receipt.json"},
		{"explicit dns", "dns", "fixture-bucket.s3.example.test", "/receipt.json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := NewObjectStorageService(config.ObjectStorageConfig{
				Endpoint: "s3.example.test",
				Region: "us-east-1",
				AccessKey: "test-access",
				SecretKey: "test-secret",
				Bucket: "fixture-bucket",
				UseSSL: true,
				BucketLookupMode: tc.mode,
			})
			if err != nil {
				t.Fatalf("new client: %v", err)
			}
			client := svc.(*objectStorageService).minioClient
			u, err := client.PresignedGetObject(context.Background(), "fixture-bucket", "receipt.json", time.Minute, nil)
			if err != nil {
				t.Fatalf("presign with SDK: %v", err)
			}
			if u.Host != tc.wantHost || u.Path != tc.wantPath {
				t.Fatalf("presigned address = %s%s, want %s%s", u.Host, u.Path, tc.wantHost, tc.wantPath)
			}
			if !strings.HasPrefix(u.Query().Get("X-Amz-Algorithm"), "AWS4-") {
				t.Fatalf("unexpected presign algorithm: %s", u.Query().Get("X-Amz-Algorithm"))
			}
		})
	}
}

func TestObjectStorageSDKAutoGeneratesSignedURL(t *testing.T) {
	svc, err := NewObjectStorageService(config.ObjectStorageConfig{
		Endpoint: "s3.example.test",
		Region: "us-east-1",
		AccessKey: "test-access",
		SecretKey: "test-secret",
		Bucket: "fixture-bucket",
		UseSSL: true,
		BucketLookupMode: "auto",
	})
	if err != nil {
		t.Fatal(err)
	}
	client := svc.(*objectStorageService).minioClient
	u, err := client.PresignedGetObject(context.Background(), "fixture-bucket", "receipt.json", time.Minute, nil)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host == "" || u.Path == "" || !strings.HasPrefix(u.Query().Get("X-Amz-Algorithm"), "AWS4-") {
		t.Fatalf("invalid SDK auto-mode signed URL: %s", u.String())
	}
}
