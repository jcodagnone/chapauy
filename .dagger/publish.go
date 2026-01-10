// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"dagger/chapauy/infra"
	"dagger/chapauy/internal/dagger"
	"fmt"

	"golang.org/x/oauth2/google"
)

func extractToken(ctx context.Context, token *dagger.Secret) (string, error) {
	var accessToken string

	if token != nil {
		var err error
		accessToken, err = token.Plaintext(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get token plaintext: %w", err)
		}
	} else {
		// Use Application Default Credentials (ADC)
		// This works if running in Cloud Build or if local environment has ADC set up
		credsObj, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return "", fmt.Errorf("failed to find default credentials: %w", err)
		}
		t, err := credsObj.TokenSource.Token()
		if err != nil {
			return "", fmt.Errorf("failed to get token from ADC: %w", err)
		}
		accessToken = t.AccessToken
	}

	if len(accessToken) > 10 {
		fmt.Printf("DEBUG: Extracted token (len=%d): %s...%s\n", len(accessToken), accessToken[:5], accessToken[len(accessToken)-5:])
	} else {
		fmt.Printf("DEBUG: Extracted token is too short or empty: %s\n", accessToken)
	}

	return accessToken, nil
}

// publishes a container to the private registry
func publish(
	ctx context.Context,
	token *dagger.Secret,
	container *dagger.Container,
	name string,
) (string, error) {

	// 4. Publish
	return container.
		WithRegistryAuth(infra.Images.RegistryAddr, "oauth2accesstoken", token).
		// Format: region-docker.pkg.dev/project/repo/image:latest
		Publish(ctx, fmt.Sprintf("%s/%s:latest", infra.Images.Registry, name))
}
