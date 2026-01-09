// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

// DesiredState returns the list of all resources that should exist in the GCP project.
// This serves as the "Inventory" or "Infrastructure as Code" definition.
func DesiredState() []Resource {
	return []Resource{
		// ---------------------------------------------------------------------
		// Platform Services
		// ---------------------------------------------------------------------
		// Services like Cloud Run, Artifact Registry, and IAM must be explicitly enabled
		// before we can create resources or deploy applications.
		&ServiceUsageResource{
			Services: []string{
				"cloudbuild.googleapis.com",           // Cloud Build for CI/CD - builds docker image
				"artifactregistry.googleapis.com",     // Artifact Registry for image storage
				"run.googleapis.com",                  // Cloud Run for hosting the frontend
				"iam.googleapis.com",                  // IAM for service account management
				"cloudresourcemanager.googleapis.com", // Resource Manager for project management
				"developerconnect.googleapis.com",     // Developer Connect for repo connections
				"cloudscheduler.googleapis.com",       // Cloud Scheduler for scheduled daily data build
			},
			DisabledServices: []string{
				"analyticshub.googleapis.com",
				"bigquery.googleapis.com",
				"bigqueryconnection.googleapis.com",
				"bigquerydatapolicy.googleapis.com",
				"bigquerydatatransfer.googleapis.com",
				"bigquerymigration.googleapis.com",
				"bigqueryreservation.googleapis.com",
				"bigquerystorage.googleapis.com",
				"dataform.googleapis.com",
				"dataplex.googleapis.com",
				"datastore.googleapis.com",
				"sql-component.googleapis.com",
			},
		},

		// ---------------------------------------------------------------------
		// Artifacts
		// ---------------------------------------------------------------------
		// Cloud Run requires container images to be stored in a registry to pull and deploy them.
		&ArtifactRegistryResource{
			RepoName:    RepoName,
			Description: "Docker repository for Chapauy",
		},

		// ---------------------------------------------------------------------
		// Identity & Access Management
		// ---------------------------------------------------------------------
		// We use a dedicated Service Account for deployments (identity isolation) rather than
		// using personal credentials or the default Compute Engine service account.
		&ServiceAccountResource{
			AccountID:   SAName,
			DisplayName: "Deploy",
			Description: "Used to deploy and manage artifacts",
		},
		// The Service Account needs specific permissions (e.g., Cloud Run Admin) to perform
		// deployment tasks. We grant these via IAM bindings at the project level.
		&IAMPolicyResource{
			SAName: SAName,
			ProjectRoles: []string{
				"roles/run.admin",                         // Cloud Run admin
				"roles/artifactregistry.admin",            // Artifact Registry admin
				"roles/iam.serviceAccountUser",            // Service Account User
				"roles/developerconnect.admin",            // Developer Connect admin
				"roles/storage.admin",                     // Storage admin for builder
				"roles/logging.logWriter",                 // Logging admin for builder
				"roles/serviceusage.serviceUsageAdmin",    // Required to enable services
				"roles/browser",                           // Required to get project number (classic role widely supported)
				"roles/serviceusage.serviceUsageConsumer", // Required for quota project usage (deploy task)
				"roles/cloudbuild.builds.editor",          // Required to trigger builds
			},
		},

		// ---------------------------------------------------------------------
		// Developer Connect Service Agent
		// ---------------------------------------------------------------------
		// The Google-managed service account for Developer Connect needs Secret Manager Admin
		// to store the OAuth token securely.
		&IAMPolicyResource{
			ServiceAgentType: "developer-connect",
			ProjectRoles: []string{
				"roles/secretmanager.admin",
			},
		},
		// The Cloud Build Service Agent needs proper permissions to access Developer Connect connections
		&IAMPolicyResource{
			ServiceAgentType: "cloud-build",
			ProjectRoles: []string{
				"roles/developerconnect.user",  // Replaces non-existent connectionAccessor
				"roles/developerconnect.admin", // For good measure, though accessor should suffice
				"roles/secretmanager.admin",    // Often needed for accessing secrets linked to connections
			},
		},
		// Legacy Cloud Build Service Account (often used for builds/triggers unless configured otherwise)
		&IAMPolicyResource{
			ServiceAgentType: "cloud-build-legacy",
			ProjectRoles: []string{
				"roles/developerconnect.user",
				"roles/developerconnect.admin",
				"roles/artifactregistry.admin",
			},
		},
		// The Cloud Scheduler Service Agent needs to be able to act as the Service Account
		// specified in the job (deploy@...) to create OIDC tokens.
		&IAMPolicyResource{
			ServiceAgentType: "cloud-scheduler",
			ProjectRoles: []string{
				"roles/iam.serviceAccountTokenCreator",
				"roles/iam.serviceAccountUser",
			},
		},

		// ---------------------------------------------------------------------
		// Source Control Connections (Developer Connect)
		// ---------------------------------------------------------------------
		&DeveloperConnectConnectionResource{
			ConnectionID: "github-repo1",
			RepoOwner:    "jcodagnone",
			RepoName:     "chapauy",
		},

		// ---------------------------------------------------------------------
		// CI/CD Triggers
		// ---------------------------------------------------------------------
		// Trigger: Main CI/CD
		// Runs on every push to master to deploy the application.
		&CloudBuildTriggerResource{
			TriggerName:    "build-master",
			Description:    "Build images when push to master",
			ConnectionID:   "github-repo1",
			RepoOwner:      "jcodagnone",
			RepoName:       "chapauy",
			BranchPattern:  "^master$",
			Filename:       "cloudbuild.yaml", // Main build file
			ServiceAccount: SAName + "@" + ProjectID + ".iam.gserviceaccount.com",
		},
		// Trigger: Daily Data Refresh
		// Runs every night to refresh the data using the custom images.
		// Note: We use a non-matching branch pattern to prevent this from triggering on push.
		// It is intended to be triggered only by Cloud Scheduler.
		&CloudBuildTriggerResource{
			TriggerName:    "daily-data-refresh",
			Description:    "Daily Data Refresh (Data Only)",
			ConnectionID:   "github-repo1",
			RepoOwner:      "jcodagnone",
			RepoName:       "chapauy",
			ManualTrigger:  true,
			Revision:       "refs/heads/master",
			Filename:       "cloudbuild-daily.yaml",
			ServiceAccount: SAName + "@" + ProjectID + ".iam.gserviceaccount.com",
		},
		// Trigger: Deploy Web+Data
		// Combines the latest Web image with the latest Data image and deploys to Cloud Run.
		&CloudBuildTriggerResource{
			TriggerName:    "deploy-web",
			Description:    "Deploy Web+Data",
			ConnectionID:   "github-repo1",
			RepoOwner:      "jcodagnone",
			RepoName:       "chapauy",
			ManualTrigger:  true,
			Revision:       "refs/heads/master",
			Filename:       "cloudbuild-deploy.yaml",
			ServiceAccount: SAName + "@" + ProjectID + ".iam.gserviceaccount.com",
		},
		// ---------------------------------------------------------------------
		// Scheduled Jobs
		// ---------------------------------------------------------------------
		&CloudSchedulerResource{
			JobName:        "daily-data-refresh-job",
			Description:    "Triggers the daily data refresh build",
			Schedule:       "0 7 * * 1-5", // 7 AM UYT daily weekdays
			TimeZone:       "America/Montevideo",
			TargetTrigger:  "daily-data-refresh", // Must match TriggerName above
			ServiceAccount: SAName + "@" + ProjectID + ".iam.gserviceaccount.com",
		},
	}
}

// MapsDesiredState returns resources needed for Google Maps Geocoding.
func MapsDesiredState() []Resource {
	return []Resource{
		&ServiceUsageResource{
			Services: []string{
				"geocoding-backend.googleapis.com", // For server-side geocoding
				"apikeys.googleapis.com",           // To create API keys
			},
		},
		&MapsResource{
			DisplayName: "ChapaUY Geocoding Key",
			Description: "Key for server-side geocoding",
			Services: []string{
				"geocoding-backend.googleapis.com",
			},
		},
	}
}
