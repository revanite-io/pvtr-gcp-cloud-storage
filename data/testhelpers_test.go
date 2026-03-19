package data

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
)

// mockStorageClient satisfies StorageClient for tests.
type mockStorageClient struct {
	attrsResp map[string]*storage.BucketAttrs
	attrsErr  map[string]error
	// Default response/error when no bucket-specific entry exists
	defaultResp *storage.BucketAttrs
	defaultErr  error
}

func (m *mockStorageClient) GetBucketAttrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error) {
	if m.attrsErr != nil {
		if err, ok := m.attrsErr[bucketName]; ok {
			return nil, err
		}
	}
	if m.attrsResp != nil {
		if resp, ok := m.attrsResp[bucketName]; ok {
			return resp, nil
		}
	}
	if m.defaultErr != nil {
		return nil, m.defaultErr
	}
	if m.defaultResp != nil {
		return m.defaultResp, nil
	}
	// Return minimal valid attrs
	return &storage.BucketAttrs{
		Location: "US",
	}, nil
}

// allMockOptions returns functional options with a minimal mock client.
func allMockOptions() []Option {
	return []Option{
		WithStorageClient(&mockStorageClient{}),
	}
}

// newMockAttrs creates a *storage.BucketAttrs with common defaults for testing.
func newMockAttrs() *storage.BucketAttrs {
	return &storage.BucketAttrs{
		Location:          "US-CENTRAL1",
		VersioningEnabled: true,
		Encryption: &storage.BucketEncryption{
			DefaultKMSKeyName: "projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key",
		},
		RetentionPolicy: &storage.RetentionPolicy{
			RetentionPeriod: 86400 * time.Second, // 1 day
			IsLocked:        true,
		},
		SoftDeletePolicy: &storage.SoftDeletePolicy{
			RetentionDuration: 7 * 24 * time.Hour, // 7 days
		},
		UniformBucketLevelAccess: storage.UniformBucketLevelAccess{
			Enabled: true,
		},
		Labels: map[string]string{
			"environment": "test",
		},
		Logging: &storage.BucketLogging{
			LogBucket:       "my-log-bucket",
			LogObjectPrefix: "access-logs/",
		},
	}
}
