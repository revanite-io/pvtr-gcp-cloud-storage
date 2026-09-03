terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 7.25"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 8.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# Required GCP APIs
resource "google_project_service" "storage" {
  service            = "storage.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "kms" {
  service            = "cloudkms.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "resourcemanager" {
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

module "pvtr_gcp_cloud_storage" {
  source = "git::https://github.com/revanite-io/pvtr-terraform.git//modules/pvtr-gcp-cloud-storage?ref=0a981f5a96421f34a9f7d4b9fc9890886ee1f830" # v0.1.3

  region                        = var.region
  bucket_name                   = var.bucket_name
  retention_policy_locked       = var.retention_policy_locked
  retention_period_seconds      = var.retention_period_seconds
  soft_delete_retention_seconds = var.soft_delete_retention_seconds
  kms_key_rotation_period       = var.kms_key_rotation_period
  log_retention_days            = var.log_retention_days
  manage_audit_config           = var.manage_audit_config
  force_destroy                 = var.force_destroy
  labels                        = var.labels

  depends_on = [
    google_project_service.storage,
    google_project_service.kms,
    google_project_service.resourcemanager,
  ]
}

output "bucket_name" {
  description = "Name of the GCS bucket under test"
  value       = module.pvtr_gcp_cloud_storage.bucket_name
}
