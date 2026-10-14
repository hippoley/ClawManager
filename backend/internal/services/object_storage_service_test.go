package services

import (
	"testing"

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
