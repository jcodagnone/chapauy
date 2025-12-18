// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// DeployService deploys the web service to Cloud Run.
func DeployService(ctx context.Context, client *GCPClient, dryRun bool) error {
	// 0. Resolve the latest image digest
	// We do this even in dry-run to verify the image exists and print the digest we would setup.
	log.Println("🔍 Resolving latest image digest...")
	imageRef, err := resolveLatestDigest(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to resolve latest image digest: %w", err)
	}
	log.Printf("   Resolved: %s\n", imageRef)

	if dryRun {
		log.Println("dry-run: Skipping deployment")
		return nil
	}

	log.Println("🚀 Deploying to Cloud Run...")
	parent := DefaultParent
	serviceID := fmt.Sprintf("%s/services/%s", parent, ServiceName)

	// Prepare the service definition
	service := &runpb.Service{
		Name: serviceID,
		Template: &runpb.RevisionTemplate{
			Containers: []*runpb.Container{
				{
					// Use specific digest to force update if changed
					Image: imageRef,
					Ports: []*runpb.ContainerPort{
						{ContainerPort: 3000},
					},
					Resources: &runpb.ResourceRequirements{
						CpuIdle:         true,
						StartupCpuBoost: true,
						Limits: map[string]string{
							"memory": "2Gi",
							"cpu":    "2",
						},
					},
				},
			},
			ExecutionEnvironment: runpb.ExecutionEnvironment_EXECUTION_ENVIRONMENT_GEN2,
			ServiceAccount:       SAName + "@" + ProjectID + ".iam.gserviceaccount.com",
			Scaling: &runpb.RevisionScaling{
				MinInstanceCount: 1,
				MaxInstanceCount: 4,
			},
			MaxInstanceRequestConcurrency: 160,
			Timeout: &durationpb.Duration{
				Seconds: 15,
			},
		},
		Ingress: runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
	}

	// Try to Get first to see if update or create
	_, err = client.RunClient.GetService(ctx, &runpb.GetServiceRequest{Name: serviceID})
	if err == nil {
		// Update
		op, err := client.RunClient.UpdateService(ctx, &runpb.UpdateServiceRequest{
			Service: service,
		})
		if err != nil {
			return fmt.Errorf("failed to update service: %w", err)
		}
		if _, err = op.Wait(ctx); err != nil {
			return fmt.Errorf("failed to wait for update operation: %w", err)
		}
		log.Println("✅ Service updated successfully")
	} else if strings.Contains(err.Error(), "NotFound") {
		// Create
		// For CreateService, the service.Name must be empty. The ID is passed via ServiceId.
		service.Name = ""
		op, err := client.RunClient.CreateService(ctx, &runpb.CreateServiceRequest{
			Parent:    parent,
			Service:   service,
			ServiceId: ServiceName,
		})
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}
		if _, err = op.Wait(ctx); err != nil {
			return fmt.Errorf("failed to wait for create operation: %w", err)
		}
		log.Println("✅ Service created successfully")
	} else {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// 4. Set IAM Policy (Allow Unauthenticated)
	log.Println("🔓 Setting IAM policy to allow unauthenticated access...")
	policy := &iampb.Policy{
		Bindings: []*iampb.Binding{
			{
				Role:    "roles/run.invoker",
				Members: []string{"allUsers"},
			},
		},
	}
	_, err = client.RunClient.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: serviceID,
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("failed to set IAM policy: %w", err)
	}
	log.Println("✅ IAM policy updated (allUsers -> roles/run.invoker)")

	return nil
}

func resolveLatestDigest(ctx context.Context, client *GCPClient) (string, error) {
	// Name format: projects/*/locations/*/repositories/*/packages/*/tags/*
	tagName := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/tags/%s",
		client.ProjectID, client.Region, RepoName, WebDataImageName, "latest")

	tag, err := client.ArtifactRegistry.GetTag(ctx, &artifactregistrypb.GetTagRequest{Name: tagName})
	if err != nil {
		return "", err
	}

	// tag.Version is the full resource name of the version
	// e.g. projects/.../versions/sha256:12345...
	parts := strings.Split(tag.Version, "/")
	versionID := parts[len(parts)-1]

	// construct image ref with digest
	// Images.Registry + "/" + WebDataImageName + "@" + versionID
	return fmt.Sprintf("%s/%s@%s", Images.Registry, WebDataImageName, versionID), nil
}
