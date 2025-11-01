// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// GoogleMapsGeocoder uses Google Maps Geocoding API.
type GoogleMapsGeocoder struct {
	apiKey     string
	httpClient *http.Client
}

// NewGoogleMapsGeocoder creates a new Google Maps geocoder.
func NewGoogleMapsGeocoder(apiKey string) *GoogleMapsGeocoder {
	return &GoogleMapsGeocoder{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type googleMapsResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"` // ROOFTOP, RANGE_INTERPOLATED, GEOMETRIC_CENTER, APPROXIMATE
		} `json:"geometry"`
		FormattedAddress string `json:"formatted_address"`
	} `json:"results"`
	Status string `json:"status"` // OK, ZERO_RESULTS, etc.
}

func (g *GoogleMapsGeocoder) Geocode(location string, department string) (*GeocodingResult, error) {
	// Build search query with department context
	// Google Maps handles intersections natively (unlike Nominatim which needs splitting)
	var searchQuery string
	if department == "" {
		searchQuery = location + ", Uruguay"
	} else {
		searchQuery = fmt.Sprintf("%s, %s, Uruguay", location, department)
	}

	params := url.Values{}
	params.Set("address", searchQuery)
	params.Set("key", g.apiKey)
	params.Set("region", "uy") // Bias to Uruguay

	reqURL := "https://maps.googleapis.com/maps/api/geocode/json?" + params.Encode()

	resp, err := g.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google maps returned status %d", resp.StatusCode)
	}

	var gmResp googleMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&gmResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if gmResp.Status != "OK" {
		return nil, fmt.Errorf("google maps status: %s", gmResp.Status)
	}

	if len(gmResp.Results) == 0 {
		return nil, fmt.Errorf("no results found for location: %s", location)
	}

	result := gmResp.Results[0]

	// Determine confidence based on location_type
	// Google Maps excels at intersections (RANGE_INTERPOLATED or GEOMETRIC_CENTER)
	confidence := "low"

	switch result.Geometry.LocationType {
	case "ROOFTOP":
		confidence = "high"
	case "RANGE_INTERPOLATED":
		confidence = "high" // Common for intersections - Google handles these well
	case "GEOMETRIC_CENTER":
		confidence = "medium" // Also good for intersections
	case "APPROXIMATE":
		confidence = "low"
	}

	return &GeocodingResult{
		Latitude:    result.Geometry.Location.Lat,
		Longitude:   result.Geometry.Location.Lng,
		Confidence:  confidence,
		Provider:    "google_maps",
		DisplayName: result.FormattedAddress,
	}, nil
}
