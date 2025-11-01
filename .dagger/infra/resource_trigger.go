// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"

	cloudbuildpb "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type CloudBuildTriggerResource struct {
	TriggerName    string // Trigger ID/Name
	Description    string
	RepoOwner      string
	RepoName       string
	BranchPattern  string
	Filename       string // Path to cloudbuild.yaml
	ServiceAccount string // Email of the service account
	ConnectionID   string // Developer Connect Connection ID (e.g., "github-conn")
	ManualTrigger  bool   // If true, no push trigger is created (manual only)
	Revision       string // Branch or ref for manual trigger (e.g., "refs/heads/master")
}

func (r *CloudBuildTriggerResource) Name() string {
	return "Cloud Build Trigger: " + r.TriggerName
}

func (r *CloudBuildTriggerResource) Key() string {
	return "trigger-" + r.TriggerName
}

func (r *CloudBuildTriggerResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	// For regional triggers, the TriggerId in GetBuildTriggerRequest must be the full resource name
	// or the simple ID if ProjectId is set.
	// However, for V2 regional triggers, we should use the parent.
	// "projects/{project}/locations/{region}/triggers/{trigger_id}"
	triggerResourceName := fmt.Sprintf("%s/triggers/%s", DefaultParent, r.TriggerName)

	existing, err := client.CloudBuild.GetBuildTrigger(ctx, &cloudbuildpb.GetBuildTriggerRequest{
		Name: triggerResourceName,
		// ProjectId and TriggerId are deprecated/legacy for global?
		// V2 prefers "Name".
	})

	if err != nil {
		// Assume not found if error
		// TODO: Check specific error code for NotFound
		return "Trigger not found (will create)", true, nil
	}

	// Compare fields
	diff := ""
	if existing.Description != r.Description {
		diff += fmt.Sprintf("Description: %s -> %s; ", existing.Description, r.Description)
	}
	// Check Trigger Config
	repoResource := fmt.Sprintf("projects/%s/locations/%s/connections/%s/repositories/%s-%s",
		ProjectID, Region, r.ConnectionID, r.RepoOwner, r.RepoName)

	if r.ManualTrigger {
		// Manual Trigger uses GitFileSource (Gen 2)
		if gfs := existing.GetGitFileSource(); gfs != nil {
			if gfs.GetRepository() != repoResource {
				diff += fmt.Sprintf("Repository: %s -> %s; ", gfs.GetRepository(), repoResource)
			}
			if gfs.Path != r.Filename {
				diff += fmt.Sprintf("Filename: %s -> %s; ", gfs.Path, r.Filename)
			}
			if gfs.Revision != r.Revision {
				diff += fmt.Sprintf("Revision: %s -> %s; ", gfs.Revision, r.Revision)
			}
		} else {
			diff += "GitFileSource missing (expected Manual trigger); "
		}
		// Should NOT have RepositoryEventConfig
		if existing.GetRepositoryEventConfig() != nil {
			diff += "RepositoryEventConfig present (expected Manual trigger); "
		}
	} else {
		if existing.GetFilename() != r.Filename {
			diff += fmt.Sprintf("Filename: %s -> %s; ", existing.GetFilename(), r.Filename)
		}
		// Event Trigger uses RepositoryEventConfig
		if rep := existing.GetRepositoryEventConfig(); rep != nil {
			if rep.Repository != repoResource {
				diff += fmt.Sprintf("Repository: %s -> %s; ", rep.Repository, repoResource)
			}
			// Push is nested
			if rep.GetPush().GetBranch() != r.BranchPattern {
				diff += fmt.Sprintf("Branch: %s -> %s; ", rep.GetPush().GetBranch(), r.BranchPattern)
			}
		} else {
			diff += "RepositoryEventConfig missing (might be V1 trigger); "
		}
		// Should NOT have GitFileSource
		if existing.GetGitFileSource() != nil {
			diff += "GitFileSource present (expected Event trigger); "
		}
	}

	// Check Service Account
	expectedSA := fmt.Sprintf("projects/%s/serviceAccounts/%s", ProjectID, r.ServiceAccount)
	if existing.ServiceAccount != expectedSA {
		diff += fmt.Sprintf("ServiceAccount: %s -> %s; ", existing.ServiceAccount, expectedSA)
	}

	// Check SourceToBuild
	if stb := existing.GetSourceToBuild(); stb != nil {
		if stb.GetRef() != r.Revision {
			diff += fmt.Sprintf("SourceToBuild.Ref: %s -> %s; ", stb.GetRef(), r.Revision)
		}
		if stb.GetRepository() != repoResource {
			diff += fmt.Sprintf("SourceToBuild.Repository: %s -> %s; ", stb.GetRepository(), repoResource)
		}
	} else {
		diff += "SourceToBuild missing; "
	}

	if diff != "" {
		return diff, true, nil
	}

	return "", false, nil
}

func (r *CloudBuildTriggerResource) Apply(ctx context.Context, client *GCPClient) error {
	saPath := fmt.Sprintf("projects/%s/serviceAccounts/%s", ProjectID, r.ServiceAccount)
	repoResource := fmt.Sprintf("projects/%s/locations/%s/connections/%s/repositories/%s-%s",
		ProjectID, Region, r.ConnectionID, r.RepoOwner, r.RepoName)

	// Note: For regional triggers, Name should be empty or user-assigned?
	// The resource name is "projects/.../triggers/..."
	// CloudBuild trigger creation allows `resource_name` ???
	// Usually we omit Name in Create and rely on ID.
	// But `r.TriggerName` is our desired ID.
	// V2 CreateBuildTriggerRequest has `trigger_id` field? No. It relies on the ID in the body?
	// Checking V2: CreateBuildTriggerRequest has `Parent` and `Trigger`.
	// The Trigger message has `Name`. If we set `Name`, it might be ignored or used?
	// Actually, `resource_name` is output only mostly?
	// We want to specifying the ID.
	// Looking at `CreateBuildTriggerRequest` logic: it does not have a separate `trigger_id` field in standard protos usually,
	// but might respect `Name` if it's in the format `projects/.../triggers/{id}`.

	trigger := &cloudbuildpb.BuildTrigger{
		Name:           r.TriggerName,
		Description:    r.Description,
		ServiceAccount: saPath,
		SourceToBuild: &cloudbuildpb.GitRepoSource{
			Source: &cloudbuildpb.GitRepoSource_Repository{
				Repository: repoResource,
			},
			Ref: r.Revision,
		},
	}

	if r.ManualTrigger {
		// Manual uses GitFileSource
		trigger.BuildTemplate = &cloudbuildpb.BuildTrigger_GitFileSource{
			GitFileSource: &cloudbuildpb.GitFileSource{
				Path: r.Filename,
				Source: &cloudbuildpb.GitFileSource_Repository{
					Repository: repoResource,
				},
				Revision: r.Revision,
			},
		}
	} else {
		// Event Trigger uses Filename (BuildTemplate) + RepositoryEventConfig
		trigger.BuildTemplate = &cloudbuildpb.BuildTrigger_Filename{
			Filename: r.Filename,
		}
		trigger.RepositoryEventConfig = &cloudbuildpb.RepositoryEventConfig{
			Repository: repoResource,
			Filter: &cloudbuildpb.RepositoryEventConfig_Push{
				Push: &cloudbuildpb.PushFilter{
					GitRef: &cloudbuildpb.PushFilter_Branch{
						Branch: r.BranchPattern,
					},
				},
			},
		}
	}

	triggerResourceName := fmt.Sprintf("%s/triggers/%s", DefaultParent, r.TriggerName)

	// Check existence again to decide Create or Update
	existing, err := client.CloudBuild.GetBuildTrigger(ctx, &cloudbuildpb.GetBuildTriggerRequest{
		Name: triggerResourceName,
	})

	if err == nil {
		// Update
		log.Printf("Updating trigger %s...", r.TriggerName)
		// For Update, we need the full resource name (UUID preferred for ID)
		trigger.Id = existing.Id
		trigger.ResourceName = existing.ResourceName
		trigger.Name = existing.Name

		paths := []string{"description", "service_account"}
		if r.ManualTrigger {
			paths = append(paths, "git_file_source")
		} else {
			paths = append(paths, "filename", "repository_event_config")
		}

		_, err = client.CloudBuild.UpdateBuildTrigger(ctx, &cloudbuildpb.UpdateBuildTriggerRequest{
			ProjectId: ProjectID,
			TriggerId: existing.Id,
			Trigger:   trigger,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: append(paths, "source_to_build"),
			},
		})
	} else {
		// Create
		log.Printf("Creating trigger %s...", r.TriggerName)
		_, err = client.CloudBuild.CreateBuildTrigger(ctx, &cloudbuildpb.CreateBuildTriggerRequest{
			Parent:  DefaultParent, // Regional Parent!
			Trigger: trigger,
		})
	}

	return err
}
