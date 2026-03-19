package data

import (
	"github.com/gemaraproj/go-gemara"

	"github.com/revanite-io/pvtr-gcp-cloud-storage/evaluation_plans/reusable_steps"
)

// --- CN01: KMS Key Trust ---

// PreventUntrustedKmsKeysForBucketRead verifies that CMEK is configured for the bucket,
// ensuring reads use a customer-managed Cloud KMS key.
func PreventUntrustedKmsKeysForBucketRead(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Encryption == nil {
		return gemara.Failed, "Bucket is not using customer-managed encryption (CMEK). Google-managed encryption does not allow KMS key trust enforcement", confidence
	}

	// In GCS, CMEK sets the default encryption key for the bucket.
	// Reads decrypt using the key the object was encrypted with.
	// To restrict which keys can be used, an Organization Policy constraint
	// (constraints/gcp.restrictCmekCryptoKeyProjects) is required.
	return gemara.NeedsReview, "CMEK is configured with key: " + payload.Encryption.DefaultKMSKeyName + ". Manual verification required to confirm Organization Policy restricts KMS key usage to trusted keys only", confidence
}

// PreventUntrustedKmsKeysForObjectRead verifies that object read requests use a trusted KMS key.
func PreventUntrustedKmsKeysForObjectRead(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return PreventUntrustedKmsKeysForBucketRead(payloadData)
}

// PreventUntrustedKmsKeysForBucketWrite verifies that write requests use a trusted KMS key.
func PreventUntrustedKmsKeysForBucketWrite(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Encryption == nil {
		return gemara.Failed, "Bucket is not using customer-managed encryption (CMEK). Writes will use Google-managed encryption, so KMS key trust cannot be enforced", confidence
	}

	// GCS CMEK sets the default key, but per-object overrides are possible
	// unless restricted by Organization Policy.
	return gemara.NeedsReview, "CMEK is configured with key: " + payload.Encryption.DefaultKMSKeyName + ". Manual verification required to confirm Organization Policy prevents writes with untrusted KMS keys", confidence
}

// PreventUntrustedKmsKeysForObjectWrite verifies that object write requests use a trusted KMS key.
func PreventUntrustedKmsKeysForObjectWrite(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return PreventUntrustedKmsKeysForBucketWrite(payloadData)
}

// --- CN02: Uniform Bucket-Level Access ---

// UniformBucketLevelAccessEnabled verifies that uniform bucket-level access is enabled,
// ensuring consistent IAM permissions across all objects.
func UniformBucketLevelAccessEnabled(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.UniformAccess == nil {
		return gemara.Unknown, "Uniform bucket-level access configuration not available", confidence
	}

	if !payload.UniformAccess.Enabled {
		return gemara.Failed, "Uniform bucket-level access is not enabled. Object-level ACLs may allow inconsistent permissions", confidence
	}

	return gemara.Passed, "Uniform bucket-level access is enabled, enforcing consistent IAM permissions across all objects", confidence
}

// UniformBucketLevelAccessEnabledForDenial verifies uniform access for the denial case.
func UniformBucketLevelAccessEnabledForDenial(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	return UniformBucketLevelAccessEnabled(payloadData)
}

// --- CN03: Bucket Deletion Recovery ---

// BucketDeletionRecoverable verifies that the bucket has mechanisms for recovery after deletion.
func BucketDeletionRecoverable(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	hasRetention := payload.Retention != nil && payload.Retention.RetentionPeriodSeconds > 0
	hasSoftDelete := payload.SoftDelete != nil && payload.SoftDelete.RetentionDurationSeconds > 0
	hasVersioning := payload.Versioning != nil && payload.Versioning.Enabled

	if !hasVersioning {
		return gemara.Failed, "Versioning is not enabled, so objects cannot be recovered after deletion", confidence
	}

	if !hasRetention && !hasSoftDelete {
		return gemara.Failed, "Neither retention policy nor soft delete policy is configured. Objects may not be recoverable after deletion", confidence
	}

	if hasRetention && hasSoftDelete {
		return gemara.Passed, "Versioning, retention policy, and soft delete are all configured, providing robust deletion recovery", confidence
	}

	if hasRetention {
		return gemara.Passed, "Versioning and retention policy are configured, allowing object recovery after deletion", confidence
	}

	return gemara.Passed, "Versioning and soft delete are configured, allowing object recovery after deletion", confidence
}

// RetentionPolicyLocked verifies that the retention policy is locked and cannot be modified.
func RetentionPolicyLocked(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Retention == nil {
		return gemara.Unknown, "Retention policy not configured", confidence
	}

	if payload.Retention.RetentionPeriodSeconds <= 0 {
		return gemara.Failed, "Retention policy has no retention period configured", confidence
	}

	if !payload.Retention.IsLocked {
		return gemara.Failed, "Retention policy is not locked. It can be reduced or removed by bucket administrators", confidence
	}

	return gemara.Passed, "Retention policy is locked and cannot be reduced or removed", confidence
}

// --- CN04: Default Retention Policy ---

// DefaultRetentionConfigured verifies that a default retention policy is configured.
func DefaultRetentionConfigured(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Retention == nil {
		return gemara.Failed, "No retention policy is configured. Uploaded objects do not receive automatic retention protection", confidence
	}

	if payload.Retention.RetentionPeriodSeconds <= 0 {
		return gemara.Failed, "Retention policy has no retention period configured", confidence
	}

	return gemara.Passed, "Bucket retention policy is configured, ensuring all uploaded objects receive automatic retention protection", confidence
}

// DeletionPreventedByRetention verifies that objects under active retention cannot be deleted.
func DeletionPreventedByRetention(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Retention == nil {
		return gemara.Failed, "No retention policy is configured, so retention-based deletion prevention is not active", confidence
	}

	if payload.Retention.RetentionPeriodSeconds <= 0 {
		return gemara.Failed, "Retention policy has no retention period configured", confidence
	}

	if !payload.Retention.IsLocked {
		return gemara.NeedsReview, "Retention policy is configured but not locked. Bucket administrators can reduce or remove the retention period. Manual verification required", confidence
	}

	return gemara.Passed, "Locked retention policy prevents deletion of objects under active retention, including by bucket administrators", confidence
}

// --- CN05: Versioning ---

// VersioningEnabled verifies that versioning is enabled on the bucket.
func VersioningEnabled(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Versioning == nil {
		return gemara.Unknown, "Versioning data not available", confidence
	}

	if !payload.Versioning.Enabled {
		return gemara.Failed, "Versioning is not enabled on the bucket", confidence
	}

	return gemara.Passed, "Versioning is enabled on the bucket", confidence
}

// NewVersionOnModification verifies that modifying an object creates a new version.
func NewVersionOnModification(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Versioning == nil {
		return gemara.Unknown, "Versioning data not available", confidence
	}

	if !payload.Versioning.Enabled {
		return gemara.Failed, "Versioning is not enabled, so new versions cannot be created on modification", confidence
	}

	return gemara.NeedsReview, "Versioning is enabled. Manual verification required to confirm that modifying an object creates a new version with a unique generation number", confidence
}

// PreviousVersionsRecoverable verifies that previous versions of objects can be recovered.
func PreviousVersionsRecoverable(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Versioning == nil {
		return gemara.Unknown, "Versioning data not available", confidence
	}

	if !payload.Versioning.Enabled {
		return gemara.Failed, "Versioning is not enabled, so previous versions cannot be recovered", confidence
	}

	return gemara.NeedsReview, "Versioning is enabled. Manual verification required to confirm that previous versions of objects can be recovered after modification", confidence
}

// VersionsRetainedOnDeletion verifies that versions are retained when an object is deleted.
func VersionsRetainedOnDeletion(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Versioning == nil {
		return gemara.Unknown, "Versioning data not available", confidence
	}

	if !payload.Versioning.Enabled {
		return gemara.Failed, "Versioning is not enabled, so versions cannot be retained on deletion", confidence
	}

	return gemara.NeedsReview, "Versioning is enabled. Manual verification required to confirm that noncurrent versions are retained when an object is deleted, allowing recovery", confidence
}

// --- CN06: Access Logging ---

// AccessLoggingConfigured verifies that access logs are stored in a separate data store.
func AccessLoggingConfigured(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Logging == nil || !payload.Logging.Enabled {
		return gemara.Failed, "Bucket logging is not configured. Access logs are not being stored in a separate data store", confidence
	}

	if payload.Logging.LogBucket == "" {
		return gemara.Failed, "Bucket logging is enabled but no target log bucket is specified", confidence
	}

	return gemara.Passed, "Bucket logging is configured, storing access logs in a separate bucket: " + payload.Logging.LogBucket, confidence
}

// LogBucketHighestSensitivity verifies that the log bucket is classified at the highest sensitivity level.
func LogBucketHighestSensitivity(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	payload, message := reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	if payload.Logging == nil || !payload.Logging.Enabled {
		return gemara.NeedsReview, "Bucket logging is not configured. If Cloud Audit Logs are used for logging, manual verification required to confirm logs are classified at the highest sensitivity level", confidence
	}

	if payload.Logging.LogBucketLabels == nil {
		return gemara.NeedsReview, "Log bucket labels not available. Manual verification required to confirm the log bucket is classified at the highest sensitivity level", confidence
	}

	sensitivity, ok := payload.Logging.LogBucketLabels["sensitivity"]
	if !ok {
		return gemara.Failed, "Log bucket does not have a 'sensitivity' label", confidence
	}

	if sensitivity != "high" {
		return gemara.Failed, "Log bucket sensitivity label is '" + sensitivity + "', expected 'high'", confidence
	}

	return gemara.Passed, "Log bucket is labeled with sensitivity=high, classified at the highest sensitivity level", confidence
}

// --- CN07: MFA Delete ---

// MfaDeleteSupported verifies that MFA Delete is available as a configuration option.
func MfaDeleteSupported(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	// GCS does not support MFA Delete as a native feature.
	// Alternative controls include locked retention policies and IAM conditions.
	return gemara.NeedsReview, "GCS does not natively support MFA Delete. Alternative controls include locked retention policies, IAM Conditions, and Organization Policy constraints. Manual verification required to confirm equivalent protections are in place", confidence
}

// MfaDeleteEnforced verifies that MFA Delete is enabled and enforced.
func MfaDeleteEnforced(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	return gemara.NeedsReview, "GCS does not natively support MFA Delete enforcement. Consider using locked retention policies to prevent object deletion, or IAM Conditions requiring specific authentication levels. Manual verification required", confidence
}

// MfaDeletionLogged verifies that deletion attempts are logged with MFA status.
func MfaDeletionLogged(payloadData any) (result gemara.Result, message string, confidence gemara.ConfidenceLevel) {
	_, message = reusable_steps.VerifyPayload(payloadData)
	if message != "" {
		return gemara.Unknown, message, confidence
	}

	// Cloud Audit Logs automatically record all deletion attempts for GCS.
	// However, MFA status is not a native concept in GCS audit logs.
	return gemara.NeedsReview, "Cloud Audit Logs record all GCS deletion attempts including authentication details. However, GCS does not have a native MFA Delete concept, so MFA validation status is not explicitly recorded. Manual verification required", confidence
}
