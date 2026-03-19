data "google_project" "current" {}

locals {
  bucket_name     = var.bucket_name != "" ? var.bucket_name : "pvtr-gcs-${random_string.suffix.result}"
  log_bucket_name = "${local.bucket_name}-logs"
  project_id      = data.google_project.current.project_id
}

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
  numeric = true
}

# -----------------------------------------------------------------------------
# Cloud KMS (for CMEK encryption)
# Maps to: CCC.ObjStor.CN01 - Trusted KMS key enforcement
# -----------------------------------------------------------------------------

resource "google_kms_key_ring" "this" {
  name     = local.bucket_name
  location = var.region
}

resource "google_kms_crypto_key" "this" {
  name            = local.bucket_name
  key_ring        = google_kms_key_ring.this.id
  rotation_period = var.kms_key_rotation_period
  labels          = var.labels

  lifecycle {
    prevent_destroy = false
  }
}

# Grant the GCS service agent permission to use the KMS key
resource "google_kms_crypto_key_iam_member" "gcs_encrypt" {
  crypto_key_id = google_kms_crypto_key.this.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.current.number}@gs-project-accounts.iam.gserviceaccount.com"
}

# -----------------------------------------------------------------------------
# GCS Bucket (main)
# Maps to: CN01 (CMEK), CN02 (uniform access), CN03 (retention/recovery),
#           CN04 (default retention), CN05 (versioning)
# -----------------------------------------------------------------------------

resource "google_storage_bucket" "this" {
  name                        = local.bucket_name
  location                    = var.region
  project                     = local.project_id
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  labels                      = var.labels

  versioning {
    enabled = true
  }

  encryption {
    default_kms_key_name = google_kms_crypto_key.this.id
  }

  retention_policy {
    is_locked        = var.retention_policy_locked
    retention_period = var.retention_period_seconds
  }

  soft_delete_policy {
    retention_duration_in_seconds = var.soft_delete_retention_seconds
  }

  logging {
    log_bucket        = google_storage_bucket.logs.id
    log_object_prefix = "access-logs/"
  }

  depends_on = [google_kms_crypto_key_iam_member.gcs_encrypt]
}

# -----------------------------------------------------------------------------
# Log Bucket (for access logs)
# Maps to: CCC.ObjStor.CN06 - Access logs in separate data store
# -----------------------------------------------------------------------------

resource "google_storage_bucket" "logs" {
  name                        = local.log_bucket_name
  location                    = var.region
  project                     = local.project_id
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  labels = merge(var.labels, {
    sensitivity = "high"
    purpose     = "access-logs"
  })

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = var.log_retention_days
    }
    action {
      type = "Delete"
    }
  }

  lifecycle_rule {
    condition {
      num_newer_versions = 3
    }
    action {
      type = "Delete"
    }
  }

  soft_delete_policy {
    retention_duration_in_seconds = var.soft_delete_retention_seconds
  }
}

# -----------------------------------------------------------------------------
# Cloud Audit Logs for GCS Data Access
# Maps to: CCC.ObjStor.CN06 (access logging), CN07.AR03 (deletion audit)
# -----------------------------------------------------------------------------

resource "google_project_iam_audit_config" "storage" {
  project = local.project_id
  service = "storage.googleapis.com"

  audit_log_config {
    log_type = "ADMIN_READ"
  }

  audit_log_config {
    log_type = "DATA_READ"
  }

  audit_log_config {
    log_type = "DATA_WRITE"
  }
}
