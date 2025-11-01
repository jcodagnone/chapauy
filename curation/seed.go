// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SeedData represents the JSON seed file format.
type SeedData struct {
	Version     string      `json:"version"`
	LastUpdated time.Time   `json:"last_updated"`
	Judgments   []*Location `json:"judgments"`
}

// ExportToJSON exports all judgments to a JSON file.
func ExportToJSON(repo LocationRepository, filepath string) error {
	judgments, err := repo.ListJudgments(nil, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("listing judgments: %w", err)
	}

	seed := &SeedData{
		Version:     "1.0",
		LastUpdated: time.Now(),
		Judgments:   judgments,
	}

	data, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	err = os.WriteFile(filepath, data, 0o600)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// ImportFromJSON imports judgments from a JSON file.
func ImportFromJSON(repo LocationRepository, filepath string) (int, error) {
	data, err := os.ReadFile(filepath) // #nosec G304 - filepath is provided by admin
	if err != nil {
		return 0, fmt.Errorf("reading file: %w", err)
	}

	var seed SeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return 0, fmt.Errorf("parsing JSON: %w", err)
	}

	imported := 0

	for _, judgment := range seed.Judgments {
		if err := repo.SaveJudgment(judgment); err != nil {
			return imported, fmt.Errorf("saving judgment for %s: %w", judgment.Location, err)
		}

		imported++
	}

	return imported, nil
}

// SeedIfEmpty seeds the database from a JSON file if no judgments exist.
func SeedIfEmpty(repo LocationRepository, filepath string) (bool, int, error) {
	count, err := repo.CountJudgments()
	if err != nil {
		return false, 0, fmt.Errorf("counting judgments: %w", err)
	}

	if count > 0 {
		return false, count, nil
	}
	// Database is empty, try to seed
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// No seed file exists, that's okay
		return false, 0, nil
	}

	imported, err := ImportFromJSON(repo, filepath)
	if err != nil {
		return false, 0, err
	}

	return true, imported, nil
}
