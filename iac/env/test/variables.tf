variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region for all resources"
  type        = string
  default     = "us-central1"
}

variable "bucket_name" {
  description = "GCS bucket name. If empty, auto-generated."
  type        = string
  default     = ""
}

variable "retention_policy_locked" {
  description = "Whether to lock the retention policy"
  type        = bool
  default     = false
}

variable "retention_period_seconds" {
  description = "Default retention period in seconds"
  type        = number
  default     = 86400
}

variable "soft_delete_retention_seconds" {
  description = "Soft delete retention duration in seconds"
  type        = number
  default     = 604800
}

variable "kms_key_rotation_period" {
  description = "Rotation period for the Cloud KMS key (e.g., '7776000s' = 90 days)"
  type        = string
  default     = "7776000s"
}

variable "log_retention_days" {
  description = "Number of days to retain objects in the log bucket"
  type        = number
  default     = 90
}

variable "manage_audit_config" {
  description = "Whether to manage the project-level audit config for storage.googleapis.com"
  type        = bool
  default     = true
}

variable "force_destroy" {
  description = "Allow Terraform to destroy buckets even if they contain objects"
  type        = bool
  default     = true
}

variable "labels" {
  description = "Labels to apply to all resources"
  type        = map(string)
  default = {
    environment = "test"
    managed_by  = "terraform"
    project     = "pvtr-gcp-cloud-storage"
  }
}
