// Copyright 2025 The ChapaUY Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package curation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoJudger_Suggest(t *testing.T) {
	articles := []Article{
		{ID: "18.9.2", Text: "Estacionar en lugar tarifado sin abonar la tarifa correspondiente."},
		{ID: "4.11", Text: "Circular sin haber realizado la inspección técnica vehicular departamental reglamentaria."},
		{ID: "21.3.1", Text: "Conductor o acompañante sin casco protector."},
	}

	aj := NewDescriptionClassifier(articles)

	tests := []struct {
		name          string
		description   string
		threshold     float64
		expectedIDs   []string
		shouldBeFound bool
	}{
		{
			name:          "Parking without paying",
			description:   "ESTACIONADO SIN ABONAR TARIFA",
			threshold:     0.5,
			expectedIDs:   []string{"18.9.2"},
			shouldBeFound: true,
		},
		{
			name:          "No vehicle inspection",
			description:   "CIRCULAR SIN INSPECCION TECNICA VEHICULAR DEPARTAMENTAL REGLAMENTARIA",
			threshold:     0.5,
			expectedIDs:   []string{"4.11"},
			shouldBeFound: true,
		},
		{
			name:          "No helmet",
			description:   "CONDUCTOR SIN CASCO",
			threshold:     0.5,
			expectedIDs:   []string{"21.3.1"},
			shouldBeFound: true,
		},
		{
			name:          "Irrelevant description",
			description:   "PASEAR AL PERRO",
			threshold:     0.5,
			expectedIDs:   []string{},
			shouldBeFound: false,
		},
		{
			name:          "Multiple articles",
			description:   "CIRCULAR SIN INSPECCION TECNICA VEHICULAR DEPARTAMENTAL REGLAMENTARIA, ESTACIONADO SIN ABONAR TARIFA",
			threshold:     0.5,
			expectedIDs:   []string{"4.11", "18.9.2"},
			shouldBeFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := aj.Suggest(tt.description, tt.threshold)
			if tt.shouldBeFound {
				assert.NotEmpty(t, suggestions)

				var foundIDs []string
				for _, s := range suggestions {
					foundIDs = append(foundIDs, s.ArticleID)
				}

				assert.ElementsMatch(t, tt.expectedIDs, foundIDs)
			} else {
				assert.Empty(t, suggestions)
			}
		})
	}
}

func TestDetectMultiArticle(t *testing.T) {
	articles := []Article{
		{ID: "15.4", Text: "No respetar señales luminosas."},
		{ID: "4.4.1", Text: "Conducir con imprudencia."},
		{ID: "21.8", Text: "No usar chaleco, campera 0 bandas retroreflectivas reglamentaria."},
		{ID: "21.3.1", Text: "Conductor o acompañante sin casco protector."},
	}

	dc := NewDescriptionClassifier(articles)

	tests := []struct {
		name        string
		description string
		threshold   float64
		isMulti     bool
	}{
		{
			name:        "Single article without comma",
			description: "NO CONDUCIR SIN CASCO",
			threshold:   0.5,
			isMulti:     false,
		},
		{
			name:        "Multi-article with distinct violations",
			description: "NO RESPETAR SEÑALES LUMINOSAS, CONDUCIR CON IMPRUDENCIA",
			threshold:   0.5,
			isMulti:     true,
		},
		{
			name:        "Single article with natural commas",
			description: "NO USAR CHALECO, CAMPERA 0 BANDAS RETROREFLECTIVAS",
			threshold:   0.5,
			isMulti:     false,
		},
		{
			name:        "Single article without suggestions",
			description: "PASEAR AL PERRO",
			threshold:   0.5,
			isMulti:     false,
		},
		{
			name:        "Irrelevant with comma",
			description: "BLAH, BLAH",
			threshold:   0.5,
			isMulti:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dc.DetectMultiArticle(tt.description, tt.threshold)
			assert.Equal(t, tt.isMulti, result)
		})
	}
}

func TestSuggestWithBreakdown(t *testing.T) {
	articles := []Article{
		{ID: "15.4", Text: "No respetar señales luminosas."},
		{ID: "4.4.1", Text: "Conducir con imprudencia."},
		{ID: "21.8", Text: "No usar chaleco, campera 0 bandas retroreflectivas reglamentaria."},
	}

	dc := NewDescriptionClassifier(articles)

	tests := []struct {
		name              string
		description       string
		threshold         float64
		expectedPartCount int
	}{
		{
			name:              "Single article",
			description:       "CONDUCIR CON IMPRUDENCIA",
			threshold:         0.5,
			expectedPartCount: 1,
		},
		{
			name:              "Multi-article",
			description:       "NO RESPETAR SEÑALES LUMINOSAS, CONDUCIR CON IMPRUDENCIA",
			threshold:         0.5,
			expectedPartCount: 2,
		},
		{
			name:              "Single article with natural comma shows parts",
			description:       "NO USAR CHALECO, CAMPERA 0 BANDAS RETROREFLECTIVAS",
			threshold:         0.5,
			expectedPartCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breakdown := dc.SuggestWithBreakdown(tt.description, tt.threshold)
			assert.Len(t, breakdown, tt.expectedPartCount)

			for _, bd := range breakdown {
				assert.NotEmpty(t, bd.Part)
			}
		})
	}
}

func TestSuggestWithExactMatch(t *testing.T) {
	// This test verifies that exact matches from memory are returned with score 1.0
	articles := []Article{
		{ID: "15.4", Text: "No respetar señales luminosas."},
		{ID: "4.4.1", Text: "Conducir con imprudencia."},
	}

	// Without classified descriptions (backward compatible)
	dc := NewDescriptionClassifier(articles)
	suggestions := dc.Suggest("CONDUCIR CON IMPRUDENCIA", 0.5)
	assert.NotEmpty(t, suggestions)
	assert.Greater(t, suggestions[0].Score, 0.8) // Should have high similarity but not 1.0

	// With classified descriptions
	classifiedDescriptions := []*Description{
		{Description: "CONDUCIR CON IMPRUDENCIA", ArticleIDs: []string{"4.4.1"}},
		{Description: "SIN CHALECO", ArticleIDs: []string{"21.8"}},
	}
	dcWithCache := NewDescriptionClassifierWithDescriptions(articles, classifiedDescriptions)

	// Exact match should return 1.0
	suggestionsExact := dcWithCache.Suggest("CONDUCIR CON IMPRUDENCIA", 0.5)
	assert.NotEmpty(t, suggestionsExact)
	assert.InDelta(t, 1.0, suggestionsExact[0].Score, 0.0001)
	assert.Equal(t, "4.4.1", suggestionsExact[0].ArticleID)

	// Unknown should still use similarity
	suggestionsUnknown := dcWithCache.Suggest("NO RESPETAR SEÑALES", 0.5)
	assert.NotEmpty(t, suggestionsUnknown)
	assert.Less(t, suggestionsUnknown[0].Score, 1.0) // Similarity match, not exact
}
