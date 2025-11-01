// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

	"cloud.google.com/go/iam/apiv1/iampb"
	"google.golang.org/protobuf/proto"
)

// IAMPolicyResource ensures that a specific member (Service Account, User, etc.)
// has the required roles on the Project.
type IAMPolicyResource struct {
	// Member identifies the principal.
	// If empty, it defaults to constructing the project Service Account email:
	// "serviceAccount:{SAName}@{ProjectID}.iam.gserviceaccount.com"
	// Examples:
	// - "serviceAccount:my-sa@project.iam.gserviceaccount.com"
	// - "user:jane@example.com"
	Member string
	SAName string // Legacy/Convenience field for project service accounts
	// ServiceAgentType allows specifying a Google-managed service agent abstractly.
	// Supported values: "developer-connect"
	ServiceAgentType string
	ProjectRoles     []string
}

func (r *IAMPolicyResource) Name() string {
	if r.ServiceAgentType != "" {
		return "IAM Policy Binding: Service Agent " + r.ServiceAgentType
	}
	if r.Member != "" {
		return "IAM Policy Binding: " + r.Member
	}
	return "IAM Policy Binding: " + r.SAName
}
func (r *IAMPolicyResource) Key() string {
	if r.ServiceAgentType != "" {
		return "iam-sa-" + r.ServiceAgentType
	}
	if r.Member != "" {
		// sanitizing for key
		return "iam-" + strings.ReplaceAll(r.Member, ":", "-")
	}
	return "iam-" + r.SAName
}

func (r *IAMPolicyResource) desiredState(client *GCPClient, current *iampb.Policy) *iampb.Policy {
	desired := proto.Clone(current).(*iampb.Policy)

	member := r.Member
	if r.ServiceAgentType == "developer-connect" {
		// Construct the service agent email dynamically using the project number
		// format: service-{projectNumber}@gcp-sa-devconnect.iam.gserviceaccount.com
		if client.ProjectNumber != "" {
			member = fmt.Sprintf("serviceAccount:service-%s@gcp-sa-devconnect.iam.gserviceaccount.com", client.ProjectNumber)
		} else {
			log.Printf("⚠️ Warning: ProjectNumber not available for Developer Connect Service Agent IAM binding.")
		}
	} else if r.ServiceAgentType == "cloud-build" {
		// format: service-{projectNumber}@gcp-sa-cloudbuild.iam.gserviceaccount.com
		if client.ProjectNumber != "" {
			member = fmt.Sprintf("serviceAccount:service-%s@gcp-sa-cloudbuild.iam.gserviceaccount.com", client.ProjectNumber)
		} else {
			log.Printf("⚠️ Warning: ProjectNumber not available for Cloud Build Service Agent IAM binding.")
		}
	} else if r.ServiceAgentType == "cloud-scheduler" {
		// format: service-{projectNumber}@gcp-sa-cloudscheduler.iam.gserviceaccount.com
		if client.ProjectNumber != "" {
			member = fmt.Sprintf("serviceAccount:service-%s@gcp-sa-cloudscheduler.iam.gserviceaccount.com", client.ProjectNumber)
		} else {
			log.Printf("⚠️ Warning: ProjectNumber not available for Cloud Scheduler Service Agent IAM binding.")
		}
	} else if r.ServiceAgentType == "cloud-build-legacy" {
		// format: {projectNumber}@cloudbuild.gserviceaccount.com
		if client.ProjectNumber != "" {
			member = fmt.Sprintf("serviceAccount:%s@cloudbuild.gserviceaccount.com", client.ProjectNumber)
		} else {
			log.Printf("⚠️ Warning: ProjectNumber not available for Legacy Cloud Build SA IAM binding.")
		}
	} else if member == "" {
		member = fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", r.SAName, client.ProjectID)
	}

	for _, role := range r.ProjectRoles {
		bindingFound := false
		for _, b := range desired.Bindings {
			if b.Role == role {
				if !slices.Contains(b.Members, member) {
					b.Members = append(b.Members, member)
				}
				bindingFound = true
				break
			}
		}
		if !bindingFound {
			desired.Bindings = append(desired.Bindings, &iampb.Binding{
				Role:    role,
				Members: []string{member},
			})
		}
	}
	return desired
}

func (r *IAMPolicyResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	req := &iampb.GetIamPolicyRequest{
		Resource: fmt.Sprintf("projects/%s", client.ProjectID),
	}
	policy, err := client.ResourceManager.GetIamPolicy(ctx, req)
	if err != nil {
		return "", false, err
	}

	desired := r.desiredState(client, policy)

	// We compare desired vs actual.
	// Since desired is derived from actual + additions, any difference implies we need to apply.
	diffs := diffMessages(desired.ProtoReflect(), policy.ProtoReflect(), "")

	if len(diffs) > 0 {
		return fmt.Sprintf("Update: %s", strings.Join(diffs, ", ")), true, nil
	}

	return "", false, nil
}

func (r *IAMPolicyResource) Apply(ctx context.Context, client *GCPClient) error {
	req := &iampb.GetIamPolicyRequest{
		Resource: fmt.Sprintf("projects/%s", client.ProjectID),
	}
	policy, err := client.ResourceManager.GetIamPolicy(ctx, req)
	if err != nil {
		return err
	}

	// If this is a Service Agent policy, ensure the Service Identity exists.
	if r.ServiceAgentType != "" {
		var serviceName string
		switch r.ServiceAgentType {
		case "developer-connect":
			serviceName = "developerconnect.googleapis.com"
		case "cloud-build", "cloud-build-legacy":
			serviceName = "cloudbuild.googleapis.com"
		case "cloud-scheduler":
			serviceName = "cloudscheduler.googleapis.com"
		default:
			// Fallback or error? For now, log warning and skip generation attempts.
			log.Printf("⚠️ Unknown ServiceAgentType '%s', skipping GenerateServiceIdentity", r.ServiceAgentType)
		}

		if serviceName != "" {
			log.Printf("Ensuring Service Identity exists for %s (using gcloud)...", serviceName)
			// Fallback to gcloud because the Go SDK method GenerateServiceIdentity is missing in the current version.
			cmd := exec.Command("gcloud", "beta", "services", "identity", "create",
				"--service="+serviceName,
				"--project="+client.ProjectID,
				"--quiet",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				// Verify if it failed because it already exists?
				// gcloud usually succeeds if it exists, simply returning the email.
				// But if it fails, we should log output.
				return fmt.Errorf("failed to generate service identity for %s: %w\nOutput: %s", serviceName, err, string(output))
			}
			log.Printf("Service Identity Check/Creation Output: %s", strings.TrimSpace(string(output)))
		}
	}

	desired := r.desiredState(client, policy)

	// Check if update is needed using Diff logic (or strict comparison here)
	if proto.Equal(desired, policy) {
		return nil
	}

	setReq := &iampb.SetIamPolicyRequest{
		Resource: fmt.Sprintf("projects/%s", client.ProjectID),
		Policy:   desired,
	}
	_, err = client.ResourceManager.SetIamPolicy(ctx, setReq)
	return err
}
