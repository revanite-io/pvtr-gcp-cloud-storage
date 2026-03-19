variable "region" {
  description = "GCP region for all resources"
  type        = string
  default     = "us-central1"
}

variable "bucket_name" {
  description = "GCS bucket name. If empty, auto-generated."
  type        = string
  default     = ""

  validation {
    condition     = var.bucket_name == "" || can(regex("^[a-z0-9][a-z0-9._-]{1,61}[a-z0-9]$", var.bucket_name))
    error_message = "Bucket name must be 3-63 characters, lowercase alphanumeric, hyphens, underscores, or periods."
  }
}

variable "retention_policy_locked" {
  description = "Whether to lock the retention policy. Once locked, it cannot be reduced or removed."
  type        = bool
  default     = false
}

variable "retention_period_seconds" {
  description = "Default retention period in seconds (e.g., 86400 = 1 day)"
  type        = number
  default     = 86400
}

variable "soft_delete_retention_seconds" {
  description = "Soft delete retention duration in seconds (default: 7 days = 604800)"
  type        = number
  default     = 604800
}

variable "kms_key_rotation_period" {
  description = "Rotation period for the Cloud KMS key (e.g., '7776000s' = 90 days)"
  type        = string
  default     = "7776000s"
}

variable "log_retention_days" {
  description = "Number of days to retain objects in the access log bucket"
  type        = number
  default     = 90
}

variable "labels" {
  description = "Labels to apply to all resources"
  type        = map(string)
  default     = {}
}
