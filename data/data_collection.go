package data

import (
	"context"
	"fmt"
	"time"

	orgpolicy "cloud.google.com/go/orgpolicy/apiv2"
	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"cloud.google.com/go/storage"
	"github.com/privateerproj/privateer-sdk/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Organization Policy constraints relevant to this bucket's project.
	// Nil when the effective policy could not be fetched (e.g. missing
	// permission), which is distinct from a fetched-but-unrestricted policy.
	OrgPolicy *OrgPolicyData

	// Sampled object generation evidence for the versioning checks.
	// Nil when versioning is disabled or the sample could not be listed.
	ObjectVersions *ObjectVersionsData

	// Resource metadata
	BucketName string
	Location   string
	Labels     map[string]string
}

// ObjectVersionsData summarizes a sample of the bucket's object generations.
// Counts are derived from a bounded listing with versions included, so they
// are evidence of observed behavior, not a full inventory.
type ObjectVersionsData struct {
	SampledCount        int // total entries sampled (live + noncurrent)
	NoncurrentCount     int // entries that are noncurrent generations
	ModifiedWithHistory int // object names with a live generation plus noncurrent history
	DeletedRetained     int // object names with only noncurrent generations (deleted but retained)
}

// OrgPolicyData contains effective Organization Policy constraints.
type OrgPolicyData struct {
	// Effective policy for constraints/gcp.restrictCmekCryptoKeyProjects
	RestrictCmekCryptoKeyProjects *ConstraintPolicy
}

// ConstraintPolicy summarizes an effective list-constraint policy.
type ConstraintPolicy struct {
	AllowAll      bool
	DenyAll       bool
	AllowedValues []string
	DeniedValues  []string
}

// Restricted reports whether the constraint limits which values are allowed.
// A policy that allows everything (explicitly or by default) is unrestricted.
func (c *ConstraintPolicy) Restricted() bool {
	if c == nil || c.AllowAll {
		return false
	}
	return c.DenyAll || len(c.AllowedValues) > 0 || len(c.DeniedValues) > 0
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

	// Create real clients if none were injected. The Organization Policy
	// client is optional: without it OrgPolicy stays nil and the affected
	// steps report NeedsReview instead of a verdict.
	if options.storageClient == nil {
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS client: %v", err)
		}
		defer func() { _ = client.Close() }()
		options.storageClient = &gcsClient{client: client}

		if options.orgPolicyClient == nil {
			if opClient, err := orgpolicy.NewClient(ctx); err == nil {
				defer func() { _ = opClient.Close() }()
				options.orgPolicyClient = &orgPolicyClient{client: opClient}
			}
		}
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

	// Organization Policy (non-critical: nil when unavailable)
	payload.OrgPolicy = fetchOrgPolicy(ctx, options.orgPolicyClient, attrs.ProjectNumber)

	// Object generation sample (non-critical: nil when unavailable)
	if attrs.VersioningEnabled {
		payload.ObjectVersions = fetchObjectVersions(ctx, options.storageClient, bucketName)
	}

	return payload, nil
}

// objectVersionSampleLimit bounds the generation listing so large buckets do
// not stall the loader; the sample only needs to observe versioning behavior.
const objectVersionSampleLimit = 1000

func fetchObjectVersions(ctx context.Context, client StorageClient, bucketName string) *ObjectVersionsData {
	entries, err := client.ListObjectVersions(ctx, bucketName, objectVersionSampleLimit)
	if err != nil {
		return nil
	}

	type nameHistory struct {
		live       bool
		noncurrent int
	}
	histories := map[string]*nameHistory{}

	sample := &ObjectVersionsData{SampledCount: len(entries)}
	for _, entry := range entries {
		history := histories[entry.Name]
		if history == nil {
			history = &nameHistory{}
			histories[entry.Name] = history
		}
		// A noncurrent generation has its deletion (supersession) time set.
		if entry.Deleted.IsZero() {
			history.live = true
		} else {
			history.noncurrent++
			sample.NoncurrentCount++
		}
	}

	for _, history := range histories {
		if history.noncurrent == 0 {
			continue
		}
		if history.live {
			sample.ModifiedWithHistory++
		} else {
			sample.DeletedRetained++
		}
	}
	return sample
}

func fetchOrgPolicy(ctx context.Context, client OrgPolicyClient, projectNumber uint64) *OrgPolicyData {
	if client == nil || projectNumber == 0 {
		return nil
	}

	name := fmt.Sprintf("projects/%d/policies/gcp.restrictCmekCryptoKeyProjects", projectNumber)
	policy, err := client.GetEffectivePolicy(ctx, name)
	if err != nil {
		// A constraint that has never been configured anywhere in the
		// hierarchy comes back NotFound; that means its default applies,
		// which for restrictCmekCryptoKeyProjects is allow-all.
		if status.Code(err) == codes.NotFound {
			return &OrgPolicyData{
				RestrictCmekCryptoKeyProjects: &ConstraintPolicy{AllowAll: true},
			}
		}
		return nil
	}

	return &OrgPolicyData{
		RestrictCmekCryptoKeyProjects: summarizePolicy(policy),
	}
}

// summarizePolicy flattens an effective list-constraint policy's rules into
// a ConstraintPolicy. An effective policy with no rules applies the
// constraint's default, which for restrictCmekCryptoKeyProjects is allow-all.
func summarizePolicy(policy *orgpolicypb.Policy) *ConstraintPolicy {
	summary := &ConstraintPolicy{}
	if policy == nil || policy.GetSpec() == nil || len(policy.GetSpec().GetRules()) == 0 {
		summary.AllowAll = true
		return summary
	}

	for _, rule := range policy.GetSpec().GetRules() {
		if rule.GetAllowAll() {
			summary.AllowAll = true
		}
		if rule.GetDenyAll() {
			summary.DenyAll = true
		}
		if values := rule.GetValues(); values != nil {
			summary.AllowedValues = append(summary.AllowedValues, values.GetAllowedValues()...)
			summary.DeniedValues = append(summary.DeniedValues, values.GetDeniedValues()...)
		}
	}
	return summary
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
