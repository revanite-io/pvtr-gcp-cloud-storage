# Privateer Plugin for GCP Cloud Storage

> [!WARNING]
> This service is in early access. Functionality and availability may change without notice.

A [Privateer](https://github.com/privateerproj/privateer-sdk) plugin that evaluates Google Cloud Storage (GCS) buckets against the [CCC Object Storage](https://commoncloudcontrols.com/catalogs/storage/object/controls) catalog controls.

## Overview

This plugin connects to a GCS bucket and evaluates its configuration against the Common Cloud Controls (CCC) Object Storage catalog. It checks encryption, access control, versioning, retention, immutability, logging, and policy compliance — producing a structured report of passed, failed, and review-needed controls.

## Prerequisites

- Go 1.26.2 or later
- A GCP project with a Cloud Storage bucket to evaluate
- GCP credentials (one of the following):
  - An active `gcloud auth application-default login` session
  - A service account key with the required roles
  - Workload Identity or `GOOGLE_APPLICATION_CREDENTIALS` environment variable

### Required GCP Permissions

The plugin performs read-only operations. Minimum required roles:

| Role | Scope | Covers |
|------|-------|--------|
| `roles/storage.objectViewer` | Bucket or project | Read bucket data, object metadata, and object data |
| `roles/storage.legacyBucketReader` | Bucket or project | Read bucket metadata including IAM policy and logging config |

## Installation

### From source

```bash
git clone https://github.com/revanite-io/pvtr-gcp-cloud-storage.git
cd pvtr-gcp-cloud-storage
make build
```

### From releases

Download the latest binary from the [releases page](https://github.com/revanite-io/pvtr-gcp-cloud-storage/releases).

## Configuration

Copy `example-config.yml` and customize it:

```bash
cp example-config.yml config.yml
```

At minimum, set the `bucketname` to your target GCS bucket:

```yaml
services:
  myService1:
    plugin: pvtr-gcp-cloud-storage
    policy:
      catalogs: ["CCC.ObjStor"]
    vars:
      bucketname: my-gcs-bucket-name
```

### Authentication

The plugin uses [Application Default Credentials (ADC)](https://cloud.google.com/docs/authentication/application-default-credentials), tried in this order:

1. **`GOOGLE_APPLICATION_CREDENTIALS`** environment variable pointing to a service account key file:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
   ```

2. **`gcloud` CLI credentials** (for local development):
   ```bash
   gcloud auth application-default login
   ```

3. **Workload Identity / attached service account** (for GCE, GKE, Cloud Run, etc.) — no configuration needed; credentials are retrieved automatically from the metadata server.

## Usage

This plugin is designed to run via Privateer. See the [Privateer documentation](https://github.com/privateerproj/privateer-sdk) for details on running plugins.

For local development and debugging:

```bash
make build
```

## Controls Evaluated

The plugin evaluates controls from the CCC.ObjStor catalog:

### CCC.ObjStor (Object Storage)

| Control | Description |
|---------|-------------|
| CN01 | Prevent requests with untrusted KMS keys |
| CN02 | Uniform bucket-level access enforcement |
| CN03 | Bucket deletion recovery and retention policy immutability |
| CN04 | Default retention policies |
| CN05 | Object versioning |
| CN06 | Access logging |
| CN07 | MFA deletion protection |

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-cov

# Build
make build

# Tidy dependencies
make tidy
```

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to get started.

Looking for something to work on? Check the [contribute page](https://github.com/revanite-io/pvtr-gcp-cloud-storage/contribute) for good first issues.

## Related Projects

- [privateer-sdk](https://github.com/privateerproj/privateer-sdk) - The SDK this plugin is built on
- [go-gemara](https://github.com/gemaraproj/go-gemara) - Assessment framework types
- [Common Cloud Controls](https://commoncloudcontrols.com) - The control catalogs evaluated by this plugin
