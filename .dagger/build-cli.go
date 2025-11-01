// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

// Builds the backed project
package main

import (
	"context"
	"dagger/chapauy/internal/dagger"
)

const (
	cliUser = "appuser" // we'll create this user in the container
)

// Builds the CLI binary
func (c *Chapauy) BuildCliBase(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=["web", "db" ]
	src *dagger.Directory,
) *dagger.Container {
	//dictates where Go stores its build cacheDir data, which includes compiled
	// packages and other build artifacts.
	const cacheDir = "/home/" + cliUser + "/.cache"
	const goBuild = cacheDir + "/go-build"

	return dag.Container().
		// we use bookworm and not something like alpine because duckdb is
		// very sensitive to musl
		From("golang:1.25.5-bookworm").
		// Create a non-root user 'appuser' to avoid running the build as root,
		// trying to improve security (process will have a different uid in the host)
		WithExec([]string{"useradd", "-m", "-u", "1000", cliUser}).
		WithWorkdir("/src").
		// try to reduce cache invalidations between builds even if dependencies changes
		WithMountedCache(
			"/go/pkg",
			dag.CacheVolume("go-pkg"),
			dagger.ContainerWithMountedCacheOpts{Owner: cliUser},
		).
		WithEnvVariable("GOCACHE", goBuild).
		WithMountedCache(
			cacheDir,
			dag.CacheVolume("go-cache"),
			dagger.ContainerWithMountedCacheOpts{Owner: cliUser},
		).
		// copy go.mod and go.sum first so the cache invalidates only when deps changes
		// we don't want plain source code  change to invalidate the dependencies cache
		WithFile("go.mod", src.File("go.mod")).
		WithFile("go.sum", src.File("go.sum")).
		WithExec([]string{"chown", "-R", cliUser + ":" + cliUser,
			"/src",
			"/home/" + cliUser,
		}).
		WithUser(cliUser).
		WithExec([]string{"go", "mod", "download"}).
		// now that we have dependencies, copy the rest of the source code
		WithUser("root").
		WithDirectory("/src", src.WithoutDirectory("web")).
		WithExec([]string{"chown", "-R", cliUser + ":" + cliUser, "/src"}).
		WithUser(cliUser).
		WithExec([]string{"go", "build", "-o", "build/chapa", "main.go"})
}

// Runs validation on CLI code
func (c *Chapauy) BuildCliValidate(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=["web", "db" ]
	src *dagger.Directory,
) *dagger.Container {
	return c.BuildCliBase(ctx, src).
		// make deps
		WithExec([]string{"go", "install", "-v", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"}).
		WithExec([]string{"go", "install", "-v", "github.com/securego/gosec/v2/cmd/gosec@latest"}).
		WithExec([]string{"go", "install", "-v", "golang.org/x/vuln/cmd/govulncheck@latest"}).
		WithExec([]string{"go", "install", "-v", "honnef.co/go/tools/cmd/staticcheck@latest"}).
		WithExec([]string{"go", "install", "-v", "github.com/google/addlicense@latest"}).
		WithExec([]string{
			"golangci-lint",
			"run",
			"--timeout",
			"5m",
			"./...",
		}).
		WithExec([]string{
			"gosec",
			"-no-fail",
			"-exclude-generated",
			"-exclude-dir", ".dagger",
			"./...",
		}).
		WithExec([]string{"govulncheck", "./..."}).
		WithExec([]string{
			"addlicense",
			"--check",
			"--ignore", "build/**",
			"--ignore", "web/**",
			"--ignore", ".dagger/internal/**",
			"-c", "The ChapaUY Authors",
			"-l", "apache",
			"-s=only",
			".",
		})
}

// Returns a container with the CLI built standalone
func (c *Chapauy) BuildCli(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=["web", "db" ]
	src *dagger.Directory,
) *dagger.Container {
	// Stage 1: Build the binary
	builder := c.BuildCliBase(ctx, src)

	// Stage 2: Create the runtime container
	return dag.Container().
		From("gcr.io/distroless/cc-debian12").
		WithWorkdir("/app").
		WithFile("/app/chapa", builder.File("/src/build/chapa")).
		WithFile("/app/judgments.json", src.File("judgments.json")).
		WithUser(distrolessUser)
}
