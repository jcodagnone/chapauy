// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

// GeocodingResult represents a geocoding result from any provider.
type GeocodingResult struct {
	Latitude    float64
	Longitude   float64
	Confidence  string // high, medium, low
	Provider    string
	DisplayName string
}

// Geocoder interface for different geocoding providers.
type Geocoder interface {
	Geocode(location string, department string) (*GeocodingResult, error)
}
