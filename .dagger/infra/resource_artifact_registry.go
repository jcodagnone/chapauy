// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArtifactRegistryResource ensures that a Docker repository exists in the specified region.
type ArtifactRegistryResource struct {
	RepoName    string
	Description string
}

func (r *ArtifactRegistryResource) Name() string {
	return fmt.Sprintf("Artifact Registry (%s)", r.RepoName)
}

func (r *ArtifactRegistryResource) Key() string { return "registry" }

func (r *ArtifactRegistryResource) desiredState() *artifactregistrypb.Repository {
	keepCount := int32(3)
	tagState := artifactregistrypb.CleanupPolicyCondition_ANY
	return &artifactregistrypb.Repository{
		Format:      artifactregistrypb.Repository_DOCKER,
		Mode:        artifactregistrypb.Repository_STANDARD_REPOSITORY,
		Description: r.Description,
		FormatConfig: &artifactregistrypb.Repository_DockerConfig{
			DockerConfig: &artifactregistrypb.Repository_DockerRepositoryConfig{
				ImmutableTags: false,
			},
		},
		VulnerabilityScanningConfig: &artifactregistrypb.Repository_VulnerabilityScanningConfig{
			EnablementConfig: artifactregistrypb.Repository_VulnerabilityScanningConfig_DISABLED,
		},
		CleanupPolicies: map[string]*artifactregistrypb.CleanupPolicy{
			"keep-most-recent-3": {
				Action: artifactregistrypb.CleanupPolicy_KEEP,
				ConditionType: &artifactregistrypb.CleanupPolicy_MostRecentVersions{
					MostRecentVersions: &artifactregistrypb.CleanupPolicyMostRecentVersions{
						KeepCount: &keepCount,
					},
				},
			},
			"delete-other-versions": {
				Action: artifactregistrypb.CleanupPolicy_DELETE,
				ConditionType: &artifactregistrypb.CleanupPolicy_Condition{
					Condition: &artifactregistrypb.CleanupPolicyCondition{
						TagState: &tagState,
					},
				},
			},
		},
	}
}

func (r *ArtifactRegistryResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	name := fmt.Sprintf("%s/repositories/%s", DefaultParent, r.RepoName)
	repo, err := client.ArtifactRegistry.GetRepository(ctx, &artifactregistrypb.GetRepositoryRequest{Name: name})
	if status.Code(err) == codes.NotFound {
		return "Create Repository", true, nil
	}
	if err != nil {
		return "", false, err
	}

	desired := r.desiredState()
	diffs := diffMessages(desired.ProtoReflect(), repo.ProtoReflect(), "")

	if len(diffs) > 0 {
		return fmt.Sprintf("Update: %s", strings.Join(diffs, ", ")), true, nil
	}

	return "", false, nil
}

func (r *ArtifactRegistryResource) Apply(ctx context.Context, client *GCPClient) error {
	// check if repository exists
	repoName := fmt.Sprintf("%s/repositories/%s", DefaultParent, r.RepoName)
	_, err := client.ArtifactRegistry.GetRepository(
		ctx,
		&artifactregistrypb.GetRepositoryRequest{
			Name: repoName,
		})

	if status.Code(err) == codes.NotFound {
		desired := r.desiredState()
		_, err := client.ArtifactRegistry.CreateRepository(
			ctx,
			&artifactregistrypb.CreateRepositoryRequest{
				Parent:       DefaultParent,
				RepositoryId: r.RepoName,
				Repository:   desired,
			},
		)
		return err
	}
	if err != nil {
		return err
	}

	// If it exists, we check if it needs update
	_, needsUpdate, err := r.Diff(ctx, client)
	if err != nil {
		return err
	}
	if needsUpdate {
		return fmt.Errorf("update not implemented for Artifact Registry drift")
	}

	return nil
}
