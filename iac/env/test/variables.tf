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

variable "log_retention_days" {
  description = "Number of days to retain objects in the log bucket"
  type        = number
  default     = 90
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
