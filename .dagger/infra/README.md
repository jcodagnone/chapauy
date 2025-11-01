# GCP Infrastructure Setup

This directory contains the Infrastructure as Code (IaC) definitions for the ChapaUY project on Google Cloud Platform. It uses Dagger and the Google Cloud Go SDKs to manage resources programmatically.

## Architecture

The infrastructure setup is defined in `resources.go` and includes:

- **Artifact Registry**: Docker repositories.
- **Service Accounts**: Specifically `deploy` for CI/CD.
- **IAM Policies**: Bindings for the service accounts.
- **Service Usage**: Enabling required APIs (Run, Scheduler, Cloud Build, Developer Connect).
- **Source Control**: **Developer Connect (2nd Gen)** for GitHub integration (`github-conn`).
- **Cloud Build Triggers**:
    - `build-master`: Deploys on push to master.
    - `daily-data-refresh`: Scheduled daily build.
- **Cloud Scheduler**:
    - `daily-data-refresh-job`: Triggers the daily build at 3 AM.

## Usage

The setup logic is encapsulated in `cmd/main.go` and can be run via the `dagger` CLI or directly with Go.

### Prerequisites

1.  **GCP Credentials**:
    You must have valid credential with permissions to create resources (e.g., Owner or Editor + Security Admin).
    ```bash
    gcloud auth application-default login
    ```

2.  **Developer Connect Permissions**:
    Ensure your user has `roles/developerconnect.admin` permissions.

### Running Setup

To apply the infrastructure configuration (idempotent):

```bash
cd .dagger
go run gcp/cmd/main.go setup --apply
```

### Dry Run

To see what changes would be made without applying them:

```bash
cd .dagger
go run gcp/cmd/main.go setup
```

### Targeting Specific Resources

You can limit the operation to a specific resource type:

```bash
# Only setup triggers
go run gcp/cmd/main.go setup --target=trigger --apply

# Only setup Developer Connect connection
go run gcp/cmd/main.go setup --target=devconnect --apply
```

## Troubleshooting

### `PermissionDenied` for Developer Connect
If you see `read access to project 'chapauy' was denied` for Developer Connect:
1.  Verify the API is enabled: `gcloud services enable developerconnect.googleapis.com`
2.  Verify your permissions.
3.  Wait a few minutes for IAM propagation.

### "Global Region" Error
Cloud Build Triggers associated with Developer Connect *must* be regional.
We use `southamerica-east1`. Ensure you are not trying to create them in `global`.
