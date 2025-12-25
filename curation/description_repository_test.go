// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"database/sql"
	"sort"
	"testing"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDescriptionDB(t *testing.T) (*sql.DB, DescriptionRepository) {
	db, err := sql.Open("duckdb", "") // In-memory database
	require.NoError(t, err)

	repo := NewDescriptionRepository(db)
	err = repo.CreateSchema()
	require.NoError(t, err)

	// Need offenses table for GetUnclassifiedDescriptions
	_, err = db.Exec(`
		CREATE TABLE offenses (
			id VARCHAR,
			db_id INTEGER,
			time VARCHAR,
			date VARCHAR,
			description VARCHAR,
			location VARCHAR,
			doc_source VARCHAR,
			article_ids VARCHAR[],
			article_codes TINYINT[]
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE locations (
			id INTEGER PRIMARY KEY,
			db_id INTEGER,
			location VARCHAR,
			canonical_location VARCHAR,
			point DOUBLE[],
			is_electronic BOOLEAN DEFAULT false,
			geocoding_method VARCHAR,
			confidence VARCHAR,
			notes VARCHAR,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	// Seed articles for tests
	articles := []Article{
		{ID: "G.1", Text: "Art 1", Code: 1, Title: "Title 1"},
		{ID: "G.2", Text: "Art 2", Code: 2, Title: "Title 2"},
		{ID: "G.3", Text: "Art 3", Code: 3, Title: "Title 3"},
		{ID: "G.4", Text: "Art 4", Code: 1, Title: "Title 1"}, // Article with no description
	}
	err = repo.SeedArticles(articles)
	require.NoError(t, err)

	return db, repo
}

func TestDescriptionSchema(t *testing.T) {
	db, _ := setupDescriptionDB(t)
	defer db.Close()

	// Check if tables exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 4, count) // Changed from 3 to 4

	err = db.QueryRow("SELECT COUNT(*) FROM descriptions").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveDescriptionClassification(t *testing.T) {
	_, repo := setupDescriptionDB(t)

	description := "TEST DESCRIPTION"
	articleIDs := []string{"G.1", "G.2"}
	expectedCodes := []int8{1, 2}

	err := repo.SaveDescriptionClassification(description, articleIDs)
	require.NoError(t, err)

	// Verify using the repo method
	saved, err := repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.NotNil(t, saved)
	assert.ElementsMatch(t, articleIDs, saved.ArticleIDs)
	assert.ElementsMatch(t, expectedCodes, saved.ArticleCodes)

	// Test update
	newArticleIDs := []string{"G.3"}
	newExpectedCodes := []int8{3}
	err = repo.SaveDescriptionClassification(description, newArticleIDs)
	require.NoError(t, err)

	// Verify update using the repo method
	saved, err = repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.NotNil(t, saved)
	assert.ElementsMatch(t, newArticleIDs, saved.ArticleIDs)
	assert.ElementsMatch(t, newExpectedCodes, saved.ArticleCodes)
}

func TestGetUnclassifiedDescriptions(t *testing.T) {
	db, repo := setupDescriptionDB(t)
	defer db.Close()

	// Insert some offenses
	_, err := db.Exec(`
		INSERT INTO offenses (description) VALUES
			('UNCLASSIFIED 1'),
			('UNCLASSIFIED 2'),
			('CLASSIFIED 1');
	`)
	require.NoError(t, err)

	// Classify one of them
	err = repo.SaveDescriptionClassification("CLASSIFIED 1", []string{"G.1"})
	require.NoError(t, err)

	unclassified, err := repo.GetUnclassifiedDescriptions(10)
	require.NoError(t, err)

	assert.Len(t, unclassified, 2)
	assert.Contains(t, unclassified, DescriptionQueueItem{Description: "UNCLASSIFIED 1", Count: 1})
	assert.Contains(t, unclassified, DescriptionQueueItem{Description: "UNCLASSIFIED 2", Count: 1})
	assert.NotContains(t, unclassified, DescriptionQueueItem{Description: "CLASSIFIED 1", Count: 0}) // Count doesn't matter for classified
}

func TestAreMultiArticlePartsClassified(t *testing.T) {
	db, repo := setupDescriptionDB(t)
	defer db.Close()

	tests := []struct {
		name          string
		description   string
		classifyParts []string
		expected      bool
	}{
		{
			name:          "Single-article not classified",
			description:   "PART 1",
			classifyParts: []string{},
			expected:      false,
		},
		{
			name:          "Single-article classified",
			description:   "PART 1",
			classifyParts: []string{"PART 1"},
			expected:      true,
		},
		{
			name:          "Multi-article with one part unclassified",
			description:   "PART 1, PART 2",
			classifyParts: []string{"PART 1"},
			expected:      false,
		},
		{
			name:          "Multi-article with all parts classified",
			description:   "PART 1, PART 2, PART 3",
			classifyParts: []string{"PART 1", "PART 2", "PART 3"},
			expected:      true,
		},
		{
			name:          "Multi-article with no parts classified",
			description:   "PART 1, PART 2",
			classifyParts: []string{},
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Classify the parts
			for _, part := range tt.classifyParts {
				err := repo.SaveDescriptionClassification(part, []string{"G.1"})
				require.NoError(t, err)
			}

			// Check if all parts are classified
			allClassified, err := repo.AreMultiArticlePartsClassified(tt.description)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, allClassified)

			// Clean up for next test
			_, err = db.Exec("DELETE FROM descriptions")
			require.NoError(t, err)
		})
	}
}

func TestGetDescriptionWithArticles(t *testing.T) {
	db, repo := setupDescriptionDB(t)
	defer db.Close()

	description := "TEST DESCRIPTION"
	articleIDs := []string{"G.1", "G.2"}
	expectedCodes := []int8{1, 2}

	// Initially, should not be found
	result, err := repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.Nil(t, result)

	// Save a classification
	err = repo.SaveDescriptionClassification(description, articleIDs)
	require.NoError(t, err)

	// Now should be found
	result, err = repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, description, result.Description)
	assert.ElementsMatch(t, articleIDs, result.ArticleIDs)
	assert.ElementsMatch(t, expectedCodes, result.ArticleCodes)
}

func TestGetReviewAssignments(t *testing.T) {
	db, repo := setupDescriptionDB(t)
	defer db.Close()

	// 1. Seed Data
	_, err := db.Exec(`
		INSERT INTO offenses (description) VALUES
			('SINGLE ASSIGNMENT DESC'),
			('SINGLE ASSIGNMENT DESC'),
			('MULTI ASSIGNMENT DESC');
	`)
	require.NoError(t, err)

	err = repo.SaveDescriptionClassification("SINGLE ASSIGNMENT DESC", []string{"G.1"})
	require.NoError(t, err)
	err = repo.SaveDescriptionClassification("MULTI ASSIGNMENT DESC", []string{"G.2", "G.3"})
	require.NoError(t, err)
	err = repo.SaveDescriptionClassification("NO OFFENSES DESC", []string{"G.2"})
	require.NoError(t, err)

	// 2. Execute
	reviewAssignments, err := repo.GetReviewAssignments()
	require.NoError(t, err)

	// 3. Assert
	require.Len(t, reviewAssignments, 3)

	// Sort reviewAssignments by code for consistent assertion order
	sort.Slice(reviewAssignments, func(i, j int) bool {
		return reviewAssignments[i].Code < reviewAssignments[j].Code
	})

	// --- Assert Code 1 ---
	code1 := reviewAssignments[0]
	assert.Equal(t, 1, code1.Code)
	assert.Equal(t, "I", code1.Roman)
	require.Len(t, code1.Articles, 2)

	// Sort articles under Code 1
	articlesCode1 := code1.Articles
	sort.Slice(articlesCode1, func(i, j int) bool {
		return articlesCode1[i].ID < articlesCode1[j].ID
	})

	// Check Article G.1
	articleG1 := articlesCode1[0]
	assert.Equal(t, "G.1", articleG1.ID)
	assert.Equal(t, "Art 1", articleG1.Text)
	require.Len(t, articleG1.Descriptions, 1)
	descG1 := articleG1.Descriptions[0]
	assert.Equal(t, "SINGLE ASSIGNMENT DESC", descG1.Description)
	assert.Equal(t, 2, descG1.OffenseCount)

	// Check Article G.4 (empty article)
	articleG4 := articlesCode1[1]
	assert.Equal(t, "G.4", articleG4.ID)
	assert.Equal(t, "Art 4", articleG4.Text)
	require.Empty(t, articleG4.Descriptions) // Expecting 0 descriptions

	// --- Assert Code 2 ---
	code2 := reviewAssignments[1]
	assert.Equal(t, 2, code2.Code)
	assert.Equal(t, "II", code2.Roman)
	require.Len(t, code2.Articles, 1)
	articleG2 := code2.Articles[0]
	assert.Equal(t, "G.2", articleG2.ID)
	assert.Equal(t, "Art 2", articleG2.Text)
	require.Len(t, articleG2.Descriptions, 1)
	descG2 := articleG2.Descriptions[0]
	assert.Equal(t, "NO OFFENSES DESC", descG2.Description)
	assert.Equal(t, 0, descG2.OffenseCount) // No offenses were inserted for "NO OFFENSES DESC"

	// --- Assert Code 3 ---
	code3 := reviewAssignments[2]
	assert.Equal(t, 3, code3.Code)
	assert.Equal(t, "III", code3.Roman)
	require.Len(t, code3.Articles, 1)
	articleG3 := code3.Articles[0]
	assert.Equal(t, "G.3", articleG3.ID)
	assert.Equal(t, "Art 3", articleG3.Text)
	require.Empty(t, articleG3.Descriptions) // "MULTI ASSIGNMENT DESC" is filtered out
}

func TestDescriptionUpdatedAt(t *testing.T) {
	_, repo := setupDescriptionDB(t)

	description := "TEST UPDATED AT"
	articleIDs := []string{"G.1"}

	// 1. Create check
	start := time.Now().Truncate(time.Second)
	err := repo.SaveDescriptionClassification(description, articleIDs)
	require.NoError(t, err)

	saved, err := repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.NotNil(t, saved)
	assert.False(t, saved.UpdatedAt.Before(start))

	// 2. Update check
	time.Sleep(1 * time.Second) // Ensure time advances
	updateStart := time.Now().Truncate(time.Second)

	err = repo.SaveDescriptionClassification(description, []string{"G.2"})
	require.NoError(t, err)

	updated, err := repo.GetDescriptionWithArticles(description)
	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.False(t, updated.UpdatedAt.Before(updateStart))
	assert.True(t, updated.UpdatedAt.After(saved.UpdatedAt))
}
