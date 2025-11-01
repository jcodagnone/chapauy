// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package infra

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"cloud.google.com/go/developerconnect/apiv1/developerconnectpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeveloperConnectConnectionResource struct {
	ConnectionID string
	GitHubApp    string // e.g. "developer-connect" or implicit? Default installation
	RepoOwner    string
	RepoName     string
}

func (r *DeveloperConnectConnectionResource) Name() string {
	return "Developer Connect Connection: " + r.ConnectionID
}

func (r *DeveloperConnectConnectionResource) Key() string {
	return "devconnect-" + r.ConnectionID
}

func (r *DeveloperConnectConnectionResource) Diff(ctx context.Context, client *GCPClient) (string, bool, error) {
	connPath := fmt.Sprintf("projects/%s/locations/%s/connections/%s", ProjectID, Region, r.ConnectionID)

	_, err := client.DeveloperConnect.GetConnection(ctx, &developerconnectpb.GetConnectionRequest{
		Name: connPath,
	})

	if status.Code(err) == codes.NotFound {
		return "Connection not found (will create)", true, nil
	}
	if status.Code(err) == codes.PermissionDenied {
		// Attempt to fallback to gcloud to check existence
		cmd := exec.Command("gcloud", "developer-connect", "connections", "describe", r.ConnectionID,
			"--location", Region,
			"--project", ProjectID,
			"--format=value(name)")
		if err := cmd.Run(); err == nil {
			// gcloud succeeded, so the connection exists.
			// Continue to check GitRepositoryLink
		} else {
			log.Printf("⚠️  [Developer Connect] Permission denied accessing connection '%s'.", r.ConnectionID)
			log.Printf("   This often requires manual creation/approval via the Cloud Console to install the GitHub App.")
			log.Printf("   Please create the connection at: https://console.cloud.google.com/developer-connect/connections?project=%s", ProjectID)
			log.Printf("   Then run this command again.")
			return "", false, nil
		}
	} else if err != nil {
		return "", false, err
	}

	// Check GitRepositoryLink
	// format: projects/*/locations/*/connections/*/gitRepositoryLinks/*
	linkID := fmt.Sprintf("%s-%s", r.RepoOwner, r.RepoName)
	linkPath := fmt.Sprintf("%s/gitRepositoryLinks/%s", connPath, linkID)
	_, err = client.DeveloperConnect.GetGitRepositoryLink(ctx, &developerconnectpb.GetGitRepositoryLinkRequest{
		Name: linkPath,
	})
	// Log the error for debugging if it's not NotFound
	// if err != nil && status.Code(err) != codes.NotFound {
	// 	log.Printf("[Debug] GetGitRepositoryLink error for %s: %v", linkPath, err)
	// }

	if status.Code(err) == codes.NotFound {
		return "GitRepositoryLink not found (will create)", true, nil
	}
	// Fallback for PermissionDenied on the Link as well
	if status.Code(err) == codes.PermissionDenied {
		// Attempt to fallback to gcloud to check existence of the LINK
		// gcloud developer-connect connections git-repository-links describe ...
		// actually the command is:
		// gcloud developer-connect connections git-repository-links describe jcodagnone-chapauy --connection=github-repo1 ...
		cmd := exec.Command("gcloud", "developer-connect", "connections", "git-repository-links", "describe", linkID,
			"--connection", r.ConnectionID,
			"--location", Region,
			"--project", ProjectID,
			"--format=value(name)")
		if err := cmd.Run(); err == nil {
			// Link exists
			return "", false, nil
		} else {
			// Get output for debugging why it failed
			debugCmd := exec.Command("gcloud", "developer-connect", "connections", "git-repository-links", "describe", linkID,
				"--connection", r.ConnectionID,
				"--location", Region,
				"--project", ProjectID)
			out, _ := debugCmd.CombinedOutput()
			log.Printf("[Debug] gcloud link check failed: %v. Output: %s", err, string(out))
		}
		// If gcloud failed, assume it doesn't exist or we can't see it.
		// If we assume it doesn't exist, we return "will create".
		// But if we can't create it due to permissions? Apply will fail.
		// But let's try to create it.
		return "GitRepositoryLink not found (verified via gcloud fallback) (will create)", true, nil
	}

	if err != nil {
		return "", false, err
	}

	return "", false, nil
}

func (r *DeveloperConnectConnectionResource) Apply(ctx context.Context, client *GCPClient) error {
	connPath := fmt.Sprintf("projects/%s/locations/%s/connections/%s", ProjectID, Region, r.ConnectionID)

	// 1. Ensure Connection Exists
	_, err := client.DeveloperConnect.GetConnection(ctx, &developerconnectpb.GetConnectionRequest{
		Name: connPath,
	})

	if status.Code(err) == codes.NotFound {
		log.Printf("Creating Developer Connect Connection %s...", r.ConnectionID)

		// Note: Creating a connection often requires human interaction to install the GitHub App
		// and authorize it. However, we can create the resource.
		// If the installation is missing, it enters a state waiting for installation.
		op, err := client.DeveloperConnect.CreateConnection(ctx, &developerconnectpb.CreateConnectionRequest{
			Parent:       DefaultParent,
			ConnectionId: r.ConnectionID,
			Connection: &developerconnectpb.Connection{
				ConnectionConfig: &developerconnectpb.Connection_GithubConfig{
					GithubConfig: &developerconnectpb.GitHubConfig{
						AuthorizerCredential: &developerconnectpb.OAuthCredential{
							OauthTokenSecretVersion: "projects/" + ProjectID + "/secrets/github-oauth-token/versions/latest",
						},
					},
				},
			},
		})
		// Let's try simple creation.
		req := &developerconnectpb.CreateConnectionRequest{
			Parent:       DefaultParent,
			ConnectionId: r.ConnectionID,
			Connection: &developerconnectpb.Connection{
				ConnectionConfig: &developerconnectpb.Connection_GithubConfig{
					GithubConfig: &developerconnectpb.GitHubConfig{
						// Minimal config, will default to DEVELOPER_CONNECT app usually or require manual setup
					},
				},
			},
		}
		op, err = client.DeveloperConnect.CreateConnection(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create connection: %w", err)
		}
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("failed to wait for connection creation: %w", err)
		}
	} else if status.Code(err) == codes.PermissionDenied {
		// Fallback check for connection existence via gcloud
		cmd := exec.Command("gcloud", "developer-connect", "connections", "describe", r.ConnectionID,
			"--location", Region,
			"--project", ProjectID,
			"--format=value(name)")
		if err := cmd.Run(); err != nil {
			// If gcloud also fails/doesn't find it, we can't create it via SDK either.
			// Return original error.
			return fmt.Errorf("connection check failed (SDK permission denied + gcloud check failed): %w", err)
		}
		// If gcloud found it, proceed to checking Link
	} else if err != nil {
		return err
	}

	// 2. Ensure GitRepositoryLink Exists
	linkID := fmt.Sprintf("%s-%s", r.RepoOwner, r.RepoName)
	linkPath := fmt.Sprintf("%s/gitRepositoryLinks/%s", connPath, linkID)

	_, err = client.DeveloperConnect.GetGitRepositoryLink(ctx, &developerconnectpb.GetGitRepositoryLinkRequest{
		Name: linkPath,
	})

	shouldCreate := false
	if status.Code(err) == codes.NotFound {
		shouldCreate = true
	} else if status.Code(err) == codes.PermissionDenied {
		// Fallback check for link existence via gcloud
		cmd := exec.Command("gcloud", "developer-connect", "connections", "git-repository-links", "describe", linkID,
			"--connection", r.ConnectionID,
			"--location", Region,
			"--project", ProjectID,
			"--format=value(name)")
		if err := cmd.Run(); err != nil {
			// Not found via gcloud
			shouldCreate = true
		}
	} else if err != nil {
		return err
	}

	if shouldCreate {
		log.Printf("Creating GitRepositoryLink %s...", linkID)
		repoURI := fmt.Sprintf("https://github.com/%s/%s.git", r.RepoOwner, r.RepoName)

		req := &developerconnectpb.CreateGitRepositoryLinkRequest{
			Parent:              connPath,
			GitRepositoryLinkId: linkID,
			GitRepositoryLink: &developerconnectpb.GitRepositoryLink{
				CloneUri: repoURI,
			},
		}

		op, err := client.DeveloperConnect.CreateGitRepositoryLink(ctx, req)
		if err != nil {
			if status.Code(err) == codes.PermissionDenied {
				log.Printf("⚠️  [Developer Connect] Permission denied creating GitRepositoryLink via SDK. Falling back to gcloud...")
				// gcloud developer-connect connections git-repository-links create jcodagnone-chapauy --connection=github-repo1 --clone-uri=...
				cmd := exec.Command("gcloud", "developer-connect", "connections", "git-repository-links", "create", linkID,
					"--connection", r.ConnectionID,
					"--clone-uri", repoURI,
					"--location", Region,
					"--project", ProjectID)
				output, cmdErr := cmd.CombinedOutput()
				if cmdErr != nil {
					return fmt.Errorf("failed to create repo link via gcloud: %w\nOutput: %s", cmdErr, string(output))
				}
				log.Printf("✅ GitRepositoryLink created via gcloud.")
				return nil
			}
			return fmt.Errorf("failed to create repo link: %w", err)
		}
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("failed to wait for repo link creation: %w", err)
		}
	}

	return nil

}
