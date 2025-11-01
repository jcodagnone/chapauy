// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jcodagnone/chapauy/spatial"
	"github.com/uber/h3-go/v4"
)

// Location represents a geocoding decision made by a user.
type Location struct {
	DbID              int            `json:"db_id"`
	Location          string         `json:"location"`
	Point             *spatial.Point `json:"point"`
	IsElectronic      bool           `json:"is_electronic"`
	GeocodingMethod   string         `json:"geocoding_method"` // radares_rutas, google_maps, manual
	Confidence        string         `json:"confidence"`       // high, medium, low
	Notes             string         `json:"notes"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CanonicalLocation string         `json:"canonical_location,omitempty"`
	H3Res1            int64          `json:"-"`
	H3Res2            int64          `json:"-"`
	H3Res3            int64          `json:"-"`
	H3Res4            int64          `json:"-"`
	H3Res5            int64          `json:"-"`
	H3Res6            int64          `json:"-"`
	H3Res7            int64          `json:"-"`
	H3Res8            int64          `json:"-"`
}

func (judgment *Location) computeH3() error {
	if judgment.Point != nil {
		latLng := h3.NewLatLng(judgment.Point.Lat, judgment.Point.Lng)
		for res := 1; res <= 8; res++ {
			cell, err := h3.LatLngToCell(latLng, res)
			if err != nil {
				return fmt.Errorf("error converting to h3 cell at res %d: %w", res, err)
			}

			switch res {
			case 1:
				judgment.H3Res1 = int64(cell)
			case 2:
				judgment.H3Res2 = int64(cell)
			case 3:
				judgment.H3Res3 = int64(cell)
			case 4:
				judgment.H3Res4 = int64(cell)
			case 5:
				judgment.H3Res5 = int64(cell)
			case 6:
				judgment.H3Res6 = int64(cell)
			case 7:
				judgment.H3Res7 = int64(cell)
			case 8:
				judgment.H3Res8 = int64(cell)
			}
		}
	} else {
		judgment.H3Res1 = 0
		judgment.H3Res2 = 0
		judgment.H3Res3 = 0
		judgment.H3Res4 = 0
		judgment.H3Res5 = 0
		judgment.H3Res6 = 0
		judgment.H3Res7 = 0
		judgment.H3Res8 = 0
	}

	return nil
}

// Location represents a single location with its description and point.
type ClusterLocation struct {
	DbID                  int           `json:"db_id"`
	Description           string        `json:"description"`
	Point                 spatial.Point `json:"point"`
	OffenseCount          int           `json:"offense_count"`
	DistanceFromPrincipal float64       `json:"distance_from_principal"`
	IsPrincipal           bool          `json:"is_principal"`
}

// LocationCluster represents a group of similar locations.
type LocationCluster struct {
	DbID          int                `json:"db_id"`
	Location      string             `json:"location"`
	DbName        string             `json:"db_name"`
	TotalOffenses int                `json:"total_offenses"`
	Locations     []*ClusterLocation `json:"locations"`
}

// LocationRepository handles persistence of location judgments.
type LocationRepository interface {
	// CreateSchema creates the locations table
	CreateSchema() error

	// SaveJudgment saves or updates a location judgment
	SaveJudgment(judgment *Location) error

	// ListJudgments returns all judgments, optionally filtered
	ListJudgments(dbID *int, location *string, limit, offset int) ([]*Location, error)

	// GetAllJudgmentsSorted returns all judgments, sorted by db_id and location
	GetAllJudgmentsSorted() ([]*Location, error)

	// BulkInsertJudgments inserts a slice of judgments into the database
	BulkInsertJudgments(judgments []*Location) error

	// CountJudgments returns the total number of judgments
	CountJudgments() (int, error)

	// GetLocationClusters retrieves a list of location clusters.
	GetLocationClusters(dbID *int) ([]*LocationCluster, error)

	// MergeLocations merges a list of locations into a single location.
	MergeLocations(dbID int, targetLocation, canonicalLocation string) error

	// DB returns the underlying database connection
	DB() *sql.DB
}

type sqlJudgmentRepository struct {
	db    *sql.DB
	dbMap map[int]string
}

// NewLocationRepository creates a new judgment repository.
func NewLocationRepository(db *sql.DB, dbMap map[int]string) LocationRepository {
	return &sqlJudgmentRepository{db: db, dbMap: dbMap}
}

// DB returns the underlying database connection for advanced queries.
func (r *sqlJudgmentRepository) DB() *sql.DB {
	return r.db
}

func (r *sqlJudgmentRepository) CreateSchema() error {
	// DuckDB needs to load the spatial extension
	_, err := r.db.Exec(`INSTALL spatial; LOAD spatial;`)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		CREATE SEQUENCE IF NOT EXISTS locations_seq START 1;

		CREATE TABLE IF NOT EXISTS locations (
			id INTEGER PRIMARY KEY DEFAULT nextval('locations_seq'),
			db_id INTEGER NOT NULL,
			location VARCHAR NOT NULL,
			canonical_location VARCHAR,
			point POINT_2D NOT NULL,
			is_electronic BOOLEAN DEFAULT FALSE,
			geocoding_method VARCHAR NOT NULL,
			confidence VARCHAR NOT NULL,
			notes TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			h3_res1 UBIGINT,
			h3_res2 UBIGINT,
			h3_res3 UBIGINT,
			h3_res4 UBIGINT,
			h3_res5 UBIGINT,
			h3_res6 UBIGINT,
			h3_res7 UBIGINT,
			h3_res8 UBIGINT,
			UNIQUE(db_id, location)
		);
	`)

	return err
}

func (r *sqlJudgmentRepository) SaveJudgment(judgment *Location) error {
	if judgment.Point == nil {
		return errors.New("point can't be null")
	}
	// Check if exists
	judgments, err := r.ListJudgments(&judgment.DbID, &judgment.Location, 1, 0)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var existing *Location
	if len(judgments) > 0 {
		existing = judgments[0]
	}

	if err = judgment.computeH3(); err != nil {
		return err
	}

	judgment.UpdatedAt = time.Now()
	if existing != nil {
		// Update
		_, err = r.db.Exec(`
			UPDATE locations
			SET point = ST_Point(?, ?), is_electronic = ?,
			    geocoding_method = ?, confidence = ?, notes = ?,
			    updated_at = ?, canonical_location = ?,
				h3_res1 = ?, h3_res2 = ?, h3_res3 = ?, h3_res4 = ?, h3_res5 = ?, h3_res6 = ?, h3_res7 = ?, h3_res8 = ?
			WHERE db_id = ? AND location = ?
		`,
			judgment.Point.Lng,
			judgment.Point.Lat,
			judgment.IsElectronic,
			judgment.GeocodingMethod,
			judgment.Confidence,
			judgment.Notes,
			judgment.UpdatedAt,
			judgment.CanonicalLocation,
			judgment.H3Res1,
			judgment.H3Res2,
			judgment.H3Res3,
			judgment.H3Res4,
			judgment.H3Res5,
			judgment.H3Res6,
			judgment.H3Res7,
			judgment.H3Res8,
			judgment.DbID,
			judgment.Location,
		)

		return err
	}

	// Insert
	judgment.CreatedAt = judgment.UpdatedAt

	return r.BulkInsertJudgments([]*Location{judgment})
}

func (r *sqlJudgmentRepository) BulkInsertJudgments(judgments []*Location) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO locations(
			db_id,
		    location,
			canonical_location,
			point,
		    is_electronic,
			geocoding_method,
		    confidence,
		    notes,
		    created_at,
		    updated_at,
			h3_res1,
			h3_res2,
			h3_res3,
			h3_res4,
			h3_res5,
			h3_res6,
			h3_res7,
			h3_res8
		)
		VALUES (?, ?, ?, ST_Point(?, ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			err = rErr // Prioritize the rollback error if commit also failed
		}

		return err
	}
	defer stmt.Close()

	for _, j := range judgments {
		cannonical := &j.CanonicalLocation
		if len(*cannonical) == 0 {
			cannonical = nil
		}

		if err = j.computeH3(); err != nil {
			return err
		}

		result, err := stmt.Exec(
			j.DbID,
			j.Location,
			cannonical,
			j.Point.Lng,
			j.Point.Lat,
			j.IsElectronic,
			j.GeocodingMethod,
			j.Confidence,
			j.Notes,
			j.CreatedAt,
			j.UpdatedAt,
			j.H3Res1,
			j.H3Res2,
			j.H3Res3,
			j.H3Res4,
			j.H3Res5,
			j.H3Res6,
			j.H3Res7,
			j.H3Res8,
		)
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				err = rErr // Prioritize the rollback error if commit also failed
			}

			return err
		}

		// Attempt to get the last inserted ID, though it's not used in this function's current logic.
		// This mirrors the change made in SaveJudgment.
		_, err = result.LastInsertId()
		if err != nil {
			// If LastInsertId is not supported by the driver, this error will be returned.
			// Rollback the transaction if an error occurs here.
			if rErr := tx.Rollback(); rErr != nil {
				err = rErr // Prioritize the rollback error if commit also failed
			}

			return fmt.Errorf("failed to get last insert ID during bulk insert: %w", err)
		}
	}
	// Note: The ID is not assigned back to j.ID as it's not used in this function.

	return tx.Commit()
}

func (r *sqlJudgmentRepository) GetJudgment(dbID int, location string) (*Location, error) {
	judgment := &Location{Point: &spatial.Point{}}

	var canonicalLocation sql.NullString

	var h3Res1, h3Res2, h3Res3, h3Res4, h3Res5, h3Res6, h3Res7, h3Res8 sql.NullInt64

	err := r.db.QueryRow(`
		SELECT db_id, location, point, is_electronic,
		       geocoding_method, confidence, notes, created_at, updated_at, canonical_location,
			   h3_res1, h3_res2, h3_res3, h3_res4, h3_res5, h3_res6, h3_res7, h3_res8
		FROM locations
		WHERE db_id = ? AND location = ?
	`, dbID, location).Scan(
		&judgment.DbID,
		&judgment.Location,
		&judgment.Point,
		&judgment.IsElectronic,
		&judgment.GeocodingMethod,
		&judgment.Confidence,
		&judgment.Notes,
		&judgment.CreatedAt,
		&judgment.UpdatedAt,
		&canonicalLocation,
		&h3Res1,
		&h3Res2,
		&h3Res3,
		&h3Res4,
		&h3Res5,
		&h3Res6,
		&h3Res7,
		&h3Res8,
	)
	if err != nil {
		return nil, err
	}

	if canonicalLocation.Valid {
		judgment.CanonicalLocation = canonicalLocation.String
	}

	if h3Res1.Valid {
		judgment.H3Res1 = h3Res1.Int64
	}

	if h3Res2.Valid {
		judgment.H3Res2 = h3Res2.Int64
	}

	if h3Res3.Valid {
		judgment.H3Res3 = h3Res3.Int64
	}

	if h3Res4.Valid {
		judgment.H3Res4 = h3Res4.Int64
	}

	if h3Res5.Valid {
		judgment.H3Res5 = h3Res5.Int64
	}

	if h3Res6.Valid {
		judgment.H3Res6 = h3Res6.Int64
	}

	if h3Res7.Valid {
		judgment.H3Res7 = h3Res7.Int64
	}

	if h3Res8.Valid {
		judgment.H3Res8 = h3Res8.Int64
	}

	return judgment, nil
}

func (r *sqlJudgmentRepository) list(query string, args []any) ([]*Location, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var judgments []*Location

	for rows.Next() {
		judgment := &Location{Point: &spatial.Point{}}

		var canonicalLocation sql.NullString

		var h3Res1, h3Res2, h3Res3, h3Res4, h3Res5, h3Res6, h3Res7, h3Res8 sql.NullInt64

		err := rows.Scan(
			&judgment.DbID, &judgment.Location,
			&judgment.Point, &judgment.IsElectronic,
			&judgment.GeocodingMethod, &judgment.Confidence, &judgment.Notes,
			&judgment.CreatedAt, &judgment.UpdatedAt, &canonicalLocation,
			&h3Res1, &h3Res2, &h3Res3, &h3Res4, &h3Res5, &h3Res6, &h3Res7, &h3Res8,
		)
		if err != nil {
			return nil, err
		}

		if canonicalLocation.Valid {
			judgment.CanonicalLocation = canonicalLocation.String
		}

		if h3Res1.Valid {
			judgment.H3Res1 = h3Res1.Int64
		}

		if h3Res2.Valid {
			judgment.H3Res2 = h3Res2.Int64
		}

		if h3Res3.Valid {
			judgment.H3Res3 = h3Res3.Int64
		}

		if h3Res4.Valid {
			judgment.H3Res4 = h3Res4.Int64
		}

		if h3Res5.Valid {
			judgment.H3Res5 = h3Res5.Int64
		}

		if h3Res6.Valid {
			judgment.H3Res6 = h3Res6.Int64
		}

		if h3Res7.Valid {
			judgment.H3Res7 = h3Res7.Int64
		}

		if h3Res8.Valid {
			judgment.H3Res8 = h3Res8.Int64
		}

		judgments = append(judgments, judgment)
	}

	return judgments, nil
}

var baseSelect = `
	SELECT db_id, location, point, is_electronic,
	       geocoding_method, confidence, notes,
		   created_at, updated_at, canonical_location,
		   h3_res1, h3_res2, h3_res3, h3_res4, h3_res5, h3_res6, h3_res7, h3_res8
	FROM locations
`

func (r *sqlJudgmentRepository) ListJudgments(dbID *int, location *string, limit, offset int) ([]*Location, error) {
	query := baseSelect

	args := []any{}

	if dbID != nil {
		query += " WHERE db_id = ?"

		args = append(args, *dbID)

		if nil != location {
			query += " AND location = ?"

			args = append(args, *location)
		}
	}

	query += " ORDER BY updated_at DESC"

	if limit > 0 {
		query += " LIMIT ? OFFSET ?"

		args = append(args, limit, offset)
	}

	return r.list(query, args)
}

func (r *sqlJudgmentRepository) CountJudgments() (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM locations",
	).Scan(&count)

	return count, err
}

func (r *sqlJudgmentRepository) GetAllJudgmentsSorted() ([]*Location, error) {
	return r.list(baseSelect+` ORDER BY db_id, location`,
		[]any{},
	)
}

func (r *sqlJudgmentRepository) GetLocationClusters(dbID *int) ([]*LocationCluster, error) {
	judgments, err := r.ListJudgments(dbID, nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("listing judgments: %w", err)
	}

	log.Printf("Fetched %d judgments for clustering", len(judgments))

	offenseCounts, err := r.getOffenseCounts()
	if err != nil {
		return nil, fmt.Errorf("getting offense counts: %w", err)
	}

	judgmentClusters := clusterJudgments(judgments, 10) // 10 meters
	log.Printf("Created %d raw clusters", len(judgmentClusters))

	result := make([]*LocationCluster, 0, len(judgmentClusters))

	for _, jc := range judgmentClusters {
		if len(jc) < 2 {
			continue // Skip clusters with only one location
		}

		var principal *ClusterLocation

		var totalOffenses int

		var clusterDbID int

		// Filter out subordinate locations that have a canonical master and the same coordinates as the principal.
		var filteredJudgments []*Location

		for _, j := range jc {
			exclude := false

			if j.CanonicalLocation != "" {
				// Check if coordinates match the principal's coordinates.
				// Note: principal is determined later, so we need to find it first or compare against the first judgment's point if it's canonical.
				// A more robust approach is to find the principal first, then filter.
				// Let's find the principal first based on offense count.
				// This requires a slight reordering or a two-pass approach.
				// For now, let's assume we can find the principal's coordinates.
				// We will determine the principal *after* filtering, or ensure the principal itself isn't filtered out if it meets criteria.
				// The requirement is to ignore locations that *already have* a canonical location AND same coordinates as the master.
				// The 'principal' is defined as the one with most offenses.
				// If the principal itself meets the criteria, it should be filtered out, and a new principal identified.
				// Let's refine: we need to identify the principal first, then filter based on its coordinates.
				// However, the principal is determined *after* creating the 'locations' slice.
				// This suggests we should filter 'jc' first, then determine the principal from the filtered list.
				// Let's find the principal from the original 'jc' to use its coordinates for filtering.
				// This requires finding the judgment with the max offense count from 'jc'.
				var principalJudgment *Location

				maxOffenses := -1

				for _, pj := range jc {
					count := offenseCounts[fmt.Sprintf("%d-%s", pj.DbID, pj.Location)]
					if count > maxOffenses {
						maxOffenses = count
						principalJudgment = pj
					}
				}

				if principalJudgment != nil && j.CanonicalLocation != "" &&
					principalJudgment.Point != nil && j.Point != nil &&
					principalJudgment.Point.Lat == j.Point.Lat && principalJudgment.Point.Lng == j.Point.Lng {
					exclude = true
				}
			}

			if !exclude {
				filteredJudgments = append(filteredJudgments, j)
			}
		}

		// If after filtering, the cluster becomes too small, skip it.
		if len(filteredJudgments) < 2 {
			continue // Skip this cluster if it has less than 2 locations after filtering
		}

		locations := make([]*ClusterLocation, len(filteredJudgments))

		for k, j := range filteredJudgments {
			count := offenseCounts[fmt.Sprintf("%d-%s", j.DbID, j.Location)]
			locations[k] = &ClusterLocation{
				DbID:         j.DbID,
				Description:  j.Location,
				Point:        *j.Point,
				OffenseCount: count,
			}
			totalOffenses += count
			clusterDbID = j.DbID // Assuming all locations in a cluster share the same db_id
		}

		// Find principal from the filtered locations
		sort.Slice(locations, func(i, j int) bool {
			return locations[i].OffenseCount > locations[j].OffenseCount
		})

		principal = locations[0]
		principal.IsPrincipal = true

		// Calculate distances from principal
		for _, loc := range locations {
			loc.DistanceFromPrincipal = principal.Point.HaversineDistance(&loc.Point)
		}

		dbName := r.dbMap[clusterDbID]

		result = append(result, &LocationCluster{
			DbID:          clusterDbID,
			Location:      principal.Description,
			DbName:        dbName,
			TotalOffenses: totalOffenses,
			Locations:     locations,
		})
	}

	// Sort clusters by total offenses
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalOffenses > result[j].TotalOffenses
	})

	log.Printf("Returning %d clusters after filtering", len(result))

	return result, nil
}

func (r *sqlJudgmentRepository) getOffenseCounts() (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT db_id, location, COUNT(*)
		FROM offenses WHERE location IS NOT NULL
		GROUP BY db_id, location
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)

	for rows.Next() {
		var dbID int

		var location string

		var count int
		if err := rows.Scan(&dbID, &location, &count); err != nil {
			return nil, err
		}

		counts[fmt.Sprintf("%d-%s", dbID, location)] = count
	}

	return counts, nil
}

func (r *sqlJudgmentRepository) MergeLocations(dbID int, targetLocation, canonicalLocation string) error {
	// Get the canonical judgment to retrieve the point
	canonicalJudgments, err := r.ListJudgments(&dbID, &canonicalLocation, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to list canonical judgment for dbID %d, location %s: %w", dbID, canonicalLocation, err)
	}

	if len(canonicalJudgments) == 0 {
		return fmt.Errorf("canonical judgment not found for dbID %d, location %s", dbID, canonicalLocation)
	}

	canonicalJudgment := canonicalJudgments[0]

	// Get the target judgment
	targetJudgments, err := r.ListJudgments(&dbID, &targetLocation, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to list target judgment for dbID %d, location %s: %w", dbID, targetLocation, err)
	}

	if len(targetJudgments) == 0 {
		return fmt.Errorf("target judgment not found for dbID %d, location %s", dbID, targetLocation)
	}

	targetJudgment := targetJudgments[0]

	// Set the canonical location
	targetJudgment.CanonicalLocation = canonicalLocation

	// Update the target's point to match the canonical one
	if canonicalJudgment.Point != nil {
		targetJudgment.Point = canonicalJudgment.Point
		targetJudgment.H3Res1 = canonicalJudgment.H3Res1
		targetJudgment.H3Res2 = canonicalJudgment.H3Res2
		targetJudgment.H3Res3 = canonicalJudgment.H3Res3
		targetJudgment.H3Res4 = canonicalJudgment.H3Res4
		targetJudgment.H3Res5 = canonicalJudgment.H3Res5
		targetJudgment.H3Res6 = canonicalJudgment.H3Res6
		targetJudgment.H3Res7 = canonicalJudgment.H3Res7
		targetJudgment.H3Res8 = canonicalJudgment.H3Res8
	}

	// Save the updated target judgment
	return r.SaveJudgment(targetJudgment)
}
