// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("duckdb", "")
	require.NoError(t, err)

	repo, _ := NewSQLOffenseRepository(db)
	err = repo.CreateSchema()
	require.NoError(t, err)

	return db
}

func TestSQLRepository_SaveTrafficOffenses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, _ := NewSQLOffenseRepository(db)

	now := time.Now().UTC()
	offenses := []*TrafficOffense{
		{
			DbID: 45,
			Document: &Document{
				DocSource: "doc1",
				DocID:     "doc1_id",
				DocDate:   now,
			},
			RecordID: 1,
			ID:       "offense1",
			Vehicle:  "AAAA123",
			VehicleInfo: &VehicleInfo{
				Country:     "UY",
				VehicleType: "AUTO",
			},
			Time:            now,
			Location:        "Some Location",
			DisplayLocation: "Some Location",
			Description:     "Speeding",
			UR:              100,
		},
		{
			DbID: 45,
			Document: &Document{
				DocSource: "doc1",
			},
			RecordID: 2,
			Error:    "Some error",
		},
	}

	err := repo.SaveTrafficOffenses(offenses)
	require.NoError(t, err)

	// Verify using raw SQL
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM offenses").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	var vehicle string
	err = db.QueryRow("SELECT vehicle FROM offenses WHERE record_id = 1").Scan(&vehicle)
	require.NoError(t, err)
	assert.Equal(t, "AAAA123", vehicle)

	var errStr string
	err = db.QueryRow("SELECT error FROM offenses WHERE record_id = 2").Scan(&errStr)
	require.NoError(t, err)
	assert.Equal(t, "Some error", errStr)
}

func TestSQLRepository_GetExtractedDocuments(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, _ := NewSQLOffenseRepository(db)

	// Insert some data
	_, err := db.Exec(`
		INSERT INTO offenses (db_id, doc_source, record_id) VALUES
			(45, 'doc1', 1),
			(45, 'doc2', 2),
			(46, 'doc3', 3)
	`)
	require.NoError(t, err)

	// Test for db 45
	docs, err := repo.GetExtractedDocuments(&DbReference{ID: 45})
	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.True(t, docs["doc1"])
	assert.True(t, docs["doc2"])
	assert.False(t, docs["doc3"])

	// Test for db 46
	docs, err = repo.GetExtractedDocuments(&DbReference{ID: 46})
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.True(t, docs["doc3"])
}

func TestSQLRepository_SaveTrafficOffenses_H3Nulls(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, _ := NewSQLOffenseRepository(db)

	now := time.Now().UTC()
	offense := &TrafficOffense{
		DbID: 45,
		Document: &Document{
			DocSource: "doc_h3",
			DocID:     "doc_h3_id",
			DocDate:   now,
		},
		RecordID: 1,
		ID:       "offense_h3",
		Vehicle:  "H3TEST",
		Time:     now,
		// H3 fields are 0 by default
	}

	err := repo.SaveTrafficOffenses([]*TrafficOffense{offense})
	require.NoError(t, err)

	var h3Res1 sql.NullInt64
	err = db.QueryRow("SELECT h3_res1 FROM offenses WHERE record_id = 1").Scan(&h3Res1)
	require.NoError(t, err)

	assert.False(t, h3Res1.Valid, "h3_res1 should be NULL")
}
