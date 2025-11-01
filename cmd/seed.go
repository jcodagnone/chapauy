// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jcodagnone/chapauy/impo"
	"github.com/spf13/cobra"
)

func newSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seeds the database with data from cmd/testdata/seed.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := os.MkdirAll(impoOptions.DbPath, 0o750); err != nil {
				return fmt.Errorf("creating db directory: %w", err)
			}
			dbpath := filepath.Join(impoOptions.DbPath, "chapauy.duckdb")

			return seedDatabase(dbpath)
		},
	}
}

func init() {
	rootCmd.AddCommand(newSeedCmd())
}

func seedDatabase(dbPath string) error {
	// remove old db if it exists
	_ = os.Remove(dbPath)
	_ = os.Remove(dbPath + ".wal")

	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	repo, err := impo.NewSQLOffenseRepository(db)
	if err != nil {
		return fmt.Errorf("initializing repository: %w", err)
	}

	if err := repo.CreateSchema(); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	jsonFile, err := os.Open("cmd/testdata/seed.json")
	if err != nil {
		return fmt.Errorf("failed to open seed.json: %w", err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var offenses []*impo.TrafficOffense
	if err := json.Unmarshal(byteValue, &offenses); err != nil {
		return fmt.Errorf("failed to unmarshal seed.json: %w", err)
	}

	// Group offenses by doc_source
	offensesBySource := make(map[string][]*impo.TrafficOffense)
	for _, o := range offenses {
		offensesBySource[o.DocSource] = append(offensesBySource[o.DocSource], o)
	}

	for _, group := range offensesBySource {
		if err := repo.SaveTrafficOffenses(group); err != nil {
			return fmt.Errorf("failed to save offenses for %s: %w", group[0].DocSource, err)
		}
	}

	fmt.Println("Database seeded successfully.")

	return nil
}
