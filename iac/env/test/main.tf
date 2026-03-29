terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

module "pvtr_gcp_cloud_storage" {
  source = "git::https://github.com/revanite-io/pvtr-terraform.git//modules/pvtr-gcp-cloud-storage?ref=9f8ca38296b4ad4b264b997ff6427285ca7aafdb" # v0.1.0

  region                        = var.region
  bucket_name                   = var.bucket_name
  retention_policy_locked       = var.retention_policy_locked
  retention_period_seconds      = var.retention_period_seconds
  soft_delete_retention_seconds = var.soft_delete_retention_seconds
  log_retention_days            = var.log_retention_days
  labels                        = var.labels
}
