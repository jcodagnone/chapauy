// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"dagger/chapauy/infra"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	// Import from .dagger/infra - this works because we'll add a replace directive
	"github.com/spf13/cobra"
)

type logWriter struct {
	writer io.Writer
}

func (w *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Fprintf(w.writer, "%s %s", time.Now().Format("2006-01-02 15:04:05"), string(bytes))
}

func init() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{writer: os.Stderr})
}

func main() {
	var target string
	var credsFile string
	var apply bool

	rootCmd := &cobra.Command{
		Use:   "infra",
		Short: "Manage GCP infrastructure for ChapaUY",
		Long:  `A standalone CLI tool for managing GCP infrastructure.`,
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup GCP infrastructure",
		Long: `This tool performs idempotent reconciliation of GCP resources,
detecting drift and optionally applying changes.

Without --target, it runs in dry-run mode showing detected drift.
With --target, it applies changes to the specified resource.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var jsonCreds string
			if credsFile != "" {
				data, err := os.ReadFile(credsFile)
				if err != nil {
					return err
				}
				jsonCreds = string(data)
			}

			if err := infra.Setup(context.Background(), jsonCreds, target, !apply, infra.DesiredState()); err != nil {
				return err
			}

			return nil
		},
	}

	setupCmd.Flags().StringVar(&target, "target", "", "Target resource to apply (services, registry, sa, iam, devconnect, trigger)")
	setupCmd.Flags().StringVar(&credsFile, "creds", "", "Path to Service Account JSON key file")
	setupCmd.Flags().BoolVar(&apply, "apply", false, "Apply changes to the specified resource")

	mapsCmd := &cobra.Command{
		Use:   "maps",
		Short: "Setup Google Maps Geocoding API and Key",
		Long:  `Enables the Geocoding API and creates an API Key restricted to it.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var jsonCreds string
			if credsFile != "" {
				data, err := os.ReadFile(credsFile)
				if err != nil {
					return err
				}
				jsonCreds = string(data)
			}

			// For maps, we probably always want to apply or at least standard behavior
			// But let's respect the apply flag or maybe default to apply?
			// The user requirement says "setups ... programmatically", implying action.
			// But sticking to the pattern: dry-run by default unless --apply is passed is safer.
			// However, for a specific "setup maps" command, user expectation is action.
			// Let's reuse the same flags.

			if err := infra.Setup(context.Background(), jsonCreds, target, !apply, infra.MapsDesiredState()); err != nil {
				return err
			}
			return nil
		},
	}
	// Reuse flags for maps command
	mapsCmd.Flags().StringVar(&credsFile, "creds", "", "Path to Service Account JSON key file")
	mapsCmd.Flags().BoolVar(&apply, "apply", false, "Apply changes")
	mapsCmd.Flags().StringVar(&target, "target", "", "Target resource to apply")

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the web service to Cloud Run",
		RunE: func(_ *cobra.Command, _ []string) error {
			var jsonCreds string
			if credsFile != "" {
				data, err := os.ReadFile(credsFile)
				if err != nil {
					return err
				}
				jsonCreds = string(data)
			}

			// Deploy command runs directly against GCP APIs
			client, err := infra.NewClient(context.Background(), []byte(jsonCreds), "", infra.ProjectID, infra.Region)
			if err != nil {
				return err
			}
			defer client.Close()

			return infra.DeployService(context.Background(), client, !apply)
		},
	}
	deployCmd.Flags().StringVar(&credsFile, "creds", "", "Path to Service Account JSON key file")
	deployCmd.Flags().BoolVar(&apply, "apply", false, "Actually deploy (default is dry-run)")

	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(mapsCmd)
	rootCmd.AddCommand(deployCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available resources",
		RunE: func(_ *cobra.Command, _ []string) error {
			allResources := append(infra.DesiredState(), infra.MapsDesiredState()...)

			// Calculate max key length for alignment
			maxKeyLen := 3 // Length of "KEY"
			for _, r := range allResources {
				if len(r.Key()) > maxKeyLen {
					maxKeyLen = len(r.Key())
				}
			}

			format := fmt.Sprintf("%%-%ds %%s\n", maxKeyLen)
			fmt.Printf(format, "KEY", "NAME")
			// Print separator
			separator := ""
			for i := 0; i < maxKeyLen; i++ {
				separator += "-"
			}
			fmt.Printf(format, separator, "----")

			// Use a map to avoid duplicates if any
			seen := make(map[string]bool)
			for _, r := range allResources {
				key := r.Key()
				if seen[key] {
					continue
				}
				seen[key] = true
				fmt.Printf(format, key, r.Name())
			}
			return nil
		},
	}
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
