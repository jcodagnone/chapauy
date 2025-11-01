// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/duckdb/duckdb-go/v2" // register duckdb driver
	"github.com/jcodagnone/chapauy/curation"
	"github.com/jcodagnone/chapauy/curation/utils"
	"github.com/jcodagnone/chapauy/impo"
	"github.com/spf13/cobra"
)

const judgmentsFile = "judgments.json"

type CurationData struct {
	Articles     []curation.Article      `json:"articles"`
	Descriptions []*curation.Description `json:"descriptions"`
	Locations    []*curation.Location    `json:"locations"`
}

var curationCmd = &cobra.Command{
	Use:   "curation",
	Short: "Manage the interactive curation workflow",
}

var curationServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the interactive geocoding web server (local only)",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := os.MkdirAll(impoOptions.DbPath, 0o750); err != nil {
			return fmt.Errorf("creating db directory: %w", err)
		}
		dbpath := filepath.Join(impoOptions.DbPath, "chapauy.duckdb")

		if _, err := os.Stat(dbpath); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("database not found at %s - run 'seed' or 'impo update' first", dbpath)
		}

		db, err := sql.Open("duckdb", dbpath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		// Build DB map
		dbMap := make(map[int]string)
		if err := impo.Each(func(ref impo.DbReference) error {
			dbMap[ref.ID] = ref.Name

			return nil
		}); err != nil {
			return fmt.Errorf("building db map: %w", err)
		}

		locRepo := curation.NewLocationRepository(db, dbMap)
		if err := locRepo.CreateSchema(); err != nil {
			return fmt.Errorf("creating geocoding schema: %w", err)
		}

		// Load radar index
		radarIndex, err := curation.LoadRadares("curation/radares.json")
		if err != nil {
			return fmt.Errorf("loading radares: %w", err)
		}

		descrRepo := curation.NewDescriptionRepository(db)
		if err := descrRepo.CreateSchema(); err != nil {
			return fmt.Errorf("creating description schema: %w", err)
		}

		server := curation.NewServer(
			locRepo,
			db, // Pass db directly
			radarIndex,
			dbMap,
		)

		fmt.Println("🗺️  Geocoding workflow server starting...")
		fmt.Println("📍 Open http://localhost:8080 in your browser")
		fmt.Println("🔒 Local only - not exposed to internet")

		return server.Run()
	},
}

var curationStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Export geocoding judgments to a file",
	Long:  `Exports all location judgments from the database to a local JSON file. The file is sorted to minimize diffs when checking into version control.`,
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		dbpath := filepath.Join(impoOptions.DbPath, "chapauy.duckdb")
		db, err := sql.Open("duckdb", dbpath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		repo := curation.NewLocationRepository(db, nil)
		locations, err := repo.GetAllJudgmentsSorted()
		if err != nil {
			return fmt.Errorf("getting location judgments: %w", err)
		}

		descrRepo := curation.NewDescriptionRepository(db)
		descriptions, err := descrRepo.GetAllDescriptionJudgmentsSorted()
		if err != nil {
			return fmt.Errorf("getting description judgments: %w", err)
		}

		articles, err := descrRepo.ListArticles()
		if err != nil {
			return fmt.Errorf("getting articles: %w", err)
		}

		data, err := json.MarshalIndent(
			CurationData{
				Articles:     articles,
				Descriptions: descriptions,
				Locations:    locations,
			},
			"",
			"  ",
		)
		if err != nil {
			return fmt.Errorf("marshaling curation data: %w", err)
		}

		if err := os.WriteFile(judgmentsFile, data, 0o600); err != nil {
			return fmt.Errorf("writing judgments file: %w", err)
		}

		fmt.Printf("✅ Exported %s location judgments, %s description judgments, and %s articles to %s\n",
			utils.FormatInt(int64(len(locations))),
			utils.FormatInt(int64(len(descriptions))),
			utils.FormatInt(int64(len(articles))),
			judgmentsFile)

		return nil
	},
}

var curationLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Import geocoding judgments from a file and backfill offenses",
	Long: `Imports judgments from the local JSON file into the database if the judgments table is empty.
After importing, it updates the offenses table with the geocoding information.`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		dbpath := filepath.Join(impoOptions.DbPath, "chapauy.duckdb")
		db, err := sql.Open("duckdb", dbpath)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		if err := ensureCurationDataLoaded(db); err != nil {
			return err
		}

		return backfillCurationData(db)
	},
}

func ensureCurationDataLoaded(db *sql.DB) error {
	locRepo := curation.NewLocationRepository(db, nil)
	if err := locRepo.CreateSchema(); err != nil {
		return fmt.Errorf("creating geocoding schema: %w", err)
	}

	descrRepo := curation.NewDescriptionRepository(db)
	if err := descrRepo.CreateSchema(); err != nil {
		return fmt.Errorf("creating description schema: %w", err)
	}

	// Try to read from the primary judgments file path, but fall back to the
	// secondary path for backward compatibility.
	data, err := os.ReadFile(judgmentsFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("reading judgments file: %w", err)
		}

		return fmt.Errorf("could not find judgments file at %s: %w", judgmentsFile, err)
	}

	var curationData CurationData
	if err := json.Unmarshal(data, &curationData); err != nil {
		return fmt.Errorf("unmarshaling curation data: %w", err)
	}

	targetLocCount := len(curationData.Locations)
	targetDescrCount := len(curationData.Descriptions)
	targetArtCount := len(curationData.Articles)

	// Check DB state
	var (
		dbLocCount   int
		dbDescrCount int
		dbArtCount   int
	)

	query := `
		SELECT
			(SELECT count(*) FROM locations) as loc_count,
			(SELECT count(*) FROM descriptions) as desc_count,
			(SELECT count(*) FROM articles) as art_count
	`
	if err := db.QueryRow(query).Scan(
		&dbLocCount,
		&dbDescrCount,
		&dbArtCount,
	); err != nil {
		return fmt.Errorf("checking db state: %w", err)
	}

	// Safety Check: Do not overwrite if DB has MORE data than the file.
	// This likely means there are local judgments/curation that haven't been exported yet.
	isUnsafe := false

	if dbLocCount > targetLocCount {
		log.Printf("⚠️  Local location judgments (%d) exceed file counts (%d). Unsaved work detected.", dbLocCount, targetLocCount)

		isUnsafe = true
	}

	if dbDescrCount > targetDescrCount {
		log.Printf("⚠️  Local description judgments (%d) exceed file counts (%d). Unsaved work detected.", dbDescrCount, targetDescrCount)

		isUnsafe = true
	}

	if dbArtCount > targetArtCount {
		log.Printf("⚠️  Local articles (%d) exceed file counts (%d). Unsaved work detected.", dbArtCount, targetArtCount)

		isUnsafe = true
	}

	if isUnsafe {
		log.Println("🛑 Skipping reload to prevent data loss. Run 'curation store' to save local changes first.")

		return nil
	}

	// Freshness Check: Reload only if the file has MORE data than the DB.
	needsReload := false

	if targetLocCount > dbLocCount {
		log.Printf("ℹ️  New location judgments available (%d > %d). Reloading...", targetLocCount, dbLocCount)

		needsReload = true
	} else if targetDescrCount > dbDescrCount {
		log.Printf("ℹ️  New description judgments available (%d > %d). Reloading...", targetDescrCount, dbDescrCount)

		needsReload = true
	} else if targetArtCount > dbArtCount {
		log.Printf("ℹ️  New articles available (%d > %d). Reloading...", targetArtCount, dbArtCount)

		needsReload = true
	}

	if !needsReload {
		log.Println("✅ Curation data is up to date. Skipping import.")

		return nil
	}

	log.Println("♻️  Reloading curation data...")

	// Clear tables
	if _, err := db.Exec("DELETE FROM locations"); err != nil {
		return fmt.Errorf("clearing locations: %w", err)
	}

	if _, err := db.Exec("DELETE FROM descriptions"); err != nil {
		return fmt.Errorf("clearing descriptions: %w", err)
	}

	if _, err := db.Exec("DELETE FROM articles"); err != nil {
		return fmt.Errorf("clearing articles: %w", err)
	}

	// Load Location Judgments
	if err := locRepo.BulkInsertJudgments(curationData.Locations); err != nil {
		return fmt.Errorf("inserting location judgments: %w", err)
	}

	log.Printf("✅ Imported %s location judgments from %s\n", utils.FormatInt(int64(len(curationData.Locations))), judgmentsFile)

	// Load Articles
	if err := descrRepo.SeedArticles(curationData.Articles); err != nil {
		return fmt.Errorf("seeding articles: %w", err)
	}

	log.Printf("✅ Imported %s articles from %s\n", utils.FormatInt(int64(len(curationData.Articles))), judgmentsFile)

	// Load Description Judgments
	if err := descrRepo.BulkInsertDescriptionJudgments(curationData.Descriptions); err != nil {
		return fmt.Errorf("inserting description judgments: %w", err)
	}

	log.Printf("✅ Imported %s description judgments from %s\n", utils.FormatInt(int64(len(curationData.Descriptions))), judgmentsFile)

	return nil
}

func backfillCurationData(db *sql.DB) error {
	repo, err := impo.NewSQLOffenseRepository(db)
	if err != nil {
		return fmt.Errorf("initializing repository: %w", err)
	}

	affected, err := repo.BackfillGeocodingData()
	if err != nil {
		return fmt.Errorf("backfilling geocoding data: %w", err)
	}

	var pendingGeocodingOffenses int

	var pendingGeocodingLocations int

	geoQuery := `
		SELECT
			COUNT(*),
			COUNT(DISTINCT location)
		FROM offenses
		WHERE point IS NULL
		AND location IS NOT NULL
		AND location != ''
	`
	if err := db.QueryRow(geoQuery).Scan(&pendingGeocodingOffenses, &pendingGeocodingLocations); err != nil {
		return fmt.Errorf("counting pending geocoding: %w", err)
	}

	log.Printf("✅ Backfilled %s offenses with geocoding data (%s pending offenses, %s unique locations)\n",
		utils.FormatInt(affected),
		utils.FormatInt(int64(pendingGeocodingOffenses)),
		utils.FormatInt(int64(pendingGeocodingLocations)))

	affected, err = repo.BackportDescriptionArticles()
	if err != nil {
		return fmt.Errorf("backporting curation data: %w", err)
	}

	var pendingOffenses int

	var pendingDescriptions int

	query := `
		SELECT
			COUNT(*),
			COUNT(DISTINCT description)
		FROM offenses
		WHERE article_ids IS NULL
		AND description IS NOT NULL
		AND description != ''
	`
	if err := db.QueryRow(query).Scan(&pendingOffenses, &pendingDescriptions); err != nil {
		return fmt.Errorf("counting pending offenses: %w", err)
	}

	log.Printf("✅ Backfilled %s offenses with description articles (%s pending offenses, %s unique descriptions)\n",
		utils.FormatInt(affected),
		utils.FormatInt(int64(pendingOffenses)),
		utils.FormatInt(int64(pendingDescriptions)))

	return nil
}

func init() {
	rootCmd.AddCommand(curationCmd)
	curationCmd.AddCommand(curationServeCmd)
	curationCmd.AddCommand(curationStoreCmd)
	curationCmd.AddCommand(curationLoadCmd)
}
