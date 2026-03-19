output "bucket_name" {
  description = "Name of the main GCS bucket"
  value       = google_storage_bucket.this.name
}

output "bucket_url" {
  description = "URL of the main GCS bucket"
  value       = google_storage_bucket.this.url
}

output "log_bucket_name" {
  description = "Name of the access log bucket"
  value       = google_storage_bucket.logs.name
}

output "kms_key_id" {
  description = "ID of the Cloud KMS key used for encryption"
  value       = google_kms_crypto_key.this.id
}

output "kms_key_ring_id" {
  description = "ID of the Cloud KMS key ring"
  value       = google_kms_key_ring.this.id
}

output "project_id" {
  description = "GCP project ID"
  value       = local.project_id
}
