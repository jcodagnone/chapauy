// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jcodagnone/chapauy/spatial"
)

// Radar represents a radar location from the GIS layer.
type Radar struct {
	Ruta       int           `json:"ruta"`
	Progresiva string        `json:"progresiva"`
	Gestion    string        `json:"gestion"`
	Descrip    string        `json:"descrip"`
	Point      spatial.Point `json:"point"`
}

// RadarIndex provides fast lookup of radars by route and kilometer marker.
type RadarIndex struct {
	radars map[string]*Radar // key: "ruta:progresiva"
}

// LoadRadares loads the radares_rutas GIS layer from JSON file.
func LoadRadares(filepath string) (*RadarIndex, error) {
	data, err := os.ReadFile(filepath) // #nosec G304 - filepath is provided by admin
	if err != nil {
		return nil, fmt.Errorf("reading radares file: %w", err)
	}

	var geoJSON struct {
		Features []struct {
			Geometry struct {
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				Ruta       int    `json:"ruta"`
				Progresiva string `json:"progresiva"`
				Gestion    string `json:"gestion"`
				Descrip    string `json:"descrip"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.Unmarshal(data, &geoJSON); err != nil {
		return nil, fmt.Errorf("parsing radares JSON: %w", err)
	}

	index := &RadarIndex{
		radars: make(map[string]*Radar),
	}

	for _, feature := range geoJSON.Features {
		// Normalize progresiva: lowercase and remove leading zeros
		progresiva := strings.ToLower(feature.Properties.Progresiva)
		progresiva = normalizeProgresiva(progresiva)

		radar := &Radar{
			Ruta:       feature.Properties.Ruta,
			Progresiva: progresiva,
			Gestion:    feature.Properties.Gestion,
			Descrip:    feature.Properties.Descrip,
			Point: spatial.Point{
				Lng: feature.Geometry.Coordinates[0],
				Lat: feature.Geometry.Coordinates[1],
			},
		}
		key := fmt.Sprintf("%d:%s", radar.Ruta, radar.Progresiva)
		index.radars[key] = radar
	}

	return index, nil
}

// RutaPattern represents a parsed RUTA location string.
type RutaPattern struct {
	OriginalLocation string
	RouteNumber      int
	Progresiva       string // e.g., "38k131", "264k038"
	Direction        string // C, D, or empty
}

// e.g., "038k131" -> "38k131", "025k025" -> "25k025".
func normalizeProgresiva(prog string) string {
	parts := strings.Split(prog, "/")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "k") {
			kmParts := strings.Split(part, "k")
			if len(kmParts) == 2 {
				km := strings.TrimLeft(kmParts[0], "0")
				if km == "" {
					km = "0"
				}

				metersStr := kmParts[1]

				meters, err := strconv.Atoi(metersStr)
				if err != nil {
					// Handle cases where meters might not be a valid number
					meters = 0
				}
				// Format meters to be 3 digits, padding with leading zeros
				metersFormatted := fmt.Sprintf("%03d", meters)

				parts[i] = fmt.Sprintf("%sk%s", km, metersFormatted)
			}
		}
	}

	return strings.Join(parts, "/")
}

// Returns nil if the location doesn't match a RUTA pattern.
func ParseRutaLocation(location string) *RutaPattern {
	location = strings.TrimSpace(location)

	patterns := []*regexp.Regexp{
		// Pattern 1: "Ruta NNN y NNNKNNN_D/C" or "Ruta NNN y NNNKNNN"
		// Allows for spaces within the route number and progresiva, and optional 'R'
		regexp.MustCompile(`(?i)ruta\s*([\d\s]+)\s*R?\s+[yY]\s*([\d\s]+)\s*k\s*([\d\s]+)(?:_([cd]))?`),
		// Pattern 2: "Ruta N y km NNN" or "RUTA NACIONAL N y km NNN"
		regexp.MustCompile(`(?i)ruta(?:\s+nacional)?\s*([\d\s]+)\s+[yY]\s*km\s*([\d\s]+)`),
		// Pattern 3: "NNN y NNNKNNN_D/C" or "NNN y NNNKNNN" (without "Ruta" prefix)
		// Allows for spaces within the route number and progresiva, and optional 'R'
		regexp.MustCompile(`(?i)^([\d\s]+)\s*R?\s+[yY]\s*([\d\s]+)\s*k\s*([\d\s]+)(?:_([cd]))?$`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(location); matches != nil {
			var ruta int

			var progresiva string

			var direction string

			// Clean route number before conversion
			cleanedRoute := strings.ReplaceAll(matches[1], " ", "")
			cleanedRoute = strings.ReplaceAll(cleanedRoute, "R", "")
			ruta, _ = strconv.Atoi(cleanedRoute)

			if pattern == patterns[0] || pattern == patterns[2] { // Pattern 1 and 3
				// Format: NNNkNNN
				cleanedKm := strings.ReplaceAll(matches[2], " ", "")
				cleanedMeters := strings.ReplaceAll(matches[3], " ", "")
				progresiva = fmt.Sprintf("%sk%s", cleanedKm, cleanedMeters)

				if len(matches) >= 5 { // Check if direction group exists
					direction = strings.ToUpper(matches[4])
				}
			} else if pattern == patterns[1] { // Pattern 2
				// Format: km NNN
				cleanedKm := strings.ReplaceAll(matches[2], " ", "")
				progresiva = cleanedKm + "k000"
			}

			if progresiva != "" {
				return &RutaPattern{
					OriginalLocation: location,
					RouteNumber:      ruta,
					Progresiva:       normalizeProgresiva(strings.ToLower(progresiva)),
					Direction:        direction,
				}
			}
		}
	}

	return nil
}

// e.g., "453k110" -> (453, 110), "51k0" -> (51, 0).
func parseProgresiva(prog string) (int, int) {
	parts := strings.Split(strings.ToLower(prog), "k")
	if len(parts) != 2 {
		return 0, 0
	}

	// Clean km and meters parts before conversion
	kmStr := strings.TrimSpace(parts[0])
	metersStr := strings.TrimSpace(parts[1])

	km, _ := strconv.Atoi(kmStr)
	meters, _ := strconv.Atoi(metersStr)

	return km, meters
}

// abs returns absolute value of integer.
func abs(x int) float64 {
	if x < 0 {
		return float64(-x)
	}

	return float64(x)
}

// FindRadar attempts to find a matching radar for the given RUTA pattern.
func (idx *RadarIndex) FindRadar(pattern *RutaPattern) *Radar {
	if pattern == nil {
		return nil
	}

	// Try exact match first
	key := fmt.Sprintf("%d:%s", pattern.RouteNumber, pattern.Progresiva)
	if radar, ok := idx.radars[key]; ok {
		return radar
	}

	// Try without direction suffix
	if pattern.Direction != "" {
		baseProgresiva := strings.TrimSuffix(pattern.Progresiva, "_"+strings.ToLower(pattern.Direction))

		key = fmt.Sprintf("%d:%s", pattern.RouteNumber, baseProgresiva)
		if radar, ok := idx.radars[key]; ok {
			return radar
		}
	}

	// Parse pattern progresiva to extract km and meters
	patternKm, patternMeters := parseProgresiva(pattern.Progresiva)

	// Try fuzzy matching: same route, close kilometer markers
	var bestMatch *Radar

	bestDistance := 1000.0 // Max 1km tolerance

	for k, radar := range idx.radars {
		if !strings.HasPrefix(k, fmt.Sprintf("%d:", pattern.RouteNumber)) {
			continue
		}

		// Check if radar progresiva contains multiple markers (e.g., "51k571/51k278")
		if strings.Contains(radar.Progresiva, "/") {
			markers := strings.SplitSeq(radar.Progresiva, "/")
			for marker := range markers {
				marker = strings.TrimSpace(marker)
				if marker == pattern.Progresiva {
					return radar
				}

				// Check proximity for each marker in range
				radarKm, radarMeters := parseProgresiva(marker)
				if radarKm == patternKm {
					distance := abs(radarMeters - patternMeters)
					if distance < bestDistance {
						bestDistance = distance
						bestMatch = radar
					}
				}
			}
		} else {
			// Single marker - check proximity
			radarKm, radarMeters := parseProgresiva(radar.Progresiva)
			if radarKm == patternKm {
				distance := abs(radarMeters - patternMeters)
				if distance < bestDistance {
					bestDistance = distance
					bestMatch = radar
				}
			}
		}
	}

	return bestMatch
}

// Returns the radar and whether it's an electronic surveillance location.
func (idx *RadarIndex) MatchLocation(location string) (*Radar, bool) {
	pattern := ParseRutaLocation(location)
	if pattern == nil {
		return nil, false
	}

	radar := idx.FindRadar(pattern)

	return radar, radar != nil
}
