// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ServiceAccountResource ensures that a specific Service Account exists.
type ServiceAccountResource struct {
	AccountID   string
	DisplayName string
	Description string
}

func (r *ServiceAccountResource) Name() string {
	return fmt.Sprintf("Service Account (%s)", r.AccountID)
}

func (r *ServiceAccountResource) Key() string { return "sa" }

func (r *ServiceAccountResource) desiredState() *adminpb.ServiceAccount {
	return &adminpb.ServiceAccount{
		DisplayName: r.DisplayName,
		Description: r.Description,
	}
}

func (r *ServiceAccountResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	sa, err := client.IAMAdmin.GetServiceAccount(ctx, &adminpb.GetServiceAccountRequest{
		Name: fmt.Sprintf(
			"projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com",
			client.ProjectID,
			r.AccountID,
			client.ProjectID,
		),
	})

	if status.Code(err) == codes.NotFound {
		return "Create Service Account", true, nil
	}
	if err != nil {
		return "", false, err
	}

	desired := r.desiredState()
	// diffMessages compares fields present in desired.
	// Note: We need to use valid protoreflect messages.
	// adminpb.ServiceAccount is a proto message.
	diffs := diffMessages(desired.ProtoReflect(), sa.ProtoReflect(), "")

	if len(diffs) > 0 {
		return fmt.Sprintf("Update: %s", strings.Join(diffs, ", ")), true, nil
	}

	return "", false, nil
}

func (r *ServiceAccountResource) Apply(ctx context.Context, client *GCPClient) error {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", r.AccountID, client.ProjectID)

	// Check existence using Get
	_, err := client.IAMAdmin.GetServiceAccount(ctx, &adminpb.GetServiceAccountRequest{
		Name: fmt.Sprintf("projects/%s/serviceAccounts/%s", client.ProjectID, email),
	})

	if status.Code(err) == codes.NotFound {
		// Create
		desired := r.desiredState()
		_, err := client.IAMAdmin.CreateServiceAccount(ctx, &adminpb.CreateServiceAccountRequest{
			Name:           fmt.Sprintf("projects/%s", client.ProjectID),
			AccountId:      r.AccountID,
			ServiceAccount: desired,
		})
		return err
	}
	if err != nil {
		return err
	}

	// Update check
	_, needsUpdate, err := r.Diff(ctx, client)
	if err != nil {
		return err
	}

	if needsUpdate {
		// Update logic
		// Note: ServiceAccount only has DisplayName updates typically supported easily.
		// However, for strict compliance we error if not implemented fully or try to update.
		// The original code errored on update. We stick to that unless we want to implement UpdateServiceAccount.
		// client.IAMAdmin.UpdateServiceAccount(ctx, &adminpb.UpdateServiceAccountRequest{...})
		return fmt.Errorf("update not implemented for Service Account drift")
	}

	return nil
}
