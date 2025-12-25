// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"context"
	"database/sql" // Added import
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	apikeys "cloud.google.com/go/apikeys/apiv2"
	"cloud.google.com/go/apikeys/apiv2/apikeyspb"
	"github.com/gin-gonic/gin"
	"github.com/jcodagnone/chapauy/spatial"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
)

type Server struct {
	geocodeRepo     LocationRepository
	descriptionRepo DescriptionRepository
	radarIndex      *RadarIndex
	geocoder        Geocoder
	dbMap           map[int]string
}

func NewServer(geocodeRepo LocationRepository, db *sql.DB, radarIndex *RadarIndex, dbMap map[int]string) *Server {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Println("GOOGLE_MAPS_API_KEY is not set. Attempting to retrieve via ADC...")

		var err error

		apiKey, err = getAPIKeyFromADC(context.Background())
		if err != nil {
			log.Printf("Failed to retrieve API key via ADC: %v", err)
			log.Print("GOOGLE_MAPS_API_KEY is not set and ADC failed. Google Maps Geocoding is required.")
		} else {
			log.Println("✅ Successfully retrieved Google Maps API Key via ADC")
		}
	}

	fmt.Println("📍 Geocoding: Google Maps (primary)")

	return &Server{
		geocodeRepo:     geocodeRepo,
		descriptionRepo: NewDescriptionRepository(db), // Create descriptionRepo here
		radarIndex:      radarIndex,
		geocoder:        NewGoogleMapsGeocoder(apiKey),
		dbMap:           dbMap,
	}
}

func getAPIKeyFromADC(ctx context.Context) (string, error) {
	// 1. Get Project ID from ADC
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("finding default credentials: %w", err)
	}

	projectID := creds.ProjectID
	if projectID == "" {
		// Fallback to known Project ID if not found in credentials
		// This happens when using user credentials without a quota project
		projectID = "chapauy-20251216"
		log.Printf("⚠️ No Project ID found in credentials. Using fallback: %s", projectID)
	}

	// 2. Create API Keys client
	client, err := apikeys.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating apikeys client: %w", err)
	}
	defer client.Close()

	// 3. List keys to find the one with the expected display name
	// This matches the DisplayName used in .dagger/gcp/resources.go (MapsDesiredState)
	const targetDisplayName = "ChapaUY Geocoding Key"

	req := &apikeyspb.ListKeysRequest{
		Parent: fmt.Sprintf("projects/%s/locations/global", projectID),
	}

	it := client.ListKeys(ctx, req)

	for {
		key, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return "", fmt.Errorf("listing keys: %w", err)
		}

		if key.DisplayName == targetDisplayName {
			// Found it!
			// ListKeys and GetKey redact the KeyString.
			// We must use GetKeyString method to retrieve the secret.
			log.Printf("Found key resource '%s', retrieving secret...", key.Name)

			getReq := &apikeyspb.GetKeyStringRequest{
				Name: key.Name,
			}

			resp, err := client.GetKeyString(ctx, getReq)
			if err != nil {
				return "", fmt.Errorf("getting key string: %w", err)
			}

			if resp.KeyString == "" {
				return "", fmt.Errorf("key '%s' found but KeyString is still empty after GetKeyString", targetDisplayName)
			}

			return resp.KeyString, nil
		}
	}

	return "", fmt.Errorf("key with display name '%s' not found in project %s", targetDisplayName, projectID)
}

func (s *Server) Run() error {
	r := gin.Default()
	r.SetHTMLTemplate(template.Must(template.New("").ParseGlob("templates/*.html")))
	r.Static("/static", "templates/static")

	r.GET("/", s.geocodeView)
	r.GET("/descriptions", s.descriptionsView)
	r.GET("/review", s.reviewView)
	r.GET("/api/databases", s.listDatabases)
	r.GET("/api/locations/queue", s.getLocationQueue)
	r.POST("/api/locations/merge", s.mergeLocations)
	r.GET("/api/locations/suggest/:db_id/*location", s.suggestCoordinates)
	r.POST("/api/locations/accept/:db_id/*location", s.acceptJudgment)
	r.GET("/api/locations/progress", s.getProgress)
	r.GET("/api/locations/judgments", s.listJudgments)
	r.GET("/api/descriptions/unclassified", s.getUnclassifiedDescriptions)
	r.GET("/api/descriptions/articles", s.listArticles)
	r.POST("/api/descriptions/classify", s.classifyDescription)
	r.GET("/api/descriptions/progress", s.getDescriptionProgress) // New endpoint
	r.POST("/api/descriptions/articles/add", s.addArticle)        // New endpoint
	r.GET("/api/descriptions/articles/search", s.searchArticles)  // New endpoint
	r.GET("/api/descriptions/suggest", s.suggestClassification)

	return r.Run("localhost:8080")
}

func (s *Server) suggestClassification(ctx *gin.Context) {
	description := ctx.Query("description")
	if description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "description query parameter is required"})

		return
	}

	articles, err := s.descriptionRepo.ListArticles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list articles"})

		return
	}

	autoJudger := NewDescriptionClassifier(articles)
	// I'll use a fixed threshold for the UI for now. 0.5 seems reasonable from previous results.
	suggestions := autoJudger.Suggest(description, 0.5)

	ctx.JSON(http.StatusOK, suggestions)
}

func (s *Server) geocodeView(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "geocode.html", nil)
}

type DatabaseInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type LocationQueueItem struct {
	DbID         int    `json:"db_id"`
	DbName       string `json:"db_name"`
	Location     string `json:"location"`
	OffenseCount int    `json:"offense_count"`
}

func (s *Server) listDatabases(ctx *gin.Context) {
	// Get all databases that have offenses with locations
	sqlRepo, ok := s.geocodeRepo.(*sqlJudgmentRepository)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid repository type"})

		return
	}

	rows, err := sqlRepo.DB().Query(`
		SELECT DISTINCT db_id
		FROM offenses
		WHERE location IS NOT NULL AND location != ''
		ORDER BY db_id
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
	defer rows.Close()

	var databases []DatabaseInfo

	for rows.Next() {
		var dbID int
		if err := rows.Scan(&dbID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		// Get database name
		if dbName, ok := s.dbMap[dbID]; ok {
			databases = append(databases, DatabaseInfo{
				ID:   dbID,
				Name: dbName,
			})
		}
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, databases)
}

func (s *Server) getLocationQueue(ctx *gin.Context) {
	log.Println("getLocationQueue handler called")

	mode := ctx.Query("mode")
	log.Printf("Request mode: %s", mode)

	if mode == "cluster" {
		log.Println("Cluster mode detected")

		dbIDParam := ctx.Query("db_id")

		var dbID *int

		if dbIDParam != "" {
			var id int
			if _, err := fmt.Sscanf(dbIDParam, "%d", &id); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid db_id parameter"})

				return
			}

			dbID = &id
		}

		clusters, err := s.geocodeRepo.GetLocationClusters(dbID)
		if err != nil {
			log.Printf("Error getting clusters: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		log.Printf("GetLocationClusters returned %d clusters", len(clusters))
		ctx.JSON(http.StatusOK, clusters)

		return
	}

	// Check for database filter
	dbIDParam := ctx.Query("db_id")

	// Build base query
	var args []any

	whereClause := ""

	if dbIDParam != "" {
		// Filter by specific database
		var dbID int
		if _, err := fmt.Sscanf(dbIDParam, "%d", &dbID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid db_id parameter"})

			return
		}

		whereClause = " AND o.db_id = ?"

		args = append(args, dbID)
	}

	query := `
		SELECT
			o.db_id,
			o.location,
			COUNT(*) as offense_count
		FROM offenses o
		LEFT JOIN locations lj
			ON o.db_id = lj.db_id AND o.location = lj.location
		WHERE o.location IS NOT NULL
			AND o.location != ''
			AND lj.id IS NULL  -- No judgment exists yet
	` + whereClause + `
		GROUP BY o.db_id, o.location
		ORDER BY offense_count DESC
		LIMIT 1000
	`

	// Get DB handle via type assertion
	sqlRepo, ok := s.geocodeRepo.(*sqlJudgmentRepository)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid repository type"})

		return
	}

	rows, err := sqlRepo.DB().Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
	defer rows.Close()

	var items []LocationQueueItem

	for rows.Next() {
		var item LocationQueueItem
		if err := rows.Scan(&item.DbID, &item.Location, &item.OffenseCount); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		// Lookup database name
		if dbName, ok := s.dbMap[item.DbID]; ok {
			item.DbName = dbName
		} else {
			item.DbName = fmt.Sprintf("DB %d", item.DbID)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, items)
}

type SuggestionResponse struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	IsElectronic    bool    `json:"is_electronic"`
	GeocodingMethod string  `json:"geocoding_method"`
	Confidence      string  `json:"confidence"`
	Notes           string  `json:"notes"`
}

func (s *Server) suggestCoordinates(ctx *gin.Context) {
	dbIDStr := ctx.Param("db_id")
	location := strings.TrimPrefix(ctx.Param("location"), "/")

	var dbID int
	if _, err := fmt.Sscanf(dbIDStr, "%d", &dbID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid db_id"})

		return
	}

	if dbID == 56 { // Tacuarembó hack
		re := regexp.MustCompile(`(?i)\s+FRENTE\s+AL\s+N°\s+`)
		location = re.ReplaceAllString(location, " ")
	}

	// Try RUTA pattern matching first
	if radar, found := s.radarIndex.MatchLocation(location); found {
		ctx.JSON(http.StatusOK, SuggestionResponse{
			Latitude:        radar.Point.Lat,
			Longitude:       radar.Point.Lng,
			IsElectronic:    true,
			GeocodingMethod: "radares_rutas",
			Confidence:      "high",
			Notes:           radar.Descrip,
		})

		return
	}

	// Fallback to standard geocoding
	department := s.dbMap[dbID]

	result, err := s.geocoder.Geocode(location, department)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no suggestion available", "details": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, SuggestionResponse{
		Latitude:        result.Latitude,
		Longitude:       result.Longitude,
		IsElectronic:    false,
		GeocodingMethod: result.Provider,
		Confidence:      result.Confidence,
		Notes:           result.DisplayName,
	})
}

type AcceptJudgmentRequest struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	IsElectronic    bool    `json:"is_electronic"`
	GeocodingMethod string  `json:"geocoding_method"`
	Confidence      string  `json:"confidence"`
	Notes           string  `json:"notes"`
}

func (s *Server) acceptJudgment(ctx *gin.Context) {
	dbIDStr := ctx.Param("db_id")
	location := strings.TrimPrefix(ctx.Param("location"), "/")

	var dbID int
	if _, err := fmt.Sscanf(dbIDStr, "%d", &dbID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid db_id"})

		return
	}

	var req AcceptJudgmentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	// Sanitizar ubicación
	location = sanitizeLocation(location)

	judgment := &Location{
		DbID:     dbID,
		Location: location,
		Point: &spatial.Point{
			Lat: req.Latitude,
			Lng: req.Longitude,
		},
		IsElectronic:    req.IsElectronic,
		GeocodingMethod: req.GeocodingMethod,
		Confidence:      req.Confidence,
		Notes:           req.Notes,
	}

	// Validar judgment antes de guardar
	if err := validateJudgment(judgment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("validación falló: %v", err)})

		return
	}

	if err := s.geocodeRepo.SaveJudgment(judgment); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error al guardar: %v", err)})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

type ProgressResponse struct {
	TotalLocations      int            `json:"total_locations"`
	GeocodedLocations   int            `json:"geocoded_locations"`
	LocationsPercentage float64        `json:"locations_percentage"`
	TotalOffenses       int            `json:"total_offenses"`
	GeocodedOffenses    int            `json:"geocoded_offenses"`
	OffensesPercentage  float64        `json:"offenses_percentage"`
	ByMethod            map[string]int `json:"by_method"`
}

// DescriptionProgressResponse holds statistics for description curation progress.
type DescriptionProgressResponse struct {
	TotalDescriptions      int     `json:"total_descriptions"`
	ClassifiedDescriptions int     `json:"classified_descriptions"`
	DescriptionsPercentage float64 `json:"descriptions_percentage"`
	TotalOffenses          int     `json:"total_offenses"`
	ClassifiedOffenses     int     `json:"classified_offenses"`
	OffensesPercentage     float64 `json:"offenses_percentage"`
}

func (s *Server) getProgress(ctx *gin.Context) {
	// Check for database filter
	dbIDParam := ctx.Query("db_id")

	sqlRepo, ok := s.geocodeRepo.(*sqlJudgmentRepository)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid repository type"})

		return
	}

	db := sqlRepo.DB()

	// Build filter conditions
	var whereClause string

	var args []any

	if dbIDParam != "" {
		var dbID int
		if _, err := fmt.Sscanf(dbIDParam, "%d", &dbID); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid db_id parameter"})

			return
		}

		whereClause = " AND o.db_id = ?"

		args = append(args, dbID)
	}

	// Total unique locations
	var totalLocations int

	query := `
		SELECT COUNT(DISTINCT o.location || '|' || o.db_id)
		FROM offenses o
		WHERE o.location IS NOT NULL AND o.location != ''` + whereClause

	err := db.QueryRow(query, args...).Scan(&totalLocations)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	// Geocoded locations count
	var geocodedLocations int

	judgmentQuery := `
		SELECT COUNT(*)
		FROM locations lj
		WHERE 1=1`
	judgmentArgs := []any{}

	if dbIDParam != "" {
		judgmentQuery += ` AND lj.db_id = ?`
		judgmentArgs = append(judgmentArgs, args...)
	}

	judgmentQuery += ` AND EXISTS (
			SELECT 1 FROM offenses o
			WHERE o.db_id = lj.db_id AND o.location = lj.location` + whereClause + `
		)`

	if whereClause != "" {
		judgmentArgs = append(judgmentArgs, args...)
	}

	err = db.QueryRow(judgmentQuery, judgmentArgs...).Scan(&geocodedLocations)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	// Total offenses
	var totalOffenses int

	offenseQuery := `
		SELECT COUNT(*)
		FROM offenses o
		WHERE o.location IS NOT NULL AND o.location != ''` + whereClause

	err = db.QueryRow(offenseQuery, args...).Scan(&totalOffenses)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	// Offenses with geocoding
	var geocodedOffenses int

	geocodedQuery := `
		SELECT COUNT(*)
		FROM offenses o
		INNER JOIN locations lj
			ON o.db_id = lj.db_id AND o.location = lj.location
		WHERE 1=1` + whereClause

	geocodedArgs := []any{}
	if dbIDParam != "" {
		geocodedArgs = append(geocodedArgs, args...)
	}

	err = db.QueryRow(geocodedQuery, geocodedArgs...).Scan(&geocodedOffenses)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	// By method breakdown
	byMethod := make(map[string]int)
	methodQuery := `
		SELECT geocoding_method, COUNT(*)
		FROM locations lj
		WHERE 1=1`
	methodArgs := []any{}

	if dbIDParam != "" {
		methodQuery += ` AND lj.db_id = ?`
		methodArgs = append(methodArgs, args...)
	}

	methodQuery += ` AND EXISTS (
			SELECT 1 FROM offenses o
			WHERE o.db_id = lj.db_id AND o.location = lj.location` + whereClause + `
		)
		GROUP BY geocoding_method`

	if whereClause != "" {
		methodArgs = append(methodArgs, args...)
	}

	rows, err := db.Query(methodQuery, methodArgs...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
	defer rows.Close()

	for rows.Next() {
		var method string

		var count int
		if err := rows.Scan(&method, &count); err != nil {
			continue
		}

		byMethod[method] = count
	}

	locPct := 0.0
	if totalLocations > 0 {
		locPct = (float64(geocodedLocations) / float64(totalLocations)) * 100
	}

	offPct := 0.0
	if totalOffenses > 0 {
		offPct = (float64(geocodedOffenses) / float64(totalOffenses)) * 100
	}

	ctx.JSON(http.StatusOK, ProgressResponse{
		TotalLocations:      totalLocations,
		GeocodedLocations:   geocodedLocations,
		LocationsPercentage: locPct,
		TotalOffenses:       totalOffenses,
		GeocodedOffenses:    geocodedOffenses,
		OffensesPercentage:  offPct,
		ByMethod:            byMethod,
	})
}

func (s *Server) getDescriptionProgress(ctx *gin.Context) {
	totalDescriptions, classifiedDescriptions, totalOffenses, classifiedOffenses, err := s.descriptionRepo.GetDescriptionProgress()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	descriptionsPercentage := 0.0
	if totalDescriptions > 0 {
		descriptionsPercentage = (float64(classifiedDescriptions) / float64(totalDescriptions)) * 100
	}

	offensesPercentage := 0.0
	if totalOffenses > 0 {
		offensesPercentage = (float64(classifiedOffenses) / float64(totalOffenses)) * 100
	}

	ctx.JSON(http.StatusOK, DescriptionProgressResponse{
		TotalDescriptions:      totalDescriptions,
		ClassifiedDescriptions: classifiedDescriptions,
		DescriptionsPercentage: descriptionsPercentage,
		TotalOffenses:          totalOffenses,
		ClassifiedOffenses:     classifiedOffenses,
		OffensesPercentage:     offensesPercentage,
	})
}

func (s *Server) listJudgments(ctx *gin.Context) {
	page := 1
	perPage := 50

	if p := ctx.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}

	if pp := ctx.Query("per_page"); pp != "" {
		if _, err := fmt.Sscanf(pp, "%d", &perPage); err != nil {
			perPage = 50
		}
	}

	offset := (page - 1) * perPage

	judgments, err := s.geocodeRepo.ListJudgments(nil, nil, perPage, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	total, err := s.geocodeRepo.CountJudgments()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"judgments": judgments,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

type MergeLocationsRequest struct {
	DbID              int    `json:"db_id"`
	TargetLocation    string `json:"target_location"`
	CanonicalLocation string `json:"canonical_location"`
}

func (s *Server) mergeLocations(ctx *gin.Context) {
	var req MergeLocationsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err := s.geocodeRepo.MergeLocations(req.DbID, req.TargetLocation, req.CanonicalLocation); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) descriptionsView(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "descriptions.html", nil)
}

func (s *Server) reviewView(ctx *gin.Context) {
	data, err := s.descriptionRepo.GetReviewAssignments()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.HTML(http.StatusOK, "descriptions_review.html", gin.H{
		"Codes": data,
	})
}

func (s *Server) getUnclassifiedDescriptions(ctx *gin.Context) {
	limit := 1000 // Default limit

	descriptions, err := s.descriptionRepo.GetUnclassifiedDescriptions(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, descriptions)
}

func (s *Server) listArticles(ctx *gin.Context) {
	articles, err := s.descriptionRepo.ListArticles()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, articles)
}

type ClassifyRequest struct {
	Description string   `json:"description"`
	ArticleIDs  []string `json:"article_ids"`
}

func (s *Server) classifyDescription(ctx *gin.Context) {
	var req ClassifyRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	err := s.descriptionRepo.SaveDescriptionClassification(req.Description, req.ArticleIDs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) addArticle(c *gin.Context) {
	var req Article
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	err := s.descriptionRepo.AddArticle(req.ID, req.Text, req.Code, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) searchArticles(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})

		return
	}

	articles, err := s.descriptionRepo.SearchArticles(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, articles)
}
