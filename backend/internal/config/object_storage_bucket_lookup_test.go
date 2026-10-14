package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestObjectStorageBucketLookupYAMLCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		mode       string
		forcedPath bool
	}{
		{"legacy omitted", "endpoint: localhost:9000\n", "", false},
		{"legacy forced path", "forcePathStyle: true\n", "", true},
		{"explicit path", "bucketLookupMode: path\n", "path", false},
		{"explicit auto", "bucketLookupMode: auto\n", "auto", false},
		{"explicit dns", "bucketLookupMode: dns\n", "dns", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg ObjectStorageConfig
			if err := yaml.Unmarshal([]byte(tt.yaml), &cfg); err != nil {
				t.Fatalf("unmarshal objectStorage: %v", err)
			}
			if cfg.BucketLookupMode != tt.mode || cfg.ForcePathStyle != tt.forcedPath {
				t.Fatalf("decoded mode=%q forcePathStyle=%v; want mode=%q forcePathStyle=%v", cfg.BucketLookupMode, cfg.ForcePathStyle, tt.mode, tt.forcedPath)
			}
		})
	}
}
