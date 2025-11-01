// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

// Builds the frontend project
package main

import (
	"context"
	"dagger/chapauy/internal/dagger"
)

const (
	distrolessUser = "65532" // nonroot user in distroless images

)

// Returns a container with the frontend built
func (c *Chapauy) BuildFrontend(
	ctx context.Context,
	// +defaultPath="/web"
	// +ignore=["**/.next/**", "**/node_modules/**", "**/chapauy.duckdb"]
	src *dagger.Directory,
	// +optional
	gitSha string,
) *dagger.Container {
	// Stage 1: Builder
	builder := dag.Container().
		From("node:24-bookworm-slim").
		WithWorkdir("/src").
		// we copy only package manager file to try to get better cache invalidations
		WithFile("package.json", src.File("package.json")).
		WithFile("pnpm-lock.yaml", src.File("pnpm-lock.yaml")).
		WithNewFile("GIT_SHA", gitSha).
		WithEnvVariable("GIT_COMMIT_SHA", gitSha).
		WithFile("pnpm-workspace.yaml", src.File("pnpm-workspace.yaml")).
		WithDirectory("patches", src.Directory("patches")).
		WithMountedCache("/root/.local/share/pnpm", dag.CacheVolume("pnpm-data")).
		WithExec([]string{"npm", "install", "-g", "pnpm@latest"}).
		WithExec([]string{
			"pnpm", "install",
			"--frozen-lockfile", // fail on unsynced lockfile
			"--ignore-scripts",  // there are many ways to spread the new worms
			"--config.ignore-scripts=true",
		}).
		// we copy the rest
		WithDirectory("/src", src, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{
				"chapauy.duckdb",
				"node_modules",
				".next",
			},
		}).
		WithExec([]string{"pnpm", "rebuild", "duckdb"}).
		WithExec([]string{"pnpm", "run", "build"})

	// Stage 2: Usage of an intermediate container to set permissions
	// Distroless images don't have a shell, so we can't run chown/mkdir inside them.
	// We use a standard debian image to prepare the filesystem.
	prepper := dag.Container().
		From("node:24-bookworm-slim").
		WithWorkdir("/app").
		// Copy the built app
		WithDirectory("/app", builder.Directory("/src/.next/standalone")).
		WithFile("/app/GIT_SHA", builder.File("/src/GIT_SHA")).
		WithDirectory("/app/.next/static", builder.Directory("/src/.next/static")).
		WithDirectory("/app/public", builder.Directory("/src/public")).
		// Create cache directory (Next.js needs this for ISR/Data Cache)
		WithExec([]string{"mkdir", "-p", "/app/.next/cache"}).
		// Set ownership to nonroot user (uid 65532 is standard for distroless)
		WithExec([]string{"chown", "-R", distrolessUser + ":" + distrolessUser, "/app/.next/cache"})

	// Stage 3: Runner
	// We use a distroless image for maximum security (no shell, no package manager)
	return dag.Container().
		From("gcr.io/distroless/nodejs24-debian12").
		WithWorkdir("/app").
		WithEnvVariable("NODE_ENV", "production").
		WithEnvVariable("PORT", "3000").
		WithEnvVariable("NEXT_TELEMETRY_DISABLED", "1").
		// Copy the prepared filesystem
		WithDirectory("/app", prepper.Directory("/app")).
		// Mount /tmp as a volume (needed by some node libs)
		WithMountedTemp("/tmp").
		// Run as nonroot user
		WithUser(distrolessUser).
		WithExposedPort(3000).
		// Use the absolute path to node in distroless
		WithEntrypoint([]string{
			"/nodejs/bin/node",
			// Enable the permission model. The idea is to restrict what node can do
			// the main idea was to prevent outgoing connections, but that isn't available
			// in node 24 -- https://nodejs.org/api/permissions.html
			"--permission",
			"--allow-fs-read=/app",
			"--allow-fs-read=/tmp",
			"--allow-fs-write=/tmp",
			"--allow-fs-write=/app/.next/cache",
			"--allow-addons", // Needed for DuckDB
			"server.js",
		})
}
