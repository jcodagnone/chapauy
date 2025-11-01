// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Updated imports
	apikeys "cloud.google.com/go/apikeys/apiv2"
	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2" // Using V2
	developerconnect "cloud.google.com/go/developerconnect/apiv1"
	admin "cloud.google.com/go/iam/admin/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"

	// The 'run' import was removed as per the user's implied change in the provided snippet.
	scheduler "cloud.google.com/go/scheduler/apiv1"
	su "cloud.google.com/go/serviceusage/apiv1"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

// GCP Configuration
const (
	ProjectID        = "chapauy-20251216"
	Region           = "us-east4" //"southamerica-east1"
	RepoName         = "prod"     // name of the artifact repository
	DataImageName    = "data"     // image name for the "Data Volume Container"
	WebDataImageName = "web-data" // image name for the Web + "Data Volume Container"
	CLIImageName     = "cli"      // name of the CLI service runner
	ServiceName      = "web"      // name of the web service runner
	SAName           = "deploy"   // name of the service account used to run API

	// DefaultParent project/location path for the default region
	DefaultParent = "projects/" + ProjectID + "/locations/" + Region
)

// Images Centralizes image references
var Images = struct {
	RegistryAddr string
	Registry     string
	CLI          string
	Data         string
	Web          string
	WebData      string
}{
	RegistryAddr: fmt.Sprintf("%s-docker.pkg.dev", Region),
	Registry:     fmt.Sprintf("%s-docker.pkg.dev/%s/%s", Region, ProjectID, RepoName),
	CLI:          fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", Region, ProjectID, RepoName, CLIImageName),
	Data:         fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", Region, ProjectID, RepoName, DataImageName),
	Web:          fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", Region, ProjectID, RepoName, ServiceName),
	WebData:      fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", Region, ProjectID, RepoName, WebDataImageName),
}

// Resource represents a single GCP infrastructure component that can
// detect drift (Diff) and reconcile itself (Apply).
type Resource interface {
	Name() string
	Key() string
	Diff(ctx context.Context, client *GCPClient) (string, bool, error)
	Apply(ctx context.Context, client *GCPClient) error
}

func Setup(
	ctx context.Context,
	jsonCreds string,
	target string,
	dryRun bool,
	resources []Resource,
) error {
	client, err := NewClient(ctx, []byte(jsonCreds), "", ProjectID, Region)
	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}
	defer client.Close()

	log.Println("Reconciling...")

	for _, result := range resources {
		// Filter by target if provided
		if target != "" && result.Key() != target && target != "platform" && target != "all" {
			continue
		}
		name := result.Name()
		diff, needed, err := result.Diff(ctx, client)
		if err != nil {
			return fmt.Errorf("failed to check resource %s: %w", name, err)
		}

		if !needed {
			log.Printf("✅ %s\n", name)
			continue
		}

		// If target is NOT set, we are in "Dry Run" / "Scan" mode.
		// We only apply if a specific target is requested.
		// EXCEPTION: "apply-all" convention or similar?
		// The original code said: "If target is NOT set ... We only apply if a specific target is requested."
		// Wait, did I mess up the logic? original code:
		// if target == "" { logs... Drift detected... continue }
		// So by default it's dry-run. User must pass target="all" or specific?
		// The comment said "Draft detected! (Run with --target=...)"
		// If I want to auto-apply on Deploy, I need to pass a target that matches.
		// Let's support target="all" or "platform" to apply everything.

		if dryRun {
			log.Printf("⚠️  %s: Drift detected! (Run with --target=%s --apply to apply)\n   diff: %s\n", name, result.Key(), diff)
			continue
		} else {
			log.Printf("⚙️  %s: Drift detected. Applying changes... (%s)\n", name, diff)
			if err := result.Apply(ctx, client); err != nil {
				return fmt.Errorf("failed to apply resource %s: %w", name, err)
			}
			log.Printf("   %s: Successfully applied.\n", name)
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////////
// Internals

// GCPClient is the concrete implementation using Google Cloud SDKs.
// It exposes the authenticated clients directly for resources to use.
type GCPClient struct {
	ProjectID     string
	ProjectNumber string
	Region        string

	// Clients (Protobuf-based)
	ServiceUsageClient *su.Client
	ArtifactRegistry   *artifactregistry.Client
	IAMAdmin           *admin.IamClient
	ResourceManager    *resourcemanager.ProjectsClient
	CloudBuild         *cloudbuild.Client
	RunClient          *run.ServicesClient
	DeveloperConnect   *developerconnect.Client
	Scheduler          *scheduler.CloudSchedulerClient
	APIKeys            *apikeys.Client
}

// NewClient creates a new authenticated GCP client.
func NewClient(ctx context.Context, jsonCreds []byte, token string, projectID, region string) (*GCPClient, error) {
	var opts []option.ClientOption
	if len(jsonCreds) > 0 {
		opts = append(opts, option.WithCredentialsJSON(jsonCreds))
	} else if token != "" {
		opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
	}
	// If jsonCreds and token are empty, we rely on Application Default Credentials (ADC)
	// We explicitly set the quota project and scopes to avoid ambiguity, especially for newer APIs
	// like Developer Connect which might be sensitive to this.
	if projectID != "" {
		opts = append(opts, option.WithQuotaProject(projectID))
	}
	opts = append(opts, option.WithScopes("https://www.googleapis.com/auth/cloud-platform"))

	// Service Usage
	suClient, err := su.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Usage client: %w", err)
	}

	// Bootstrap: Ensure Service Usage and Resource Manager are enabled.
	// We need Resource Manager to get the Project Number, but we can't get it if the API is disabled.
	log.Printf("Bootstrapping: Ensuring critical APIs are enabled for %s...", projectID)
	bootstrapReq := &serviceusagepb.BatchEnableServicesRequest{
		Parent: "projects/" + projectID,
		ServiceIds: []string{
			"serviceusage.googleapis.com",
			"cloudresourcemanager.googleapis.com",
		},
	}
	op, err := suClient.BatchEnableServices(ctx, bootstrapReq)
	if err != nil {
		// If we don't have permission to enable services, we shouldn't fail the whole client creation.
		// The services might already be enabled.
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.PermissionDenied {
			log.Printf("⚠️  Warning: Permission denied to enable services (%v). Assuming services are already enabled and continuing.", err)
		} else {
			return nil, fmt.Errorf("failed to bootstrap enable services: %w", err)
		}
	} else {
		if _, err = op.Wait(ctx); err != nil {
			return nil, fmt.Errorf("failed to wait for bootstrap enable: %w", err)
		}
	}

	// Artifact Registry
	ar, err := artifactregistry.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Artifact Registry client: %w", err)
	}

	// IAM
	iamAdmin, err := admin.NewIamClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM client: %w", err)
	}

	// Resource Manager (Project IAM Policy)
	rmClient, err := resourcemanager.NewProjectsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Resource Manager client: %w", err)
	}

	// Cloud Build
	cbClient, err := cloudbuild.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Build client: %w", err)
	}

	// Cloud Run
	runClient, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %w", err)
	}

	// Developer Connect
	devConnect, err := developerconnect.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Developer Connect client: %w", err)
	}

	// Cloud Scheduler
	schedClient, err := scheduler.NewCloudSchedulerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Scheduler client: %w", err)
	}

	// API Keys
	apiKeysClient, err := apikeys.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create API Keys client: %w", err)
	}

	// Fetch Project Number
	p, err := rmClient.GetProject(ctx, &resourcemanagerpb.GetProjectRequest{
		Name: "projects/" + projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	// Project Name in format "projects/12345"
	// Actually proto has Name like "projects/123". But let's check field.
	// The resource name of the project is "projects/{project_id}".
	// But we want the number. "projects/12345".
	// The returned Project message has a `Name` field which IS "projects/{number}".
	// And `ProjectId` field which is the string ID.
	// Let's parse the number from Name.
	projectNumber := p.Name // "projects/12345"
	if len(projectNumber) > 9 && projectNumber[:9] == "projects/" {
		projectNumber = projectNumber[9:]
	}

	return &GCPClient{
		ProjectID:          projectID,
		ProjectNumber:      projectNumber,
		Region:             region,
		ServiceUsageClient: suClient,
		ArtifactRegistry:   ar,
		IAMAdmin:           iamAdmin,
		ResourceManager:    rmClient,
		CloudBuild:         cbClient,
		RunClient:          runClient,
		DeveloperConnect:   devConnect,
		Scheduler:          schedClient,
		APIKeys:            apiKeysClient,
	}, nil
}

func (c *GCPClient) Close() error {
	// Close all clients that need closing
	if err := c.ArtifactRegistry.Close(); err != nil {
		return err
	}
	if err := c.ServiceUsageClient.Close(); err != nil {
		return err
	}
	if err := c.IAMAdmin.Close(); err != nil {
		return err
	}
	if err := c.ResourceManager.Close(); err != nil {
		return err
	}
	if err := c.CloudBuild.Close(); err != nil {
		return err
	}
	if err := c.RunClient.Close(); err != nil {
		return err
	}
	if err := c.Scheduler.Close(); err != nil {
		return err
	}
	if err := c.APIKeys.Close(); err != nil {
		return err
	}
	return nil
}
