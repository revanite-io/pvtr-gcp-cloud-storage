package data

import (
	"context"

	orgpolicy "cloud.google.com/go/orgpolicy/apiv2"
	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"cloud.google.com/go/storage"
)

// StorageClient abstracts the GCS client for testing.
type StorageClient interface {
	GetBucketAttrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error)
}

// OrgPolicyClient abstracts the Organization Policy client for testing.
type OrgPolicyClient interface {
	GetEffectivePolicy(ctx context.Context, name string) (*orgpolicypb.Policy, error)
}

// loaderOptions holds optional dependencies for LoadWithOptions.
type loaderOptions struct {
	storageClient   StorageClient
	orgPolicyClient OrgPolicyClient
}

// Option configures the Loader.
type Option func(*loaderOptions)

// WithStorageClient overrides the default GCS client.
func WithStorageClient(c StorageClient) Option {
	return func(o *loaderOptions) { o.storageClient = c }
}

// WithOrgPolicyClient overrides the default Organization Policy client.
func WithOrgPolicyClient(c OrgPolicyClient) Option {
	return func(o *loaderOptions) { o.orgPolicyClient = c }
}

// gcsClient wraps the real GCS storage client.
type gcsClient struct {
	client *storage.Client
}

func (c *gcsClient) GetBucketAttrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error) {
	return c.client.Bucket(bucketName).Attrs(ctx)
}

// orgPolicyClient wraps the real Organization Policy client.
type orgPolicyClient struct {
	client *orgpolicy.Client
}

func (c *orgPolicyClient) GetEffectivePolicy(ctx context.Context, name string) (*orgpolicypb.Policy, error) {
	return c.client.GetEffectivePolicy(ctx, &orgpolicypb.GetEffectivePolicyRequest{Name: name})
}
