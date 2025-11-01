// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/scheduler/apiv1/schedulerpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type CloudSchedulerResource struct {
	JobName        string // Name of the job
	Description    string
	Schedule       string // Cron expression e.g. "0 3 * * *"
	TimeZone       string // "America/Montevideo"
	TargetTrigger  string // Name of the trigger to run
	ServiceAccount string // Service Account to use
}

func (r *CloudSchedulerResource) Name() string {
	return "Cloud Scheduler Job: " + r.JobName
}

func (r *CloudSchedulerResource) Key() string {
	return "scheduler-" + r.JobName
}

func (r *CloudSchedulerResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	jobName := fmt.Sprintf("%s/jobs/%s", DefaultParent, r.JobName)

	existing, err := client.Scheduler.GetJob(ctx, &schedulerpb.GetJobRequest{
		Name: jobName,
	})

	if err != nil {
		// Assume not found
		return "Job not found (will create)", true, nil
	}

	diff := ""
	if existing.Schedule != r.Schedule {
		diff += fmt.Sprintf("Schedule: %s -> %s; ", existing.Schedule, r.Schedule)
	}
	if existing.TimeZone != r.TimeZone {
		diff += fmt.Sprintf("TimeZone: %s -> %s; ", existing.TimeZone, r.TimeZone)
	}

	// Target check (HTTP Target)
	// We expect an HTTP target pointing to Cloud Build API
	httpTarget := existing.GetHttpTarget()
	if httpTarget == nil {
		diff += "Target: Not HTTP; "
	} else {
		// Expected URL: https://cloudbuild.googleapis.com/v1/projects/{project}/locations/{region}/triggers/{trigger}:run
		expectedURI := fmt.Sprintf("https://cloudbuild.googleapis.com/v1/projects/%s/locations/%s/triggers/%s:run", ProjectID, Region, r.TargetTrigger)
		if httpTarget.Uri != expectedURI {
			diff += fmt.Sprintf("URI: %s -> %s; ", httpTarget.Uri, expectedURI)
		}

		expectedSA := r.ServiceAccount
		oauthToken := httpTarget.GetOauthToken()
		if oauthToken == nil || oauthToken.ServiceAccountEmail != expectedSA {
			currentEmail := ""
			if oauthToken != nil {
				currentEmail = oauthToken.ServiceAccountEmail
			}
			diff += fmt.Sprintf("SA: %s -> %s; ", currentEmail, expectedSA)
		}

		if oauthToken != nil && oauthToken.Scope != "https://www.googleapis.com/auth/cloud-platform" {
			diff += fmt.Sprintf("Scope: %s -> cloud-platform; ", oauthToken.Scope)
		}

		// Body check
		expectedBody := []byte(`{}`)
		if !bytes.Equal(httpTarget.Body, expectedBody) {
			diff += fmt.Sprintf("Body length: %d -> %d; ", len(httpTarget.Body), len(expectedBody))
		}

		// Headers check
		if httpTarget.Headers["Content-Type"] != "application/json" {
			diff += fmt.Sprintf("Header Content-Type: %s -> application/json; ", httpTarget.Headers["Content-Type"])
		}
	}

	if diff != "" {
		return diff, true, nil
	}

	return "", false, nil
}

func (r *CloudSchedulerResource) Apply(ctx context.Context, client *GCPClient) error {
	jobName := fmt.Sprintf("%s/jobs/%s", DefaultParent, r.JobName)

	// Construct Target URI
	// POST https://cloudbuild.googleapis.com/v1/projects/{project}/locations/{region}/triggers/{trigger}:run
	uri := fmt.Sprintf("https://cloudbuild.googleapis.com/v1/projects/%s/locations/%s/triggers/%s:run", ProjectID, Region, r.TargetTrigger)

	// Body: {}
	// For regional triggers, we rely on the trigger's own SourceToBuild configuration.
	// Providing "branchName" often fails with "Unknown name" in the regional RunTrigger API.
	body := []byte(`{}`)

	job := &schedulerpb.Job{
		Name:        jobName,
		Description: r.Description,
		Schedule:    r.Schedule,
		TimeZone:    r.TimeZone,
		Target: &schedulerpb.Job_HttpTarget{
			HttpTarget: &schedulerpb.HttpTarget{
				Uri:        uri,
				HttpMethod: schedulerpb.HttpMethod_POST,
				Body:       body,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				AuthorizationHeader: &schedulerpb.HttpTarget_OauthToken{
					OauthToken: &schedulerpb.OAuthToken{
						ServiceAccountEmail: r.ServiceAccount,
						Scope:               "https://www.googleapis.com/auth/cloud-platform",
					},
				},
			},
		},
	}

	_, err := client.Scheduler.GetJob(ctx, &schedulerpb.GetJobRequest{Name: jobName})
	if err == nil {
		// Update
		log.Printf("Updating Scheduler Job %s...", r.JobName)
		_, err = client.Scheduler.UpdateJob(ctx, &schedulerpb.UpdateJobRequest{
			Job: job,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"description", "schedule", "time_zone", "http_target"},
			},
		})
	} else {
		// Create
		log.Printf("Creating Scheduler Job %s...", r.JobName)
		_, err = client.Scheduler.CreateJob(ctx, &schedulerpb.CreateJobRequest{
			Parent: DefaultParent,
			Job:    job,
		})
	}

	return err
}
