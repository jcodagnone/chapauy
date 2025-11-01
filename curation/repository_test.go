// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/jcodagnone/chapauy/spatial"
)

func setupTestDB(t *testing.T) (*sql.DB, LocationRepository) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	repo := NewLocationRepository(db, map[int]string{})
	if err := repo.CreateSchema(); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db, repo
}

func TestCreateSchema(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	// Verify table exists
	var tableName string

	err := db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'locations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table not created: %v", err)
	}

	if tableName != "locations" {
		t.Errorf("Expected table 'locations', got '%s'", tableName)
	}
}

func TestSaveAndGetJudgment(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	lat := -34.8822366
	lon := -56.1529602

	judgment := &Location{
		DbID:     6,
		Location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
		Point: &spatial.Point{
			Lat: lat,
			Lng: lon,
		},
		IsElectronic:    false,
		GeocodingMethod: "manual",
		Confidence:      "high",
		Notes:           "Verified intersection in Montevideo",
	}

	// Save
	err := repo.SaveJudgment(judgment)
	if err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	dbID := 6
	location := "AV 8 DE OCTUBRE Y AV CENTENARIO"

	judgments, err := repo.ListJudgments(&dbID, &location, 1, 0)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(judgments) == 0 {
		t.Fatalf("ListJudgments() returned no judgment for dbID %d, location %s", dbID, location)
	}

	retrieved := judgments[0]

	if retrieved.DbID != 6 {
		t.Errorf("DbID = %d, want 6", retrieved.DbID)
	}

	if retrieved.Location != judgment.Location {
		t.Errorf("Location = %s, want %s", retrieved.Location, judgment.Location)
	}

	if retrieved.Point.Lat != lat {
		t.Errorf("Latitude = %f, want %f", retrieved.Point.Lat, lat)
	}

	if retrieved.Point.Lng != lon {
		t.Errorf("Longitude = %f, want %f", retrieved.Point.Lng, lon)
	}

	if retrieved.IsElectronic != false {
		t.Errorf("IsElectronic = %v, want false", retrieved.IsElectronic)
	}

	if retrieved.GeocodingMethod != "manual" {
		t.Errorf("GeocodingMethod = %s, want manual", retrieved.GeocodingMethod)
	}

	if retrieved.Confidence != "high" {
		t.Errorf("Confidence = %s, want high", retrieved.Confidence)
	}
}

func TestUpdateJudgment(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	lat1 := -34.8822366
	lon1 := -56.1529602

	judgment := &Location{
		DbID:     6,
		Location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
		Point: &spatial.Point{
			Lat: lat1,
			Lng: lon1,
		},
		IsElectronic:    false,
		GeocodingMethod: "manual",
		Confidence:      "low",
		Notes:           "Initial guess",
	}

	// Save
	err := repo.SaveJudgment(judgment)
	if err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	originalUpdatedAt := judgment.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	// Update
	lat2 := -34.8822400
	lon2 := -56.1529600
	judgment.Point.Lat = lat2
	judgment.Point.Lng = lon2
	judgment.Confidence = "high"
	judgment.Notes = "Corrected after review"

	err = repo.SaveJudgment(judgment)
	if err != nil {
		t.Fatalf("SaveJudgment() update error = %v", err)
	}

	// Retrieve updated
	dbID := 6
	location := "AV 8 DE OCTUBRE Y AV CENTENARIO"

	judgments, err := repo.ListJudgments(&dbID, &location, 1, 0)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(judgments) == 0 {
		t.Fatalf("ListJudgments() returned no judgment for dbID %d, location %s", dbID, location)
	}

	retrieved := judgments[0]

	if retrieved.Point.Lat != lat2 {
		t.Errorf("Latitude = %f, want %f", retrieved.Point.Lat, lat2)
	}

	if retrieved.Confidence != "high" {
		t.Errorf("Confidence = %s, want high", retrieved.Confidence)
	}

	if retrieved.Notes != "Corrected after review" {
		t.Errorf("Notes = %s, want 'Corrected after review'", retrieved.Notes)
	}

	if !retrieved.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be after original")
	}
}

func TestSaveElectronicJudgment(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	lat := -34.5912
	lon := -56.2629

	judgment := &Location{
		DbID:     65,
		Location: "RUTA 005 Y 038K131_D",
		Point: &spatial.Point{
			Lat: lat,
			Lng: lon,
		},
		IsElectronic:    true,
		GeocodingMethod: "radares_rutas",
		Confidence:      "high",
		Notes:           "Matched to Juanicó radar",
	}

	err := repo.SaveJudgment(judgment)
	if err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	dbID := 65
	location := "RUTA 005 Y 038K131_D"

	judgments, err := repo.ListJudgments(&dbID, &location, 1, 0)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(judgments) == 0 {
		t.Fatalf("ListJudgments() returned no judgment for dbID %d, location %s", dbID, location)
	}

	retrieved := judgments[0]

	if !retrieved.IsElectronic {
		t.Error("Expected IsElectronic to be true")
	}

	if retrieved.GeocodingMethod != "radares_rutas" {
		t.Errorf("GeocodingMethod = %s, want radares_rutas", retrieved.GeocodingMethod)
	}
}

func TestListJudgments(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Add multiple judgments
	judgments := []*Location{
		{DbID: 6, Location: "Location 1", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}},
		{DbID: 6, Location: "Location 2", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}},
		{DbID: 45, Location: "Location 3", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}},
	}

	for _, j := range judgments {
		if err := repo.SaveJudgment(j); err != nil {
			t.Fatalf("SaveJudgment() error = %v", err)
		}
	}

	// List all
	all, err := repo.ListJudgments(nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 judgments, got %d", len(all))
	}

	// List filtered by db_id
	dbID := 6

	filtered, err := repo.ListJudgments(&dbID, nil, 0, 0)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 judgments for db_id 6, got %d", len(filtered))
	}

	// Test pagination
	paginated, err := repo.ListJudgments(nil, nil, 2, 1)
	if err != nil {
		t.Fatalf("ListJudgments() error = %v", err)
	}

	if len(paginated) != 2 {
		t.Errorf("Expected 2 judgments with limit 2, got %d", len(paginated))
	}
}

func TestCountJudgments(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Initial count
	count, err := repo.CountJudgments()
	if err != nil {
		t.Fatalf("CountJudgments() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 judgments, got %d", count)
	}

	// Add judgments
	if err := repo.SaveJudgment(&Location{DbID: 6, Location: "Loc 1", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}}); err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	if err := repo.SaveJudgment(&Location{DbID: 6, Location: "Loc 2", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}}); err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	if err := repo.SaveJudgment(&Location{DbID: 45, Location: "Loc 3", GeocodingMethod: "manual", Point: &spatial.Point{Lat: 1, Lng: 1}}); err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	// Count all
	count, err = repo.CountJudgments()
	if err != nil {
		t.Fatalf("CountJudgments() error = %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 judgments, got %d", count)
	}
}

func TestJSONExportImport(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	tempFile := "/tmp/test_judgments.json"
	defer os.Remove(tempFile)

	// Add some judgments
	lat1 := -34.8822366
	lon1 := -56.1529602
	lat2 := -34.5912
	lon2 := -56.2629

	judgments := []*Location{
		{
			DbID:     6,
			Location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
			Point: &spatial.Point{
				Lat: lat1,
				Lng: lon1,
			},
			IsElectronic:    false,
			GeocodingMethod: "manual",
			Confidence:      "high",
			Notes:           "Test location 1",
		},
		{
			DbID:     65,
			Location: "RUTA 005 Y 038K131_D",
			Point: &spatial.Point{
				Lat: lat2,
				Lng: lon2,
			},
			IsElectronic:    true,
			GeocodingMethod: "radares_rutas",
			Confidence:      "high",
			Notes:           "Test location 2",
		},
	}

	for _, j := range judgments {
		if err := repo.SaveJudgment(j); err != nil {
			t.Fatalf("SaveJudgment() error = %v", err)
		}
	}

	// Export
	err := ExportToJSON(repo, tempFile)
	if err != nil {
		t.Fatalf("ExportToJSON() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Fatal("JSON file was not created")
	}

	// Create new database and import
	db2, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("Failed to open second test database: %v", err)
	}
	defer db2.Close()

	repo2 := NewLocationRepository(db2, map[int]string{})
	if err := repo2.CreateSchema(); err != nil {
		t.Fatalf("Failed to create schema in second database: %v", err)
	}

	imported, err := ImportFromJSON(repo2, tempFile)
	if err != nil {
		t.Fatalf("ImportFromJSON() error = %v", err)
	}

	if imported != 2 {
		t.Errorf("Expected 2 imported judgments, got %d", imported)
	}

	// Verify imported data
	dbID := 6
	location := "AV 8 DE OCTUBRE Y AV CENTENARIO"

	judgments, err = repo2.ListJudgments(&dbID, &location, 1, 0)
	if err != nil {
		t.Fatalf("ListJudgments() after import error = %v", err)
	}

	if len(judgments) == 0 {
		t.Fatalf("ListJudgments() returned no judgment for dbID %d, location %s", dbID, location)
	}

	retrieved := judgments[0]

	if retrieved.Location != "AV 8 DE OCTUBRE Y AV CENTENARIO" {
		t.Errorf("Imported location mismatch: got %s", retrieved.Location)
	}

	if retrieved.Point.Lat != lat1 {
		t.Errorf("Imported latitude mismatch: got %f, want %f", retrieved.Point.Lat, lat1)
	}
}

func TestSeedIfEmpty(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	tempFile := "/tmp/test_seed.json"
	defer os.Remove(tempFile)

	// Create seed file
	lat := -34.8822366
	lon := -56.1529602

	judgment := &Location{
		DbID:     6,
		Location: "Seed Location",
		Point: &spatial.Point{
			Lat: lat,
			Lng: lon,
		},
		GeocodingMethod: "manual",
	}
	if err := repo.SaveJudgment(judgment); err != nil {
		t.Fatalf("SaveJudgment() error = %v", err)
	}

	if err := ExportToJSON(repo, tempFile); err != nil {
		t.Fatalf("ExportToJSON() error = %v", err)
	}

	// Clear database
	if _, err := db.Exec("DELETE FROM locations"); err != nil {
		t.Fatalf("db.Exec() error = %v", err)
	}

	// Test seeding
	seeded, count, err := SeedIfEmpty(repo, tempFile)
	if err != nil {
		t.Fatalf("SeedIfEmpty() error = %v", err)
	}

	if !seeded {
		t.Error("Expected database to be seeded")
	}

	if count != 1 {
		t.Errorf("Expected 1 seeded judgment, got %d", count)
	}

	// Test that it doesn't seed again
	seeded, count, err = SeedIfEmpty(repo, tempFile)
	if err != nil {
		t.Fatalf("SeedIfEmpty() second call error = %v", err)
	}

	if seeded {
		t.Error("Expected database not to be seeded again")
	}

	if count != 1 {
		t.Errorf("Expected count to be 1 (existing), got %d", count)
	}
}

func TestMergeLocations(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// 1. Create canonical and target judgments
	canonicalJudgment := &Location{
		DbID:     1,
		Location: "Canonical Location",
		Point: &spatial.Point{
			Lat: 10.0,
			Lng: 20.0,
		},
		GeocodingMethod: "manual",
		Confidence:      "high",
	}
	if err := repo.SaveJudgment(canonicalJudgment); err != nil {
		t.Fatalf("Failed to save canonical judgment: %v", err)
	}

	targetJudgment := &Location{
		DbID:     1,
		Location: "Target Location",
		Point: &spatial.Point{
			Lat: 30.0,
			Lng: 40.0,
		},
		GeocodingMethod: "manual",
		Confidence:      "medium",
	}
	if err := repo.SaveJudgment(targetJudgment); err != nil {
		t.Fatalf("Failed to save target judgment: %v", err)
	}

	// 2. Call MergeLocations
	err := repo.MergeLocations(1, "Target Location", "Canonical Location")
	if err != nil {
		t.Fatalf("MergeLocations failed: %v", err)
	}

	// 3. Get the updated target judgment
	dbID := 1
	location := "Target Location"
	updatedTargets, err := repo.ListJudgments(&dbID, &location, 1, 0)
	updatedTarget := updatedTargets[0]

	if err != nil {
		t.Fatalf("Failed to get updated target judgment: %v", err)
	}

	// 4. Check if the CanonicalLocation field is set
	if updatedTarget.CanonicalLocation != "Canonical Location" {
		t.Errorf("Expected CanonicalLocation to be 'Canonical Location', got '%s'", updatedTarget.CanonicalLocation)
	}

	// 5. Check if the coordinates are updated
	if updatedTarget.Point.Lat != 10.0 || updatedTarget.Point.Lng != 20.0 {
		t.Errorf("Expected target coordinates to be (10.0, 20.0), got (%f, %f)", updatedTarget.Point.Lat, updatedTarget.Point.Lng)
	}
}
