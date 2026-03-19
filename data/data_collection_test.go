package data

import (
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/privateerproj/privateer-sdk/config"
)

func testConfig(bucketName string) *config.Config {
	return &config.Config{
		Vars: map[string]interface{}{
			"bucketname": bucketName,
		},
	}
}

// --- LoadWithOptions ---

func TestLoadWithOptions_MissingBucketName(t *testing.T) {
	cfg := &config.Config{Vars: map[string]interface{}{}}
	_, err := LoadWithOptions(cfg, allMockOptions()...)
	if err == nil {
		t.Fatal("expected error for missing bucketname")
	}
}

func TestLoadWithOptions_MinimalSuccess(t *testing.T) {
	result, err := LoadWithOptions(testConfig("my-bucket"), allMockOptions()...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload, ok := result.(Payload)
	if !ok {
		t.Fatalf("expected Payload, got %T", result)
	}

	if payload.BucketName != "my-bucket" {
		t.Errorf("BucketName = %q, want %q", payload.BucketName, "my-bucket")
	}

	if payload.Location != "US" {
		t.Errorf("Location = %q, want %q", payload.Location, "US")
	}
}

func TestLoadWithOptions_AttrsError(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultErr: errors.New("access denied"),
		}),
	}
	_, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err == nil {
		t.Fatal("expected error when attrs fetch fails")
	}
}

func TestLoadWithOptions_FullAttrs(t *testing.T) {
	attrs := newMockAttrs()
	logBucketAttrs := &storage.BucketAttrs{
		Labels: map[string]string{"sensitivity": "high"},
	}

	opts := []Option{
		WithStorageClient(&mockStorageClient{
			attrsResp: map[string]*storage.BucketAttrs{
				"my-bucket":     attrs,
				"my-log-bucket": logBucketAttrs,
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Location != "US-CENTRAL1" {
		t.Errorf("Location = %q, want %q", payload.Location, "US-CENTRAL1")
	}
}

// --- Versioning ---

func TestLoadWithOptions_Versioning(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				VersioningEnabled: true,
				Location:          "US",
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Versioning == nil {
		t.Fatal("Versioning is nil")
	}
	if !payload.Versioning.Enabled {
		t.Error("Versioning.Enabled should be true")
	}
}

func TestLoadWithOptions_VersioningDisabled(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				VersioningEnabled: false,
				Location:          "US",
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Versioning == nil {
		t.Fatal("Versioning is nil")
	}
	if payload.Versioning.Enabled {
		t.Error("Versioning.Enabled should be false")
	}
}

// --- Encryption ---

func TestLoadWithOptions_Encryption(t *testing.T) {
	kmsKey := "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key"
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				Location: "US",
				Encryption: &storage.BucketEncryption{
					DefaultKMSKeyName: kmsKey,
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Encryption == nil {
		t.Fatal("Encryption is nil")
	}
	if payload.Encryption.DefaultKMSKeyName != kmsKey {
		t.Errorf("DefaultKMSKeyName = %q, want %q", payload.Encryption.DefaultKMSKeyName, kmsKey)
	}
}

func TestLoadWithOptions_EncryptionNoCMEK(t *testing.T) {
	result, err := LoadWithOptions(testConfig("my-bucket"), allMockOptions()...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Encryption != nil {
		t.Error("expected nil Encryption when no CMEK configured")
	}
}

// --- Retention ---

func TestLoadWithOptions_Retention(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				Location: "US",
				RetentionPolicy: &storage.RetentionPolicy{
					RetentionPeriod: 86400 * time.Second,
					IsLocked:        true,
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Retention == nil {
		t.Fatal("Retention is nil")
	}
	if payload.Retention.RetentionPeriodSeconds != 86400 {
		t.Errorf("RetentionPeriodSeconds = %d, want 86400", payload.Retention.RetentionPeriodSeconds)
	}
	if !payload.Retention.IsLocked {
		t.Error("IsLocked should be true")
	}
}

func TestLoadWithOptions_RetentionNone(t *testing.T) {
	result, err := LoadWithOptions(testConfig("my-bucket"), allMockOptions()...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Retention != nil {
		t.Error("expected nil Retention when no policy configured")
	}
}

// --- SoftDelete ---

func TestLoadWithOptions_SoftDelete(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				Location: "US",
				SoftDeletePolicy: &storage.SoftDeletePolicy{
					RetentionDuration: 7 * 24 * time.Hour,
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.SoftDelete == nil {
		t.Fatal("SoftDelete is nil")
	}
	expected := int64((7 * 24 * time.Hour).Seconds())
	if payload.SoftDelete.RetentionDurationSeconds != expected {
		t.Errorf("RetentionDurationSeconds = %d, want %d", payload.SoftDelete.RetentionDurationSeconds, expected)
	}
}

func TestLoadWithOptions_SoftDeleteNone(t *testing.T) {
	result, err := LoadWithOptions(testConfig("my-bucket"), allMockOptions()...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.SoftDelete != nil {
		t.Error("expected nil SoftDelete when no policy configured")
	}
}

// --- UniformAccess ---

func TestLoadWithOptions_UniformAccess(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				Location: "US",
				UniformBucketLevelAccess: storage.UniformBucketLevelAccess{
					Enabled: true,
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.UniformAccess == nil {
		t.Fatal("UniformAccess is nil")
	}
	if !payload.UniformAccess.Enabled {
		t.Error("UniformAccess.Enabled should be true")
	}
}

// --- Logging ---

func TestLoadWithOptions_Logging(t *testing.T) {
	logBucket := "my-log-bucket"
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			attrsResp: map[string]*storage.BucketAttrs{
				"my-bucket": {
					Location: "US",
					Logging: &storage.BucketLogging{
						LogBucket:       logBucket,
						LogObjectPrefix: "access-logs/",
					},
				},
				logBucket: {
					Labels: map[string]string{
						"sensitivity": "high",
						"purpose":     "access-logs",
					},
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Logging == nil {
		t.Fatal("Logging is nil")
	}
	if !payload.Logging.Enabled {
		t.Error("Logging.Enabled should be true")
	}
	if payload.Logging.LogBucket != logBucket {
		t.Errorf("LogBucket = %q, want %q", payload.Logging.LogBucket, logBucket)
	}
	if payload.Logging.LogBucketLabels == nil {
		t.Fatal("LogBucketLabels is nil")
	}
	if payload.Logging.LogBucketLabels["sensitivity"] != "high" {
		t.Errorf("sensitivity label = %q, want %q", payload.Logging.LogBucketLabels["sensitivity"], "high")
	}
}

func TestLoadWithOptions_LoggingDisabled(t *testing.T) {
	result, err := LoadWithOptions(testConfig("my-bucket"), allMockOptions()...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Logging == nil {
		t.Fatal("Logging is nil")
	}
	if payload.Logging.Enabled {
		t.Error("Logging.Enabled should be false when no logging configured")
	}
}

func TestLoadWithOptions_LoggingLogBucketError(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			attrsResp: map[string]*storage.BucketAttrs{
				"my-bucket": {
					Location: "US",
					Logging: &storage.BucketLogging{
						LogBucket: "missing-log-bucket",
					},
				},
			},
			attrsErr: map[string]error{
				"missing-log-bucket": errors.New("not found"),
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Logging == nil {
		t.Fatal("Logging is nil")
	}
	if !payload.Logging.Enabled {
		t.Error("Logging.Enabled should be true even if log bucket labels fail")
	}
	if payload.Logging.LogBucketLabels != nil {
		t.Error("LogBucketLabels should be nil when log bucket fetch fails")
	}
}

// --- Labels ---

func TestLoadWithOptions_Labels(t *testing.T) {
	opts := []Option{
		WithStorageClient(&mockStorageClient{
			defaultResp: &storage.BucketAttrs{
				Location: "US",
				Labels: map[string]string{
					"env":     "test",
					"project": "pvtr",
				},
			},
		}),
	}
	result, err := LoadWithOptions(testConfig("my-bucket"), opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := result.(Payload)

	if payload.Labels == nil {
		t.Fatal("Labels is nil")
	}
	if payload.Labels["env"] != "test" {
		t.Errorf("Labels[env] = %q, want %q", payload.Labels["env"], "test")
	}
}
