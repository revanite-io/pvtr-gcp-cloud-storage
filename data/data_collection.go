package data

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/privateerproj/privateer-sdk/config"
)

// Payload contains all GCS bucket data required for evaluation steps.
type Payload struct {
	Config *config.Config

	// Bucket versioning configuration
	Versioning *VersioningData

	// Server-side encryption configuration (CMEK)
	Encryption *EncryptionData

	// Bucket retention policy
	Retention *RetentionData

	// Soft delete policy
	SoftDelete *SoftDeleteData

	// Uniform bucket-level access configuration
	UniformAccess *UniformAccessData

	// Access logging configuration
	Logging *LoggingData

	// Resource metadata
	BucketName string
	Location   string
	Labels     map[string]string
}

// VersioningData contains GCS bucket versioning configuration.
type VersioningData struct {
	Enabled bool
}

// EncryptionData contains GCS bucket encryption configuration (CMEK).
type EncryptionData struct {
	DefaultKMSKeyName string // e.g. "projects/p/locations/l/keyRings/kr/cryptoKeys/k"
}

// RetentionData contains GCS bucket retention policy configuration.
type RetentionData struct {
	RetentionPeriodSeconds int64
	IsLocked               bool
}

// SoftDeleteData contains GCS bucket soft delete policy configuration.
type SoftDeleteData struct {
	RetentionDurationSeconds int64
}

// UniformAccessData contains GCS uniform bucket-level access configuration.
type UniformAccessData struct {
	Enabled bool
}

// LoggingData contains GCS bucket access logging configuration.
type LoggingData struct {
	Enabled         bool
	LogBucket       string
	LogPrefix       string
	LogBucketLabels map[string]string
}

// Loader is the SDK-compatible entrypoint.
func Loader(cfg *config.Config) (any, error) {
	return LoadWithOptions(cfg)
}

// LoadWithOptions is the testable entrypoint with functional options.
func LoadWithOptions(cfg *config.Config, opts ...Option) (any, error) {
	options := &loaderOptions{}
	for _, opt := range opts {
		opt(options)
	}

	bucketName := cfg.GetString("bucketname")
	if bucketName == "" {
		return nil, fmt.Errorf("required config 'bucketname' is not provided")
	}

	payload := Payload{
		Config:     cfg,
		BucketName: bucketName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client if not injected
	if options.storageClient == nil {
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS client: %v", err)
		}
		defer func() { _ = client.Close() }()
		options.storageClient = &gcsClient{client: client}
	}

	// Fetch bucket attributes (critical)
	attrs, err := options.storageClient.GetBucketAttrs(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bucket attributes: %v", err)
	}

	payload.Location = attrs.Location
	payload.Labels = attrs.Labels

	// Versioning
	payload.Versioning = &VersioningData{Enabled: attrs.VersioningEnabled}

	// Encryption (CMEK)
	payload.Encryption = fetchEncryption(attrs)

	// Retention policy
	payload.Retention = fetchRetention(attrs)

	// Soft delete policy
	payload.SoftDelete = fetchSoftDelete(attrs)

	// Uniform bucket-level access
	payload.UniformAccess = &UniformAccessData{
		Enabled: attrs.UniformBucketLevelAccess.Enabled,
	}

	// Access logging
	payload.Logging = fetchLogging(ctx, options.storageClient, attrs)

	return payload, nil
}

func fetchEncryption(attrs *storage.BucketAttrs) *EncryptionData {
	if attrs.Encryption == nil || attrs.Encryption.DefaultKMSKeyName == "" {
		return nil
	}
	return &EncryptionData{
		DefaultKMSKeyName: attrs.Encryption.DefaultKMSKeyName,
	}
}

func fetchRetention(attrs *storage.BucketAttrs) *RetentionData {
	if attrs.RetentionPolicy == nil {
		return nil
	}
	return &RetentionData{
		RetentionPeriodSeconds: int64(attrs.RetentionPolicy.RetentionPeriod.Seconds()),
		IsLocked:               attrs.RetentionPolicy.IsLocked,
	}
}

func fetchSoftDelete(attrs *storage.BucketAttrs) *SoftDeleteData {
	if attrs.SoftDeletePolicy == nil || attrs.SoftDeletePolicy.RetentionDuration <= 0 {
		return nil
	}
	return &SoftDeleteData{
		RetentionDurationSeconds: int64(attrs.SoftDeletePolicy.RetentionDuration.Seconds()),
	}
}

func fetchLogging(ctx context.Context, client StorageClient, attrs *storage.BucketAttrs) *LoggingData {
	logging := &LoggingData{}

	if attrs.Logging == nil || attrs.Logging.LogBucket == "" {
		return logging
	}

	logging.Enabled = true
	logging.LogBucket = attrs.Logging.LogBucket
	logging.LogPrefix = attrs.Logging.LogObjectPrefix

	// Fetch log bucket labels
	logAttrs, err := client.GetBucketAttrs(ctx, attrs.Logging.LogBucket)
	if err == nil && logAttrs != nil {
		logging.LogBucketLabels = logAttrs.Labels
	}

	return logging
}
