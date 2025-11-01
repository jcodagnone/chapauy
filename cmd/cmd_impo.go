// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2" // register duckdb driver
	"github.com/jcodagnone/chapauy/impo"
	"github.com/spf13/cobra"
)

var impoCmd = &cobra.Command{
	Use:   "impo",
	Short: "Acceso a las base de datos",
}

var impoListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista las base de datos disponibles",
	RunE: func(_ *cobra.Command, _ []string) error {
		a, b, c := strings.Repeat("─", 2), strings.Repeat("─", 14), strings.Repeat("─", 60)
		fmt.Println("Base de datos disponibles:")
		fmt.Printf("╭─%2s─┬─%-14s─┬─%-60s╮\n", a, b, c)
		fmt.Printf("│ %2s │ %-14s │ %-60s│\n", "Id", "Nombre", "Ubicación")
		fmt.Printf("├─%2s─┼─%-14s─┼─%-60s┤\n", a, b, c)
		err := impo.Each(func(db impo.DbReference) error {
			fmt.Printf("│ %2d │ %-14s │ %-60s│\n", db.ID, db.Name, db.SeedURL)

			return nil
		})
		fmt.Printf("╰─%2s─┴─%-14s─┴─%-60s╯\n", a, b, c)

		return err
	},
}
var impoOptions = &impo.ClientOptions{}

func dbArg(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}

		if _, err := impo.Find(args[0]); err != nil {
			return err
		}
	}

	return nil
}

var impoUpdateCmd = &cobra.Command{
	Use:   "update <db>",
	Short: "Actualiza el contenido local para una base de datos",
	Args:  dbArg,
	RunE: func(_ *cobra.Command, args []string) error {
		var metrics impo.ClientMetrics
		var err error

		db, err := sql.Open("duckdb", filepath.Join(impoOptions.DbPath, "chapauy.duckdb"))
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Close()

		if err := ensureCurationDataLoaded(db); err != nil {
			return fmt.Errorf("loading curation data: %w", err)
		}

		repo, err := impo.NewSQLOffenseRepository(db)
		if err != nil {
			return fmt.Errorf("initializing repository: %w", err)
		}
		if err := repo.CreateSchema(); err != nil {
			return fmt.Errorf("creating table: %w", err)
		}

		if err := repo.LoadCaches(); err != nil {
			// It's acceptable if caches fail to load (e.g. tables don't exist yet),
			// enrichment will just be skipped.
			// However, since we just created schema (or ensured it exists),
			// failure here might indicate a real issue or empty tables.
			// Given the user's request "if wasn't called save would ignore the filling",
			// we can log a warning or just proceed.
			// Let's return error to be safe, or log.
			// The user said "if wasn't called save would ignore the filling".
			// So if LoadCaches fails, we should probably just log and continue?
			// But LoadCaches returns error.
			// Let's assume we want to fail if something is wrong, but maybe not if tables are missing?
			// But CreateSchema ensures tables exist (at least offenses).
			// Curation tables might be missing if not loaded.
			// loadLocationCache queries `locations` table.
			// If `locations` table doesn't exist, `loadLocationCache` will fail.
			// So we should probably ignore error if it's about missing table?
			// Or better: ensureCurationDataLoaded ensures tables exist.
			// So LoadCaches should succeed.
			return fmt.Errorf("loading caches: %w", err)
		}

		if len(args) == 0 {
			err = impo.Each(func(db impo.DbReference) error {
				impoOptions.UserAgent = fmt.Sprintf("chapauy/%s (+https://github.com/jcodagnone/chapauy)", Version)
				c := impo.NewImpoClient(impoOptions, &db, repo)
				err = c.Update()
				metrics.Merge(&c.Metrics)

				return err
			})
		} else {
			db, er := impo.Find(args[0])
			if er != nil {
				return er
			}
			impoOptions.UserAgent = fmt.Sprintf("chapauy/%s (+https://github.com/jcodagnone/chapauy)", Version)
			c := impo.NewImpoClient(impoOptions, db, repo)
			err = c.Update()
			metrics.Merge(&c.Metrics)
		}
		if !impoOptions.SkipSearch {
			log.Printf(
				"Total search phase metrics - %d new records from a total of %d records across %d pages",
				metrics.SearchTotalStored,
				metrics.SearchTotalRecords,
				metrics.SearchPages,
			)
		}
		if !impoOptions.SkipDownload {
			log.Printf(
				"Total download phase metrics - %d successful, %d failed",
				metrics.DownloadsOk,
				metrics.DownloadsErr,
			)
		}
		if !impoOptions.SkipExtract {
			log.Printf(
				"Total extraction phase metrics - %d new records, %d errors from %d documents, %d successful and %d failed.",
				metrics.NewRecords,
				metrics.NewErrors,
				metrics.SuccessfulDocs+metrics.FailedDocs,
				metrics.SuccessfulDocs,
				metrics.FailedDocs,
			)
		}

		if err == nil {
			if bfErr := backfillCurationData(db); bfErr != nil {
				return fmt.Errorf("backfilling curation data: %w", bfErr)
			}
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(impoCmd)
	impoCmd.AddCommand(impoListCmd)
	impoCmd.AddCommand(impoUpdateCmd)
	impoCmd.PersistentFlags().StringVar(
		&impoOptions.DbPath,
		"db-path",
		"db",
		"Directorio base donde almacenar el estado",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.SkipSearch,
		"skip-search",
		false,
		"Evita la fase de descubrimiento de nuevos documentos",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.SearchFull,
		"search-full",
		false,
		"Al descubrir nuevos documentos, transita por todas las páginas de la búsqueda",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.SkipDownload,
		"skip-download",
		false,
		"Evita la fase de descarga de documentos faltantes",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.SkipExtract,
		"skip-extract",
		false,
		"Evita la fase de extracción de datos de los documentos descargados",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.ExtractFull,
		"extract-full",
		false,
		"En la fase de extracción, procesa todos los documentos y no solo los pendientes",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.SkipErrDocs,
		"skip-extract-errors",
		false,
		"En la fase de extracción, evita almacenar documentos con al menos un error",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.DryRun,
		"dry-run",
		false,
		"No persiste ningun cambio",
	)

	impoUpdateCmd.PersistentFlags().IntVar(
		&impoOptions.SearchDepth,
		"search-max-depth",
		25,
		"En la fase de descubrimento, el número de páginas máximo a seguir",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.EnableHTTPTrace,
		"trace-http",
		false,
		"Display HTTP requests-responses",
	)
	impoUpdateCmd.PersistentFlags().BoolVar(
		&impoOptions.EnableHTTPBodyTrace,
		"trace-http-body",
		false,
		"Display HTTP requests-responses bodies",
	)
	impoUpdateCmd.PersistentFlags().IntVar(
		&impoOptions.ExtractMaxProcs,
		"extract-max-procs",
		0,
		"Max number of processes to use in the extraction phase. Defaults to the number of CPUs",
	)
}
