package data

import (
	"context"

	"cloud.google.com/go/storage"
)

// StorageClient abstracts the GCS client for testing.
type StorageClient interface {
	GetBucketAttrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error)
}

// loaderOptions holds optional dependencies for LoadWithOptions.
type loaderOptions struct {
	storageClient StorageClient
}

// Option configures the Loader.
type Option func(*loaderOptions)

// WithStorageClient overrides the default GCS client.
func WithStorageClient(c StorageClient) Option {
	return func(o *loaderOptions) { o.storageClient = c }
}

// gcsClient wraps the real GCS storage client.
type gcsClient struct {
	client *storage.Client
}

func (c *gcsClient) GetBucketAttrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error) {
	return c.client.Bucket(bucketName).Attrs(ctx)
}
