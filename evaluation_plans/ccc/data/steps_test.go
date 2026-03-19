package data

import (
	"testing"

	"github.com/gemaraproj/go-gemara"

	d "github.com/revanite-io/pvtr-gcp-cloud-storage/data"
)

// --- PreventUntrustedKmsKeysForBucketRead ---

func TestPreventUntrustedKmsKeysForBucketRead(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "CMEK configured returns NeedsReview",
			payload: d.Payload{
				Encryption: &d.EncryptionData{
					DefaultKMSKeyName: "projects/p/locations/l/keyRings/kr/cryptoKeys/k",
				},
			},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "no CMEK returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PreventUntrustedKmsKeysForBucketRead(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- PreventUntrustedKmsKeysForBucketWrite ---

func TestPreventUntrustedKmsKeysForBucketWrite(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "CMEK configured returns NeedsReview",
			payload: d.Payload{
				Encryption: &d.EncryptionData{
					DefaultKMSKeyName: "projects/p/locations/l/keyRings/kr/cryptoKeys/k",
				},
			},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "no CMEK returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := PreventUntrustedKmsKeysForBucketWrite(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- UniformBucketLevelAccessEnabled ---

func TestUniformBucketLevelAccessEnabled(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "enabled returns Passed",
			payload: d.Payload{
				UniformAccess: &d.UniformAccessData{Enabled: true},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "disabled returns Failed",
			payload: d.Payload{
				UniformAccess: &d.UniformAccessData{Enabled: false},
			},
			wantResult: gemara.Failed,
		},
		{
			name:       "nil UniformAccess returns Unknown",
			payload:    d.Payload{},
			wantResult: gemara.Unknown,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := UniformBucketLevelAccessEnabled(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- BucketDeletionRecoverable ---

func TestBucketDeletionRecoverable(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "versioning + retention returns Passed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
				Retention:  &d.RetentionData{RetentionPeriodSeconds: 86400, IsLocked: true},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "versioning + soft delete returns Passed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
				SoftDelete: &d.SoftDeleteData{RetentionDurationSeconds: 604800},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "versioning + both returns Passed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
				Retention:  &d.RetentionData{RetentionPeriodSeconds: 86400},
				SoftDelete: &d.SoftDeleteData{RetentionDurationSeconds: 604800},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "versioning only returns Failed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
			},
			wantResult: gemara.Failed,
		},
		{
			name: "no versioning returns Failed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400},
			},
			wantResult: gemara.Failed,
		},
		{
			name:       "empty payload returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := BucketDeletionRecoverable(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- RetentionPolicyLocked ---

func TestRetentionPolicyLocked(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "locked returns Passed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400, IsLocked: true},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "not locked returns Failed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400, IsLocked: false},
			},
			wantResult: gemara.Failed,
		},
		{
			name: "no period returns Failed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 0},
			},
			wantResult: gemara.Failed,
		},
		{
			name:       "nil Retention returns Unknown",
			payload:    d.Payload{},
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := RetentionPolicyLocked(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- DefaultRetentionConfigured ---

func TestDefaultRetentionConfigured(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "configured with period returns Passed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "no period returns Failed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 0},
			},
			wantResult: gemara.Failed,
		},
		{
			name:       "nil Retention returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := DefaultRetentionConfigured(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- DeletionPreventedByRetention ---

func TestDeletionPreventedByRetention(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "locked retention returns Passed",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400, IsLocked: true},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "unlocked retention returns NeedsReview",
			payload: d.Payload{
				Retention: &d.RetentionData{RetentionPeriodSeconds: 86400, IsLocked: false},
			},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "no retention returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := DeletionPreventedByRetention(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- VersioningEnabled ---

func TestVersioningEnabled(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "enabled returns Passed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "disabled returns Failed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: false},
			},
			wantResult: gemara.Failed,
		},
		{
			name:       "nil Versioning returns Unknown",
			payload:    d.Payload{},
			wantResult: gemara.Unknown,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := VersioningEnabled(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- NewVersionOnModification ---

func TestNewVersionOnModification(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "versioning enabled returns NeedsReview",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: true},
			},
			wantResult: gemara.NeedsReview,
		},
		{
			name: "versioning disabled returns Failed",
			payload: d.Payload{
				Versioning: &d.VersioningData{Enabled: false},
			},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := NewVersionOnModification(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- AccessLoggingConfigured ---

func TestAccessLoggingConfigured(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "logging enabled returns Passed",
			payload: d.Payload{
				Logging: &d.LoggingData{
					Enabled:   true,
					LogBucket: "my-log-bucket",
				},
			},
			wantResult: gemara.Passed,
		},
		{
			name:       "no logging returns Failed",
			payload:    d.Payload{},
			wantResult: gemara.Failed,
		},
		{
			name: "logging enabled no bucket returns Failed",
			payload: d.Payload{
				Logging: &d.LoggingData{Enabled: true},
			},
			wantResult: gemara.Failed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := AccessLoggingConfigured(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- LogBucketHighestSensitivity ---

func TestLogBucketHighestSensitivity(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name: "log bucket labeled high returns Passed",
			payload: d.Payload{
				Logging: &d.LoggingData{
					Enabled:         true,
					LogBucketLabels: map[string]string{"sensitivity": "high"},
				},
			},
			wantResult: gemara.Passed,
		},
		{
			name: "log bucket labeled low returns Failed",
			payload: d.Payload{
				Logging: &d.LoggingData{
					Enabled:         true,
					LogBucketLabels: map[string]string{"sensitivity": "low"},
				},
			},
			wantResult: gemara.Failed,
		},
		{
			name: "log bucket no sensitivity label returns Failed",
			payload: d.Payload{
				Logging: &d.LoggingData{
					Enabled:         true,
					LogBucketLabels: map[string]string{"other": "label"},
				},
			},
			wantResult: gemara.Failed,
		},
		{
			name: "no labels available returns NeedsReview",
			payload: d.Payload{
				Logging: &d.LoggingData{Enabled: true},
			},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "no logging returns NeedsReview",
			payload:    d.Payload{},
			wantResult: gemara.NeedsReview,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := LogBucketHighestSensitivity(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- MfaDeleteSupported ---

func TestMfaDeleteSupported(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name:       "valid payload returns NeedsReview",
			payload:    d.Payload{},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := MfaDeleteSupported(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- MfaDeleteEnforced ---

func TestMfaDeleteEnforced(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name:       "valid payload returns NeedsReview",
			payload:    d.Payload{},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := MfaDeleteEnforced(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// --- MfaDeletionLogged ---

func TestMfaDeletionLogged(t *testing.T) {
	tests := []struct {
		name       string
		payload    any
		wantResult gemara.Result
	}{
		{
			name:       "valid payload returns NeedsReview",
			payload:    d.Payload{},
			wantResult: gemara.NeedsReview,
		},
		{
			name:       "wrong type returns Unknown",
			payload:    "not a payload",
			wantResult: gemara.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := MfaDeletionLogged(tt.payload)
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}
