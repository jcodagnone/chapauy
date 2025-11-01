// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"dagger/chapauy/infra"
	"dagger/chapauy/internal/dagger"
	"fmt"
	"log"
)

type Chapauy struct{}

// Configures the necessary GCP resources idempotently. Detects drifts.
func (c *Chapauy) InfraSetup(
	ctx context.Context,
	// Service Account Key JSON (Must have Project Owner or Editor role to creates resources initially)
	// gcloud auth application-default login
	// dagger call infra-setup --creds=file:$HOME/.config/gcloud/application_default_credentials.json
	// +optional
	creds *dagger.Secret,
	// Optional target resource to apply (e.g. services, registry, sa, iam)
	// +optional
	target string,
) (string, error) {
	// 1. Authenticate
	var jsonCreds string
	if creds != nil {
		var err error
		jsonCreds, err = creds.Plaintext(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get credentials: %w", err)
		}
	}

	// If target is empty, we default to dry-run (scan).
	// If target is set, we apply changes to that target (or all if platform/all).
	dryRun := (target == "")
	err := infra.Setup(ctx, jsonCreds, target, dryRun, infra.DesiredState())
	if err != nil {
		return "", err
	}
	return "Infrastructure setup completed successfully", nil
}

// Builds and publishes all base containers
func (c *Chapauy) BuildAndPublish(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=["db" ]
	src *dagger.Directory,
	token *dagger.Secret,
	// +optional
	gitSha string,
) error {
	cli := c.BuildCli(ctx, src.
		WithoutDirectory("web").
		WithoutDirectory("db"),
	)
	web := c.BuildFrontend(ctx, src.
		Directory("web").
		WithoutDirectory("node_modules").
		WithoutDirectory("chapauy.duckdb").
		WithoutDirectory(".next"),
		gitSha,
	)

	accessToken, err := extractToken(ctx, token)
	if err != nil {
		return err
	}
	token = dag.SetSecret("gcp-token", accessToken)

	if _, err = publish(ctx, token, cli, infra.CLIImageName); err != nil {
		return fmt.Errorf("failed to publish cli: %w", err)
	}

	if _, err = publish(ctx, token, web, infra.ServiceName); err != nil {
		return fmt.Errorf("failed to publish web: %w", err)
	}

	return nil
}

// Performs the daily synchronization of data and redeploy of the web service.
func (c *Chapauy) DataRefresh(
	ctx context.Context,
	// Access Token (optional, used for registry operations)
	// +optional
	token *dagger.Secret,
	// Dry run mode (builds but does not publish)
	// +optional
	dryRun bool,
) error {
	log.Printf("Starting Data Update...\n CLI: %s\n Data: %s\n Web: %s\n", infra.Images.CLI, infra.Images.Data, infra.Images.Web)

	accessToken, err := extractToken(ctx, token)
	if err != nil {
		return err
	}

	tokenSecret := dag.SetSecret("gcp-token", accessToken)

	// We pull the data image to get the current DB state
	dataCtr := dag.Container().
		WithRegistryAuth(infra.Images.RegistryAddr, "oauth2accesstoken", tokenSecret).
		From(infra.Images.Data)

	// We use the CLI image to run the update
	// Note: CLI runs as user 1000 (appuser) or 65532 (distroless) usually.
	// We run as root to ensure we can write to the mounted volume and avoid permission issues.
	// We expect the entrypoint to be compatible or we override it.
	// The binary is at /app/chapa.
	cliCtr := dag.Container().
		WithRegistryAuth(infra.Images.RegistryAddr, "oauth2accesstoken", tokenSecret).
		From(infra.Images.CLI).
		WithUser("root").
		WithDirectory("/app/db", dataCtr.Directory("/app/db")).
		WithExec([]string{"/app/chapa", "impo", "update"})

	// Force execution to verify the update command runs successfully
	if _, err := cliCtr.Sync(ctx); err != nil {
		return fmt.Errorf("failed to execute update command: %w", err)
	}

	// 4. Capture Updated Data
	updatedDb := cliCtr.Directory("/app/db")

	// 5. Publish Updated Data Image
	// Reconstruct the data image structure (Filesystem + DB)
	// DataBootstrap created it as: WithWorkdir("/app").WithDirectory("db", stateDir)
	newDataCtr := dag.Container().
		WithWorkdir("/app").
		WithDirectory("db", updatedDb)

	if dryRun {
		log.Printf("dry-run: Skipping publish for %s", newDataCtr)
	} else {
		if _, err := publish(ctx, tokenSecret, newDataCtr, infra.DataImageName); err != nil {
			return fmt.Errorf("failed to publish updated data: %w", err)
		}
		log.Println("✅ Published updated data image")
	}

	return nil
}

// Builds the Web+Data image by injecting the latest data into the web image
func (c *Chapauy) BuildWebData(
	ctx context.Context,
	// Access Token (optional, used for registry operations)
	// +optional
	token *dagger.Secret,
) error {
	accessToken, err := extractToken(ctx, token)
	if err != nil {
		return err
	}
	tokenSecret := dag.SetSecret("gcp-token", accessToken)

	// 1. Get latest Data image
	dataCtr := dag.Container().
		WithRegistryAuth(infra.Images.RegistryAddr, "oauth2accesstoken", tokenSecret).
		From(infra.Images.Data)

	// 2. Get latest Web image
	webCtr := dag.Container().
		WithRegistryAuth(infra.Images.RegistryAddr, "oauth2accesstoken", tokenSecret).
		From(infra.Images.Web)

	// 3. Inject Data into Web
	// We assume the data is at /app/db in the data image
	// And needs to be at /app/chapauy.duckdb in the web image
	// Note: DataRefresh logic used:
	// updatedDb := cliCtr.Directory("/app/db")
	// webCtr.WithFile("/app/chapauy.duckdb", updatedDb.File("chapauy.duckdb"))

	dbFile := dataCtr.Directory("/app/db").File("chapauy.duckdb")

	webDataCtr := webCtr.
		WithUser("root"). // Switch to root to write file
		WithFile("/app/chapauy.duckdb", dbFile).
		WithUser(distrolessUser) // Switch back to nonroot for runtime

	if _, err := publish(ctx, tokenSecret, webDataCtr, infra.WebDataImageName); err != nil {
		return fmt.Errorf("failed to publish updated web-data image: %w", err)
	}
	log.Println("✅ Published updated web-data image")

	return nil
}

// Deploy triggers a deployment of the latest web service image to Cloud Run.
func (c *Chapauy) Deploy(
	ctx context.Context,
	// Service Account Key JSON (optional, falls back to ADC)
	// +optional
	creds *dagger.Secret,
	// Access Token (optional, alternative to creds/ADC)
	// +optional
	token *dagger.Secret,
	// Dry run mode
	// +optional
	dryRun bool,
) error {
	// 1. Resolve Credentials
	var jsonCreds []byte
	if creds != nil {
		jsonCredsStr, err := creds.Plaintext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get creds plaintext: %w", err)
		}
		jsonCreds = []byte(jsonCredsStr)
	}

	var tokenStr string
	if token != nil {
		var err error
		tokenStr, err = token.Plaintext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get token plaintext: %w", err)
		}
	}

	if dryRun {
		log.Println("dry-run: Skipping deployment")
		return nil
	}

	// 3. Deploy
	infraClient, err := infra.NewClient(ctx, jsonCreds, tokenStr, infra.ProjectID, infra.Region)
	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}
	defer infraClient.Close()

	if err := infra.DeployService(ctx, infraClient, dryRun); err != nil {
		return fmt.Errorf("failed to deploy service: %w", err)
	}

	return nil
}
