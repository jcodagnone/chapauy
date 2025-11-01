// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/jcodagnone/chapauy/curation/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLocationRepository is a mock implementation of LocationRepository for testing.
type MockLocationRepository struct{}

func (m *MockLocationRepository) CreateSchema() error            { return nil }
func (m *MockLocationRepository) SaveJudgment(_ *Location) error { return nil }
func (m *MockLocationRepository) GetJudgment(_ int, _ string) (*Location, error) {
	return nil, sql.ErrNoRows
}

func (m *MockLocationRepository) ListJudgments(_ *int, _ *string, _, _ int) ([]*Location, error) {
	return nil, nil
}
func (m *MockLocationRepository) CountJudgments() (int, error) { return 0, nil }
func (m *MockLocationRepository) MergeLocations(_ int, _, _ string) error {
	return nil
}
func (m *MockLocationRepository) GetLocationClusters(_ *int) ([]*LocationCluster, error) {
	return nil, nil
}
func (m *MockLocationRepository) BulkInsertJudgments(_ []*Location) error     { return nil }
func (m *MockLocationRepository) DB() *sql.DB                                 { return nil }
func (m *MockLocationRepository) GetAllJudgmentsSorted() ([]*Location, error) { return nil, nil } // Added missing method // Added missing method // Added missing method

// setupServerTest initializes a Gin router and a curation.Server for testing.
func setupServerTest(t *testing.T) (*gin.Engine, *Server, *sql.DB, DescriptionRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	db, descriptionRepo := setupDescriptionDB(t)

	// Use mock repositories for other dependencies
	geocodeRepo := &MockLocationRepository{}
	radarIndex := &RadarIndex{radars: make(map[string]*Radar)} // Initialize empty RadarIndex

	server := NewServer(geocodeRepo, db, radarIndex, map[int]string{}) // Pass db directly

	// Register API routes
	// Note: listDatabases is removed
	router.GET("/api/locations/queue", server.getLocationQueue)
	router.GET("/api/descriptions/unclassified", server.getUnclassifiedDescriptions)
	router.GET("/api/descriptions/articles", server.listArticles)
	router.POST("/api/descriptions/classify", server.classifyDescription)
	router.GET("/api/descriptions/progress", server.getDescriptionProgress)
	router.POST("/api/descriptions/articles/add", server.addArticle)
	router.GET("/api/descriptions/articles/search", server.searchArticles)
	router.GET("/api/descriptions/suggest", server.suggestClassification)

	return router, server, db, descriptionRepo
}

func TestSuggestClassificationAPI(t *testing.T) {
	router, _, db, repo := setupServerTest(t)
	defer db.Close()

	// Seed articles
	err := repo.AddArticle("18.9.2", "Estacionar en lugar tarifado sin abonar la tarifa correspondiente.", 18, "Estacionamiento")
	require.NoError(t, err)
	err = repo.AddArticle("4.11", "Circular sin haber realizado la inspección técnica vehicular departamental reglamentaria.", 4, "Circulación")
	require.NoError(t, err)
	err = repo.AddArticle("21.3.1", "Conductor o acompañante sin casco protector.", 21, "Seguridad")
	require.NoError(t, err)

	// Test with composite description
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/descriptions/suggest?description=ESTACIONADO%20SIN%20ABONAR%20TARIFA,%20CONDUCTOR%20SIN%20CASCO", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var suggestions []Suggestion
	err = json.Unmarshal(w.Body.Bytes(), &suggestions)
	require.NoError(t, err)
	assert.Len(t, suggestions, 2)

	foundIDs := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		foundIDs = append(foundIDs, s.ArticleID)
	}

	assert.Contains(t, foundIDs, "18.9.2")
	assert.Contains(t, foundIDs, "21.3.1")
}

func TestGetUnclassifiedDescriptionsAPI(t *testing.T) {
	router, _, db, repo := setupServerTest(t)
	defer db.Close()

	// Seed some offenses
	_, err := db.Exec(`
		INSERT INTO offenses (db_id, description) VALUES
			(1, 'UNCLASSIFIED 1'),
			(1, 'UNCLASSIFIED 1'),
			(1, 'UNCLASSIFIED 2'),
			(2, 'UNCLASSIFIED 3'),
			(2, 'CLASSIFIED 1');
	`)
	require.NoError(t, err)

	// Classify one description
	err = repo.AddArticle("A1", "Article 1", 1, "Test")
	require.NoError(t, err)
	err = repo.SaveDescriptionClassification("CLASSIFIED 1", []string{"A1"})
	require.NoError(t, err)

	// Test without db_id filter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/descriptions/unclassified", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var descriptions []DescriptionQueueItem
	err = json.Unmarshal(w.Body.Bytes(), &descriptions)
	require.NoError(t, err)
	assert.Len(t, descriptions, 3)

	// Check content - no db_id or db_name expected
	assert.Contains(t, descriptions, DescriptionQueueItem{Description: "UNCLASSIFIED 1", Count: 2})
	assert.Contains(t, descriptions, DescriptionQueueItem{Description: "UNCLASSIFIED 2", Count: 1})
	assert.Contains(t, descriptions, DescriptionQueueItem{Description: "UNCLASSIFIED 3", Count: 1})
}

func TestGetDescriptionProgressAPI(t *testing.T) {
	router, _, db, repo := setupServerTest(t)
	defer db.Close()

	// Seed some offenses
	_, err := db.Exec(`
		INSERT INTO offenses (db_id, description) VALUES
			(1, 'DESC A'),
			(1, 'DESC A'),
			(1, 'DESC B'),
			(2, 'DESC C'),
			(2, 'DESC D');
	`)
	require.NoError(t, err)

	// Classify some descriptions
	err = repo.AddArticle("ART1", "Article 1", 1, "Test")
	require.NoError(t, err)
	err = repo.SaveDescriptionClassification("DESC A", []string{"ART1"})
	require.NoError(t, err)
	err = repo.AddArticle("ART2", "Article 2", 2, "Test")
	require.NoError(t, err)
	err = repo.SaveDescriptionClassification("DESC C", []string{"ART2"})
	require.NoError(t, err)

	// Test without db_id filter
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/descriptions/progress", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var progress DescriptionProgressResponse
	err = json.Unmarshal(w.Body.Bytes(), &progress)
	require.NoError(t, err)
	assert.Equal(t, 4, progress.TotalDescriptions)      // A, B, C, D
	assert.Equal(t, 2, progress.ClassifiedDescriptions) // A, C
	assert.InDelta(t, 50.0, progress.DescriptionsPercentage, 0.01)
}

func TestAddArticleAPI(t *testing.T) {
	router, _, db, _ := setupServerTest(t)
	defer db.Close()

	// Test adding a new article
	newArticle := Article{ID: "NEW1", Text: "New Article Description", Code: 99, Title: "Testing"}
	body, _ := json.Marshal(newArticle)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/descriptions/articles/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool

	var err error // Declare err here
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])
	t.Logf("API Response: %s", w.Body.String())
	t.Logf("New Article ID: %s, Description: %s", newArticle.ID, newArticle.Text)

	// Verify it's in the DB
	var text string
	err = db.QueryRow("SELECT text FROM articles WHERE id = ?", "NEW1").Scan(&text)
	t.Logf("DB Query Error: %v", err)
	t.Logf("Scanned Text: %s", text)
	require.NoError(t, err)
	assert.Equal(t, "New Article Description", text)

	// Test adding an existing article (should update)
	updatedArticle := Article{ID: "NEW1", Text: "Updated Description", Code: 99, Title: "Testing Updated"}
	body, _ = json.Marshal(updatedArticle)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/descriptions/articles/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])
	t.Logf("API Response (Update): %s", w.Body.String())
	t.Logf("Updated Article ID: %s, Description: %s", updatedArticle.ID, updatedArticle.Text)

	err = db.QueryRow("SELECT text FROM articles WHERE id = ?", "NEW1").Scan(&text)
	t.Logf("DB Query Error (Update): %v", err)
	t.Logf("Scanned Text (Update): %s", text)
	require.NoError(t, err)
	assert.Equal(t, "Updated Description", text)
}

func TestSearchArticlesAPI(t *testing.T) {
	router, _, db, repo := setupServerTest(t)
	defer db.Close()

	// Seed some articles
	err := repo.AddArticle("ART1", "Article one about traffic", 1, "Traffic")
	require.NoError(t, err)
	err = repo.AddArticle("ART2", "Another article", 2, "General")
	require.NoError(t, err)
	err = repo.AddArticle("RULE3", "Rule about speed limits", 3, "Speed")
	require.NoError(t, err)

	// Test search by ID
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/descriptions/articles/search?query=ART", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var articles []Article
	err = json.Unmarshal(w.Body.Bytes(), &articles)
	require.NoError(t, err)
	assert.Len(t, articles, 2)
	assert.Contains(t, articles, Article{ID: "ART1", Text: "Article one about traffic", Code: 1, Title: "Traffic"})
	assert.Contains(t, articles, Article{ID: "ART2", Text: "Another article", Code: 2, Title: "General"})

	// Test search by text
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/descriptions/articles/search?query=traffic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	articles = []Article{} // Reset
	err = json.Unmarshal(w.Body.Bytes(), &articles)
	require.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Contains(t, articles, Article{ID: "ART1", Text: "Article one about traffic", Code: 1, Title: "Traffic"})

	// Test search with no results
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/descriptions/articles/search?query=nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	articles = []Article{} // Reset
	err = json.Unmarshal(w.Body.Bytes(), &articles)
	require.NoError(t, err)
	assert.Empty(t, articles)

	// Test empty query
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/descriptions/articles/search", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestClassifyDescriptionAPI(t *testing.T) {
	router, _, db, repo := setupServerTest(t)
	defer db.Close()

	// Seed an unclassified offense
	_, err := db.Exec(`INSERT INTO offenses (db_id, description) VALUES (1, 'DESC TO CLASSIFY');`)
	require.NoError(t, err)

	// Add some articles
	err = repo.AddArticle("ART1", "Article one", 1, "Test")
	require.NoError(t, err)
	err = repo.AddArticle("ART2", "Article two", 1, "Test")
	require.NoError(t, err)

	// Classify the description
	classifyReq := ClassifyRequest{
		Description: "DESC TO CLASSIFY",
		ArticleIDs:  []string{"ART1", "ART2"},
	}
	body, _ := json.Marshal(classifyReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/descriptions/classify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"])

	// Verify classification in DB
	var scannedArticleIDs any
	err = db.QueryRow("SELECT article_ids FROM descriptions WHERE description = ?", "DESC TO CLASSIFY").Scan(&scannedArticleIDs)
	require.NoError(t, err)

	savedArticleIDs, ok := utils.AnyToStringSlice(scannedArticleIDs)
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{"ART1", "ART2"}, savedArticleIDs)

	// Verify it's no longer in unclassified queue
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/descriptions/unclassified", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var descriptions []DescriptionQueueItem
	err = json.Unmarshal(w.Body.Bytes(), &descriptions)
	require.NoError(t, err)
	assert.Empty(t, descriptions) // Should be empty
}
