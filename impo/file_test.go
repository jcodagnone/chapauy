// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

// for testing purposes, if your SearchResultEntry is defined differently,
// you can use a simple definition here:
// type SearchResultEntry struct {
//	   ID string
//	   Data string
// }

// TestFileStore_Upsert covers different branches of FileStore.Upsert.
func TestFileStore_Upsert(t *testing.T) {
	t.Run("Upsert_NewFile", func(t *testing.T) {
		// Create a temporary directory to act as the root for FileStore.
		tmpDir := t.TempDir()
		fs := NewFileStore(tmpDir, &DbReference{ID: 45})

		// Define new entries.
		entries := []SearchResultEntry{
			{Href: "01_2025"}, // only ID is needed for testing branch logic
			{Href: "02_2025"},
		}

		// Call Upsert. Since no file exists yet, it should create one.
		if n, err := fs.Upsert(entries, false); err != nil || n != 2 {
			t.Fatalf("Upsert failed: %d, %v", n, err)
		}
		// Build expected file path.
		data, err := os.ReadFile(fs.dbpath())
		if err != nil {
			t.Fatalf("failed to read notifications file: %v", err)
		}

		// Unmarshal the JSON file into a map.
		var m map[string]SearchResultEntry
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		// Verify that both entries have been inserted.
		if len(m) != 2 {
			t.Errorf("expected 2 entries, got %d", len(m))
		}

		if _, ok := m["01_2025"]; !ok {
			t.Error("01_2025 not found in stored map")
		}

		if _, ok := m["02_2025"]; !ok {
			t.Error("02_2025 not found in stored map")
		}
	})

	t.Run("Upsert_ExistingFile", func(t *testing.T) {
		tmpDir := t.TempDir()

		fs := NewFileStore(tmpDir, &DbReference{ID: 45})
		if err := os.MkdirAll(fs.root, 0o700); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		// Construct initial map with one entry.
		initial := map[string]SearchResultEntry{
			"01_2025": {Href: "01_2025"},
		}

		initialData, err := json.MarshalIndent(initial, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal initial JSON: %v", err)
		}

		if err := os.WriteFile(fs.dbpath(), initialData, 0o600); err != nil {
			t.Fatalf("failed to write initial notifications file: %v", err)
		}

		// Now prepare new entries with one duplicate (01_2025) and one new (02_2025).
		newEntries := []SearchResultEntry{
			{Href: "01_2025"}, // duplicate; should not override
			{Href: "02_2025"},
		}
		if n, err := fs.Upsert(newEntries, false); err != nil || n != 1 {
			t.Fatalf("Upsert failed - expected 1 but got %d, %v", n, err)
		}

		// Verify that the file now contains both 01_2025 and 02_2025.
		data, err := os.ReadFile(fs.dbpath())
		if err != nil {
			t.Fatalf("failed to read notifications file: %v", err)
		}

		var result map[string]SearchResultEntry
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to unmarshal notifications JSON: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result))
		}

		if _, ok := result["01_2025"]; !ok {
			t.Error("01_2025 not found in stored map")
		}

		if _, ok := result["02_2025"]; !ok {
			t.Error("02_2025 not found in stored map")
		}
	})

	t.Run("Upsert_InvalidJSON", func(t *testing.T) {
		tmpDir := t.TempDir()

		fs := NewFileStore(tmpDir, &DbReference{ID: 45})
		if err := os.MkdirAll(fs.root, 0o700); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}

		// Create a notifications file with invalid JSON.
		invalidContent := []byte("invalid json")
		if err := os.WriteFile(fs.dbpath(), invalidContent, 0o600); err != nil {
			t.Fatalf("failed to write invalid notifications file: %v", err)
		}

		// Attempt to upsert new entries; expect an error because JSON is invalid.
		newEntries := []SearchResultEntry{
			{Href: "01_2025"},
		}

		if n, err := fs.Upsert(newEntries, false); err == nil || n != 0 {
			t.Errorf("expected error on invalid JSON content, got nil")
		} else if !errors.Is(err, &json.SyntaxError{}) {
			// If not directly a SyntaxError, you can check the error string.
			if msg := err.Error(); msg == "" || msg == "nil" {
				t.Errorf("expected a descriptive error, got: %v", err)
			}
		}
	})
}
