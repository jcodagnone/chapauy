// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"dagger/chapauy/infra"
	"dagger/chapauy/internal/dagger"
	"fmt"
)

// Creates the initial state image from a local directory
func (c *Chapauy) DataBootstrap(
	ctx context.Context,
	// +defaultPath="db"
	stateDir *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		WithWorkdir("/app").
		WithDirectory("db", stateDir)
}

func (c *Chapauy) DataBootstrapAndPublish(
	ctx context.Context,
	// +defaultPath="db"
	stateDir *dagger.Directory,
	token *dagger.Secret,
) error {
	if _, err := publish(ctx, token, c.DataBootstrap(ctx, stateDir), infra.DataImageName); err != nil {
		return fmt.Errorf("failed to publish data: %w", err)
	}
	return nil
}
