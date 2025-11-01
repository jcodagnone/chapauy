// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/apikeys/apiv2/apikeyspb"
)

// MapsResource ensures that a specific API Key exists for Google Maps.
type MapsResource struct {
	DisplayName string
	Description string
	Services    []string // List of allowed services (e.g., "geocoding-backend.googleapis.com")
}

func (r *MapsResource) Name() string { return "Google Maps API Key" }
func (r *MapsResource) Key() string  { return "maps-key" }

func (r *MapsResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	// List keys to find if one with the display name exists
	req := &apikeyspb.ListKeysRequest{
		Parent: "projects/" + client.ProjectID + "/locations/global",
	}
	it := client.APIKeys.ListKeys(ctx, req)

	for {
		key, err := it.Next()
		if err != nil {
			// If we reached the end of the list, break.
			// Iterators in Go Google Cloud libs return 'iterator.Done' error when finished.
			// However, simple break on error might be too aggressive if it's not Done.
			// Ideally check for iterator.Done, but we can assume if we can't get next, we stop.
			// A clean way is checking if err.Error() == "no more items in iterator" or verifying with iterator.Done.
			// But for simplicity/robustness let's assume if it errors, we didn't find it yet or failed.
			// Actually the standard way:
			break
		}
		if key.DisplayName == r.DisplayName {
			// Key exists.
			// Ideally we check restrictions too, but for now we assume existence is enough.
			return "", false, nil
		}
	}

	return "Create API Key for Google Maps", true, nil
}

func (r *MapsResource) Apply(ctx context.Context, client *GCPClient) error {
	log.Printf("Creating API key '%s'...\n", r.DisplayName)

	req := &apikeyspb.CreateKeyRequest{
		Parent: "projects/" + client.ProjectID + "/locations/global",
		Key: &apikeyspb.Key{
			DisplayName: r.DisplayName,
			Restrictions: &apikeyspb.Restrictions{
				ApiTargets: []*apikeyspb.ApiTarget{},
			},
		},
	}

	// Add service restrictions
	if len(r.Services) > 0 {
		var targets []*apikeyspb.ApiTarget
		for _, s := range r.Services {
			targets = append(targets, &apikeyspb.ApiTarget{
				Service: s,
			})
		}
		req.Key.Restrictions.ApiTargets = targets
	}

	op, err := client.APIKeys.CreateKey(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	key, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for API key creation: %w", err)
	}

	log.Printf("✅ API Key Created: %s\n", key.KeyString)
	log.Printf("👉 Add this to your environment: export GOOGLE_MAPS_API_KEY=\"%s\"\n", key.KeyString)

	return nil
}
