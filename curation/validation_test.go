// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"testing"

	"github.com/jcodagnone/chapauy/spatial"
)

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{
			name:    "valid montevideo coordinates",
			lat:     -34.9011,
			lon:     -56.1645,
			wantErr: false,
		},
		{
			name:    "valid maldonado coordinates",
			lat:     -34.9234,
			lon:     -54.9483,
			wantErr: false,
		},
		{
			name:    "latitude too high",
			lat:     91.0,
			lon:     -56.0,
			wantErr: true,
		},
		{
			name:    "latitude too low",
			lat:     -91.0,
			lon:     -56.0,
			wantErr: true,
		},
		{
			name:    "longitude too high",
			lat:     -34.0,
			lon:     181.0,
			wantErr: true,
		},
		{
			name:    "longitude too low",
			lat:     -34.0,
			lon:     -181.0,
			wantErr: true,
		},
		{
			name:    "outside uruguay - too far north",
			lat:     -28.0,
			lon:     -56.0,
			wantErr: true,
		},
		{
			name:    "outside uruguay - too far south",
			lat:     -37.0,
			lon:     -56.0,
			wantErr: true,
		},
		{
			name:    "outside uruguay - too far east",
			lat:     -34.0,
			lon:     -51.0,
			wantErr: true,
		},
		{
			name:    "outside uruguay - too far west",
			lat:     -34.0,
			lon:     -60.0,
			wantErr: true,
		},
		{
			name:    "edge case - north boundary",
			lat:     -29.0,
			lon:     -56.0,
			wantErr: false,
		},
		{
			name:    "edge case - south boundary",
			lat:     -36.0,
			lon:     -56.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCoordinates(tt.lat, tt.lon)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCoordinates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJudgment(t *testing.T) {
	validLat := -34.9011
	validLon := -56.1645

	tests := []struct {
		name    string
		j       *Location
		wantErr bool
	}{
		{
			name: "valid judgment",
			j: &Location{
				DbID:     6,
				Location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
				Point: &spatial.Point{
					Lat: validLat,
					Lng: validLon,
				},
				GeocodingMethod: "google_maps",
				Confidence:      "high",
				Notes:           "Test location",
			},
			wantErr: false,
		},
		{
			name:    "nil judgment",
			j:       nil,
			wantErr: true,
		},
		{
			name: "empty location",
			j: &Location{
				DbID:     6,
				Location: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only location",
			j: &Location{
				DbID:     6,
				Location: "   ",
			},
			wantErr: true,
		},
		{
			name: "location too long",
			j: &Location{
				DbID:     6,
				Location: string(make([]byte, 501)),
			},
			wantErr: true,
		},
		{
			name: "invalid coordinates",
			j: &Location{
				DbID:     6,
				Location: "Test",
				Point: &spatial.Point{
					Lat: 91.0,
					Lng: validLon,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid geocoding method",
			j: &Location{
				DbID:            6,
				Location:        "Test",
				GeocodingMethod: "invalid_method",
			},
			wantErr: true,
		},
		{
			name: "invalid confidence",
			j: &Location{
				DbID:       6,
				Location:   "Test",
				Confidence: "invalid_confidence",
			},
			wantErr: true,
		},
		{
			name: "notes too long",
			j: &Location{
				DbID:     6,
				Location: "Test",
				Notes:    string(make([]byte, 1001)),
			},
			wantErr: true,
		},
		{
			name: "valid judgment without coordinates",
			j: &Location{
				DbID:     6,
				Location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJudgment(tt.j)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJudgment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     string
	}{
		{
			name:     "normal location",
			location: "AV 8 DE OCTUBRE Y AV CENTENARIO",
			want:     "AV 8 DE OCTUBRE Y AV CENTENARIO",
		},
		{
			name:     "location with leading whitespace",
			location: "  AV 8 DE OCTUBRE",
			want:     "AV 8 DE OCTUBRE",
		},
		{
			name:     "location with trailing whitespace",
			location: "AV 8 DE OCTUBRE  ",
			want:     "AV 8 DE OCTUBRE",
		},
		{
			name:     "location with both leading and trailing whitespace",
			location: "  AV 8 DE OCTUBRE  ",
			want:     "AV 8 DE OCTUBRE",
		},
		{
			name:     "location too long gets truncated",
			location: string(make([]byte, 600)),
			want:     string(make([]byte, 500)),
		},
		{
			name:     "empty location",
			location: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLocation(tt.location)
			if got != tt.want {
				t.Errorf("SanitizeLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
