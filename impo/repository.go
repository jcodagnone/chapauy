// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jcodagnone/chapauy/curation/utils"
	"github.com/jcodagnone/chapauy/spatial"
)

// OffenseRepository defines the interface for database operations.
type OffenseRepository interface {
	//////// Extraction
	// LoadCaches eagerly loads necessary caches (e.g. location, description) for enrichment.
	LoadCaches() error
	// CreateSchema creates the database schema.
	CreateSchema() error
	// SaveTrafficOffenses saves a list of traffic offenses to the database.
	SaveTrafficOffenses(offenses []*TrafficOffense) error
	// GetExtractedDocuments returns a list of all the documents that have been extracted.
	GetExtractedDocuments(db *DbReference) (map[string]bool, error)

	//////// Geocoding Integration
	// BackfillGeocodingData updates offenses with geocoding data from location_judgments table
	BackfillGeocodingData() (int64, error)
	// BackportDescriptionArticles updates offenses with curated article and section data
	BackportDescriptionArticles() (int64, error)
}

// ArticleLabel represents a label for an article.
type ArticleLabel struct {
	Label      string
	Normalized string
}

type locationData struct {
	CanonicalLocation string
	DisplayLocation   string
	Point             spatial.Point
	H3Res1            uint64
	H3Res2            uint64
	H3Res3            uint64
	H3Res4            uint64
	H3Res5            uint64
	H3Res6            uint64
	H3Res7            uint64
	H3Res8            uint64
}

type descriptionData struct {
	ArticleIDs   []string
	ArticleCodes []int8
}

type locationKey struct {
	DbID     int
	Location string
}

type sqlOffenseRepository struct {
	db *sql.DB
	// Cache for article labels (ID -> ArticleLabel)
	articleCache map[string]ArticleLabel
	// Cache for article code labels (Code -> ArticleLabel)
	articleCodeCache map[string]ArticleLabel
	// Cache for location data
	locationCache map[locationKey]locationData
	// Cache for description data
	descriptionCache map[string]descriptionData
}

func NewSQLOffenseRepository(db *sql.DB) (OffenseRepository, error) {
	// DuckDB needs to load the spatial extension
	_, err := db.Exec(`INSTALL spatial; LOAD spatial;`)
	if err != nil {
		return nil, err
	}

	repo := &sqlOffenseRepository{db: db}
	repo.loadArticleCache()

	return repo, nil
}

// loadArticleCache loads article data from the database into memory caches
// to provide labels for article IDs and codes in dimension results.
func (r *sqlOffenseRepository) loadArticleCache() {
	r.articleCache = make(map[string]ArticleLabel)
	r.articleCodeCache = make(map[string]ArticleLabel)

	rows, err := r.db.Query("SELECT id, text, code, title FROM articles")
	if err != nil {
		// Log error but don't fail, just continue without cache
		fmt.Printf("Error loading articles cache: %v\n", err)

		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, text, title string

		var code int64
		if err := rows.Scan(&id, &text, &code, &title); err != nil {
			continue
		}

		labelID := fmt.Sprintf("%s - %s", id, text)
		r.articleCache[id] = ArticleLabel{
			Label:      labelID,
			Normalized: utils.LowerASCIIFolding(labelID),
		}

		labelCode := fmt.Sprintf("%d - %s", code, title)
		r.articleCodeCache[strconv.FormatInt(code, 10)] = ArticleLabel{
			Label:      labelCode,
			Normalized: utils.LowerASCIIFolding(labelCode),
		}
	}
}

func (r *sqlOffenseRepository) LoadCaches() error {
	if err := r.loadLocationCache(); err != nil {
		return err
	}

	if err := r.loadDescriptionCache(); err != nil {
		return err
	}

	return nil
}

func (r *sqlOffenseRepository) loadLocationCache() error {
	r.locationCache = make(map[locationKey]locationData)

	rows, err := r.db.Query(`
		SELECT
			db_id, location, canonical_location, point,
			h3_res1, h3_res2, h3_res3, h3_res4,
			h3_res5, h3_res6, h3_res7, h3_res8
		FROM locations
		WHERE canonical_location IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("querying locations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var k locationKey

		var d locationData

		if err := rows.Scan(
			&k.DbID, &k.Location, &d.CanonicalLocation, &d.Point,
			&d.H3Res1, &d.H3Res2, &d.H3Res3, &d.H3Res4,
			&d.H3Res5, &d.H3Res6, &d.H3Res7, &d.H3Res8,
		); err != nil {
			return fmt.Errorf("scanning location: %w", err)
		}

		d.DisplayLocation = k.Location
		r.locationCache[k] = d
	}

	return nil
}

func (r *sqlOffenseRepository) loadDescriptionCache() error {
	r.descriptionCache = make(map[string]descriptionData)

	rows, err := r.db.Query("SELECT description, article_ids, article_codes FROM descriptions")
	if err != nil {
		return fmt.Errorf("querying descriptions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var desc string

		var d descriptionData

		var idsVal, codesVal any

		if err := rows.Scan(&desc, &idsVal, &codesVal); err != nil {
			return fmt.Errorf("scanning description: %w", err)
		}

		if ids, ok := utils.AnyToStringSlice(idsVal); ok {
			d.ArticleIDs = ids
		}

		if codes, ok := utils.AnyToInt8Slice(codesVal); ok {
			d.ArticleCodes = codes
		}

		r.descriptionCache[utils.LowerASCIIFolding(desc)] = d
	}

	return nil
}

func (r *sqlOffenseRepository) CreateSchema() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS offenses (
			db_id INTEGER NOT NULL,
			doc_id VARCHAR,
			doc_date DATE,
			doc_source VARCHAR NOT NULL,
			record_id INTEGER NOT NULL,
			offense_id VARCHAR,
			vehicle VARCHAR,
			vehicle_country CHAR(2),
			vehicle_type VARCHAR,
			"time" TIMESTAMPTZ,
			time_year USMALLINT,
			location VARCHAR,
			display_location VARCHAR,
			description VARCHAR,
			ur INTEGER,
			error VARCHAR,
			point POINT_2D,
			h3_res1 UBIGINT,
			h3_res2 UBIGINT,
			h3_res3 UBIGINT,
			h3_res4 UBIGINT,
			h3_res5 UBIGINT,
			h3_res6 UBIGINT,
			h3_res7 UBIGINT,
			h3_res8 UBIGINT
		);

		ALTER TABLE offenses ADD COLUMN IF NOT EXISTS article_ids VARCHAR[];
		ALTER TABLE offenses ADD COLUMN IF NOT EXISTS article_codes TINYINT[];

	`)

	return err
}

func (r *sqlOffenseRepository) GetExtractedDocuments(db *DbReference) (map[string]bool, error) {
	rows, err := r.db.Query("SELECT DISTINCT doc_source FROM offenses WHERE db_id = ?", db.ID)
	if err != nil {
		return nil, fmt.Errorf("querying existing documents: %w", err)
	}
	defer rows.Close()

	existingDocs := make(map[string]bool)

	for rows.Next() {
		var docSource string
		if err := rows.Scan(&docSource); err != nil {
			return nil, fmt.Errorf("scanning existing document: %w", err)
		}

		existingDocs[docSource] = true
	}

	return existingDocs, nil
}

func nve(v string) any {
	var ret any
	if len(v) == 0 {
		ret = nil
	} else {
		ret = v
	}

	return ret
}

func (r *sqlOffenseRepository) enrichOffense(o *TrafficOffense) {
	// 1. Geocoding
	if o.Location != "" {
		key := locationKey{DbID: o.DbID, Location: o.Location}
		if locData, ok := r.locationCache[key]; ok {
			o.Point = &locData.Point
			o.H3Res1 = locData.H3Res1
			o.H3Res2 = locData.H3Res2
			o.H3Res3 = locData.H3Res3
			o.H3Res4 = locData.H3Res4
			o.H3Res5 = locData.H3Res5
			o.H3Res6 = locData.H3Res6
			o.H3Res7 = locData.H3Res7
			o.H3Res8 = locData.H3Res8

			if locData.CanonicalLocation != "" {
				o.Location = locData.CanonicalLocation
				o.DisplayLocation = locData.DisplayLocation
			}
		}
	}

	// 2. Description / Articles
	if o.Description != "" {
		normDesc := utils.LowerASCIIFolding(o.Description)
		if data, ok := r.descriptionCache[normDesc]; ok {
			o.ArticleIDs = data.ArticleIDs
			o.ArticleCodes = data.ArticleCodes
		} else if strings.Contains(o.Description, ",") {
			classify := func(part string) (utils.Classification, bool, error) {
				normPart := utils.LowerASCIIFolding(part)
				if info, ok := r.descriptionCache[normPart]; ok {
					return utils.Classification{
						ArticleIDs:   info.ArticleIDs,
						ArticleCodes: info.ArticleCodes,
					}, true, nil
				}

				return utils.Classification{}, false, nil
			}

			result, found, _ := utils.ResolveMultiArticle(o.Description, classify)
			if found {
				o.ArticleIDs = result.ArticleIDs
				o.ArticleCodes = result.ArticleCodes
			}
		}
	}
}

func nz(v uint64) any {
	if v == 0 {
		return nil
	}

	return v
}

func (r *sqlOffenseRepository) SaveTrafficOffenses(offenses []*TrafficOffense) error {
	if len(offenses) == 0 {
		return nil
	}

	// Caches should be loaded via LoadCaches() at startup.
	// If caches are nil, enrichment will simply be skipped for those parts.

	for _, o := range offenses {
		r.enrichOffense(o)
	}

	docSource := offenses[0].DocSource
	tx, err := r.db.Begin()

	if err != nil {
		return fmt.Errorf("starting transaction for %s: %w", docSource, err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction for %s: %v", docSource, err)
		}
	}()

	if _, err := tx.Exec("DELETE FROM offenses WHERE doc_source = ?", docSource); err != nil {
		return fmt.Errorf("deleting records for %s: %w", docSource, err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO offenses (
			db_id, doc_id, doc_date, doc_source, record_id, offense_id,
			vehicle, vehicle_country, vehicle_type, time, time_year, location, display_location, description, ur, error,
			point,
			h3_res1, h3_res2, h3_res3, h3_res4, h3_res5, h3_res6, h3_res7, h3_res8,
			article_ids, article_codes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, EXTRACT(YEAR FROM ?::TIMESTAMPTZ), ?, ?, ?, ?, ?, ST_Point(?, ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range offenses {
		var countryHint string
		if record.VehicleInfo != nil {
			countryHint = record.VehicleInfo.Country
		}

		info, _ := AnalyzeVehicleID(record.Vehicle, countryHint)

		var vehicleType sql.NullString
		if info.VehicleType != "" {
			vehicleType.String = info.VehicleType
			vehicleType.Valid = true
		}

		var offenseError sql.NullString
		if record.Error != "" {
			offenseError.String = record.Error
			offenseError.Valid = true
		}

		var lng, lat any
		if record.Point != nil {
			lng = record.Point.Lng
			lat = record.Point.Lat
		}

		_, err := stmt.Exec(
			record.DbID,
			record.DocID,
			record.DocDate,
			record.DocSource,
			record.RecordID,
			record.ID,
			record.Vehicle,
			nve(info.Country),
			vehicleType,
			record.Time,
			record.Time, // For time_year extraction
			nve(record.Location),
			nve(record.DisplayLocation),
			nve(record.Description),
			record.UR,
			offenseError,
			lng,
			lat,
			nz(record.H3Res1),
			nz(record.H3Res2),
			nz(record.H3Res3),
			nz(record.H3Res4),
			nz(record.H3Res5),
			nz(record.H3Res6),
			nz(record.H3Res7),
			nz(record.H3Res8),
			record.ArticleIDs,
			record.ArticleCodes,
		)
		if err != nil {
			return fmt.Errorf("inserting record for %s: %w", docSource, err)
		}
	}

	return tx.Commit()
}

func (r *sqlOffenseRepository) BackfillGeocodingData() (int64, error) {
	var n int64

	for _, q := range []string{
		// first we apply the canonical names
		`
		UPDATE offenses
		SET
			location = lj.canonical_location,
			display_location = lj.location
		FROM
			locations lj
		WHERE
		        lj.canonical_location IS NOT NULL
			AND offenses.db_id = lj.db_id
			AND offenses.location = lj.location
			AND offenses.display_location IS NULL
		`,
		// then we apply the geocoding information
		`
			UPDATE offenses
			SET
				point = lj.point,
				h3_res1 = lj.h3_res1,
				h3_res2 = lj.h3_res2,
				h3_res3 = lj.h3_res3,
				h3_res4 = lj.h3_res4,
				h3_res5 = lj.h3_res5,
				h3_res6 = lj.h3_res6,
				h3_res7 = lj.h3_res7,
				h3_res8 = lj.h3_res8
			FROM
				locations lj
			WHERE
				offenses.db_id = lj.db_id
				AND offenses.location = lj.location
				AND offenses.point IS NULL
		`,
	} {
		result, err := r.db.Exec(q)
		if err != nil {
			return n, fmt.Errorf("backfilling geocoding data: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return n, fmt.Errorf("getting rows affected: %w", err)
		}

		n += rowsAffected
	}

	return n, nil
}

// BackportDescriptionArticles updates offenses with curated article and section data.
func (r *sqlOffenseRepository) BackportDescriptionArticles() (int64, error) {
	var totalRowsAffected int64

	queries := []string{
		// 1. Update single-part descriptions (and direct multi-part matches)
		`
		UPDATE offenses
		SET
			article_ids = d.article_ids,
			article_codes = d.article_codes
		FROM descriptions d
		WHERE
			offenses.article_ids IS NULL
			AND offenses.description IS NOT NULL
			AND offenses.description = d.description
		`,
	}

	for _, query := range queries {
		result, err := r.db.Exec(query)
		if err != nil {
			return totalRowsAffected, fmt.Errorf("backporting curation data: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalRowsAffected, fmt.Errorf("getting rows affected: %w", err)
		}

		totalRowsAffected += rowsAffected
	}

	// 2. Update multi-article descriptions
	multiAffected, err := r.backportMultiArticleDescriptions()
	if err != nil {
		return totalRowsAffected, fmt.Errorf("backporting multi-article descriptions: %w", err)
	}

	totalRowsAffected += multiAffected

	return totalRowsAffected, nil
}

func (r *sqlOffenseRepository) backportMultiArticleDescriptions() (int64, error) {
	// 1. Load all classified descriptions into memory
	rows, err := r.db.Query("SELECT description, article_ids, article_codes FROM descriptions")
	if err != nil {
		return 0, fmt.Errorf("loading descriptions: %w", err)
	}
	defer rows.Close()

	// Map normalized description -> info
	type descInfo struct {
		ids   []string
		codes []int8
	}

	knownDescriptions := make(map[string]descInfo)

	for rows.Next() {
		var d string

		var idsVal, codesVal any
		if err := rows.Scan(&d, &idsVal, &codesVal); err != nil {
			return 0, fmt.Errorf("scanning description: %w", err)
		}

		ids, ok := utils.AnyToStringSlice(idsVal)
		if !ok {
			continue
		}

		codes, ok := utils.AnyToInt8Slice(codesVal)
		if !ok {
			continue
		}

		norm := utils.LowerASCIIFolding(d)
		knownDescriptions[norm] = descInfo{ids: ids, codes: codes}
	}

	// 2. Get pending multi-article descriptions
	// We fetch descriptions that contain a comma and are not yet backported.
	pendingQuery := `
		SELECT DISTINCT description
		FROM offenses
		WHERE article_ids IS NULL
		AND description LIKE '%,%'
	`

	pendingRows, err := r.db.Query(pendingQuery)
	if err != nil {
		return 0, fmt.Errorf("getting pending descriptions: %w", err)
	}

	defer pendingRows.Close()

	var pending []string

	for pendingRows.Next() {
		var desc string
		if err := pendingRows.Scan(&desc); err != nil {
			return 0, fmt.Errorf("scanning pending description: %w", err)
		}

		pending = append(pending, desc)
	}

	// 3. Process each pending description
	var backportedCount int64

	updateQuery := `
		UPDATE offenses
		SET article_ids = ?, article_codes = ?
		WHERE description = ?
	`

	// Define classifier closure
	classify := func(part string) (utils.Classification, bool, error) {
		normPart := utils.LowerASCIIFolding(part)

		info, ok := knownDescriptions[normPart]
		if !ok {
			return utils.Classification{}, false, nil
		}

		return utils.Classification{
			ArticleIDs:   info.ids,
			ArticleCodes: info.codes,
		}, true, nil
	}

	for _, desc := range pending {
		result, found, err := utils.ResolveMultiArticle(desc, classify)
		if err != nil {
			return backportedCount, fmt.Errorf("resolving multi-article description %q: %w", desc, err)
		}

		if found && len(result.ArticleIDs) > 0 {
			// Update the offense with the aggregated articles
			if _, err := r.db.Exec(updateQuery, result.ArticleIDs, result.ArticleCodes, desc); err != nil {
				return backportedCount, fmt.Errorf("updating offense %q: %w", desc, err)
			}

			backportedCount++
		}
	}

	return backportedCount, nil
}
