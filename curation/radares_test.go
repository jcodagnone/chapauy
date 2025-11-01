// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"testing"
)

func TestParseRutaLocation(t *testing.T) {
	tests := []struct {
		name       string
		location   string
		wantRoute  int
		wantProg   string
		wantDir    string
		shouldFail bool
	}{
		{
			name:      "Standard format with direction D",
			location:  "Ruta 005 y 038K131_D",
			wantRoute: 5,
			wantProg:  "38k131",
			wantDir:   "D",
		},
		{
			name:      "Standard format with direction C",
			location:  "Ruta 001 y 051K278_C",
			wantRoute: 1,
			wantProg:  "51k278",
			wantDir:   "C",
		},
		{
			name:      "Without direction suffix",
			location:  "Ruta 009 y 264K038",
			wantRoute: 9,
			wantProg:  "264k038",
			wantDir:   "",
		},
		{
			name:      "Three digit route",
			location:  "Ruta 102 y 045K900",
			wantRoute: 102,
			wantProg:  "45k900",
			wantDir:   "",
		},
		{
			name:      "Single digit route with km notation",
			location:  "Ruta 3 y km 453",
			wantRoute: 3,
			wantProg:  "453k000",
			wantDir:   "",
		},
		{
			name:      "Nacional route with km",
			location:  "RUTA NACIONAL 3 y km 383",
			wantRoute: 3,
			wantProg:  "383k000",
			wantDir:   "",
		},
		{
			name:      "Mixed case",
			location:  "RUTA 008 Y 025K025_D",
			wantRoute: 8,
			wantProg:  "25k025",
			wantDir:   "D",
		},
		{
			name:      "Extra spaces",
			location:  "Ruta  102  y  024K220_D",
			wantRoute: 102,
			wantProg:  "24k220",
			wantDir:   "D",
		},
		{
			name:       "Not a RUTA pattern - street intersection",
			location:   "AV 8 DE OCTUBRE Y AV CENTENARIO",
			shouldFail: true,
		},
		{
			name:       "Not a RUTA pattern - named location",
			location:   "Ruta Interbalnearia y Milton Lussich",
			shouldFail: true,
		},
		{
			name:       "Not a RUTA pattern - avenue named Ruta",
			location:   "Ruta AV W F ALDUNATE y AV C RACINE",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := ParseRutaLocation(tt.location)

			if tt.shouldFail {
				if pattern != nil {
					t.Errorf("Expected parsing to fail for %q, but got pattern: %+v", tt.location, pattern)
				}

				return
			}

			if pattern == nil {
				t.Fatalf("ParseRutaLocation(%q) returned nil, expected pattern", tt.location)
			}

			if pattern.RouteNumber != tt.wantRoute {
				t.Errorf("RouteNumber = %d, want %d", pattern.RouteNumber, tt.wantRoute)
			}

			if pattern.Progresiva != tt.wantProg {
				t.Errorf("Progresiva = %q, want %q", pattern.Progresiva, tt.wantProg)
			}

			if pattern.Direction != tt.wantDir {
				t.Errorf("Direction = %q, want %q", pattern.Direction, tt.wantDir)
			}

			if pattern.OriginalLocation != tt.location {
				t.Errorf("OriginalLocation = %q, want %q", pattern.OriginalLocation, tt.location)
			}
		})
	}
}

func TestLoadRadares(t *testing.T) {
	index, err := LoadRadares("radares.json")
	if err != nil {
		t.Fatalf("LoadRadares() error = %v", err)
	}

	if len(index.radars) == 0 {
		t.Fatal("LoadRadares() returned empty index")
	}

	// Check that we loaded approximately the expected number
	if len(index.radars) < 100 {
		t.Errorf("Expected at least 100 radars, got %d", len(index.radars))
	}

	t.Logf("Loaded %d radars", len(index.radars))
}

func TestFindRadar(t *testing.T) {
	index, err := LoadRadares("radares.json")
	if err != nil {
		t.Fatalf("LoadRadares() error = %v", err)
	}

	tests := []struct {
		name        string
		location    string
		shouldMatch bool
		wantRoute   int
		wantLat     float64 // approximate
		wantLon     float64 // approximate
	}{
		{
			name:        "Match Ruta 3 km 453 (fuzzy - radar at 453k110)",
			location:    "Ruta 3 y km 453",
			shouldMatch: true,
			wantRoute:   3,
		},
		{
			name:        "Match Ruta 1 51k571 (range match)",
			location:    "Ruta 001 y 051K571_D",
			shouldMatch: true,
			wantRoute:   1,
		},
		{
			name:        "Match Ruta 1 51k278 (range match alternate)",
			location:    "Ruta 001 y 051K278_C",
			shouldMatch: true,
			wantRoute:   1,
		},
		{
			name:        "Match Ruta 5 38K131",
			location:    "Ruta 005 y 038K131_D",
			shouldMatch: true,
			wantRoute:   5,
			wantLat:     -34.59, // approximate
			wantLon:     -56.26,
		},
		{
			name:        "Match Ruta 8 25K025",
			location:    "Ruta 008 y 025K025_D",
			shouldMatch: true,
			wantRoute:   8,
		},
		{
			name:        "Match Ruta 9 264K038",
			location:    "Ruta 009 y 264K038",
			shouldMatch: true,
			wantRoute:   9,
		},
		{
			name:        "No match - non-existent marker",
			location:    "Ruta 999 y 999K999_D",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := ParseRutaLocation(tt.location)
			if pattern == nil {
				t.Fatalf("Failed to parse location %q", tt.location)
			}

			radar := index.FindRadar(pattern)

			if tt.shouldMatch && radar == nil {
				t.Errorf("Expected to find radar for %q, but got nil", tt.location)

				return
			}

			if !tt.shouldMatch && radar != nil {
				t.Errorf("Expected no match for %q, but found radar: %+v", tt.location, radar)

				return
			}

			if radar != nil {
				if radar.Ruta != tt.wantRoute {
					t.Errorf("Ruta = %d, want %d", radar.Ruta, tt.wantRoute)
				}

				if tt.wantLat != 0 {
					latDiff := radar.Point.Lat - tt.wantLat
					if latDiff < -0.01 || latDiff > 0.01 {
						t.Errorf("Latitude = %f, want approximately %f", radar.Point.Lat, tt.wantLat)
					}
				}

				if tt.wantLon != 0 {
					lonDiff := radar.Point.Lng - tt.wantLon
					if lonDiff < -0.01 || lonDiff > 0.01 {
						t.Errorf("Longitude = %f, want approximately %f", radar.Point.Lng, tt.wantLon)
					}
				}

				t.Logf("Found radar: Ruta %d, %s at (%f, %f) - %s",
					radar.Ruta, radar.Progresiva, radar.Point.Lat, radar.Point.Lng, radar.Descrip)
			}
		})
	}
}

func TestMatchLocation(t *testing.T) {
	index, err := LoadRadares("radares.json")
	if err != nil {
		t.Fatalf("LoadRadares() error = %v", err)
	}

	tests := []struct {
		name           string
		location       string
		wantElectronic bool
	}{
		{
			name:           "Electronic - RUTA with radar",
			location:       "Ruta 005 y 038K131_D",
			wantElectronic: true,
		},
		{
			name:           "Non-electronic - street intersection",
			location:       "AV 8 DE OCTUBRE Y AV CENTENARIO",
			wantElectronic: false,
		},
		{
			name:           "Non-electronic - named ruta",
			location:       "Ruta Interbalnearia y Milton Lussich",
			wantElectronic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar, isElectronic := index.MatchLocation(tt.location)

			if isElectronic != tt.wantElectronic {
				t.Errorf("isElectronic = %v, want %v for %q", isElectronic, tt.wantElectronic, tt.location)
			}

			if isElectronic && radar == nil {
				t.Error("isElectronic is true but radar is nil")
			}

			if radar != nil {
				t.Logf("Matched: %s -> Ruta %d, %s (%f, %f)",
					tt.location, radar.Ruta, radar.Descrip, radar.Point.Lat, radar.Point.Lng)
			}
		})
	}
}
