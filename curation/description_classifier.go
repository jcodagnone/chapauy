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
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/jcodagnone/chapauy/curation/utils"
)

// Suggestion represents a suggested article for a given description.
type Suggestion struct {
	ArticleID string  `json:"ArticleID"`
	Text      string  `json:"Text"`
	Score     float64 `json:"Score"`
}

// DescriptionClassifier suggests articles for a given description based on cosine similarity.
// It pre-processes a set of known articles into vector representations (bag-of-words)
// to efficiently compare new descriptions against them.
// It also caches already-classified descriptions in memory for exact match lookup.
type DescriptionClassifier struct {
	articles              []Article                 // The list of all known regulation articles
	vectors               map[string]map[string]int // Pre-computed word vectors for each article, keyed by ArticleID
	classifiedByDesc      map[string][]string       // Cache of classified descriptions: description -> article_ids
	classifiedByDescLower map[string]string         // Lowercase version for case-insensitive lookup: lowercase -> original
}

// NewDescriptionClassifier creates a new DescriptionClassifier.
// It initializes the classifier by vectorizing all provided articles.
func NewDescriptionClassifier(articles []Article) *DescriptionClassifier {
	return NewDescriptionClassifierWithDescriptions(articles, nil)
}

// NewDescriptionClassifierWithDescriptions creates a new DescriptionClassifier with pre-loaded classified descriptions.
// This loads descriptions into memory for fast exact-match lookup, similar to how articles are cached.
func NewDescriptionClassifierWithDescriptions(articles []Article, classifiedDescriptions []*Description) *DescriptionClassifier {
	dc := &DescriptionClassifier{
		articles:              articles,
		vectors:               make(map[string]map[string]int),
		classifiedByDesc:      make(map[string][]string),
		classifiedByDescLower: make(map[string]string),
	}

	// Pre-vectorize all articles for faster lookups
	for _, article := range articles {
		dc.vectors[article.ID] = vectorize(article.Text)
	}

	// Load classified descriptions into memory
	for _, desc := range classifiedDescriptions {
		dc.classifiedByDesc[desc.Description] = desc.ArticleIDs
		// Store lowercase version for case-insensitive lookup
		dc.classifiedByDescLower[utils.LowerASCIIFolding(desc.Description)] = desc.Description
	}

	return dc
}

// Suggest returns a list of suggested articles for a given description.
// It handles composite descriptions by analyzing the full string and then each comma-separated part.
// The suggestions are de-duplicated, keeping the highest score for each article, and then sorted by score.
func (dc *DescriptionClassifier) Suggest(description string, threshold float64) []Suggestion {
	allSuggestions := make(map[string]Suggestion) // Use a map to store unique suggestions by ArticleID

	// 1. Analyze the whole description first
	for _, s := range dc.suggest(description, threshold) {
		allSuggestions[s.ArticleID] = s
	}

	// 2. If the description contains a comma, analyze each part separately
	// This helps in cases where multiple distinct offenses are listed in one description.
	parts := strings.Split(description, ",")
	if len(parts) > 1 {
		for _, part := range parts {
			// Trim spaces around the comma-separated part
			for _, s := range dc.suggest(strings.TrimSpace(part), threshold) {
				// If a suggestion for this ArticleID already exists, we keep the one with the higher score.
				if existing, ok := allSuggestions[s.ArticleID]; !ok || s.Score > existing.Score {
					allSuggestions[s.ArticleID] = s
				}
			}
		}
	}

	// Convert the map of unique suggestions back into a slice
	result := make([]Suggestion, 0, len(allSuggestions))
	for _, s := range allSuggestions {
		result = append(result, s)
	}

	// Sort the final list of suggestions in descending order of their similarity score
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result
}

// suggest performs the core similarity analysis on a single string (either the full description or a part of it).
// It converts the description into a word vector and then calculates its cosine similarity against
// all pre-vectorized articles, returning suggestions that meet the specified threshold.
// If an exact match exists in the in-memory cache, it's returned with score 1.0 (perfect match).
func (dc *DescriptionClassifier) suggest(description string, threshold float64) []Suggestion {
	var suggestions []Suggestion

	trimmedDesc := strings.TrimSpace(description)

	// Check if this exact description is already classified (check both exact and case-insensitive)
	var articleIDs []string
	if ids, ok := dc.classifiedByDesc[trimmedDesc]; ok {
		articleIDs = ids
	} else {
		// Try case-insensitive lookup
		lowerDesc := utils.LowerASCIIFolding(trimmedDesc)
		if originalDesc, ok := dc.classifiedByDescLower[lowerDesc]; ok {
			articleIDs = dc.classifiedByDesc[originalDesc]
		}
	}

	if len(articleIDs) > 0 {
		// Found an exact match in the cache - return it with perfect score
		for _, articleID := range articleIDs {
			// Find the article details
			for _, article := range dc.articles {
				if article.ID == articleID {
					suggestions = append(suggestions, Suggestion{
						ArticleID: article.ID,
						Text:      article.Text,
						Score:     1.0, // Perfect match - already classified
					})

					break
				}
			}
		}
		// Return exact matches (they should be sorted by score, which is 1.0)
		sort.Slice(suggestions, func(i, j int) bool {
			return suggestions[i].Score > suggestions[j].Score
		})

		return suggestions
	}

	descVector := vectorize(trimmedDesc) // Vectorize the input description

	for _, article := range dc.articles {
		articleVector := dc.vectors[article.ID]              // Retrieve the pre-computed vector for the article
		score := cosineSimilarity(descVector, articleVector) // Calculate cosine similarity

		// If the similarity score meets the threshold, add it as a suggestion
		if score >= threshold {
			suggestions = append(suggestions, Suggestion{
				ArticleID: article.ID,
				Text:      article.Text,
				Score:     score,
			})
		}
	}

	return suggestions
}

// nonAlphanumericRegex is used to remove non-alphanumeric characters during text cleaning.
var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

// cleanString removes non-alphanumeric characters from a string.
func cleanString(s string) string {
	return nonAlphanumericRegex.ReplaceAllString(utils.LowerASCIIFolding(s), "")
}

// vectorize converts a given text into a bag-of-words frequency map (vector).
// This involves cleaning, lowercasing, tokenization, and accent removal.
func vectorize(text string) map[string]int {
	// Clean the string (remove special characters) and convert to lowercase
	text = cleanString(text)
	words := strings.Fields(text) // Split the text into individual words (tokens)
	vector := make(map[string]int)

	for _, word := range words {
		vector[word]++ // Increment word count in the vector
	}

	return vector
}

// cosineSimilarity calculates the cosine similarity between two word vectors (frequency maps).
// Cosine similarity measures the cosine of the angle between two vectors, ranging from 0 to 1,
// indicating how similar the documents are regardless of their size.
// A score of 1 means the vectors are identical, 0 means they are completely dissimilar.
func cosineSimilarity(v1, v2 map[string]int) float64 {
	dotProduct := 0 // Stores the dot product of the two vectors

	// Calculate the dot product: sum of (v1[word] * v2[word]) for common words
	for k, v := range v1 {
		if v2[k] > 0 { // Only consider words present in both vectors
			dotProduct += v * v2[k]
		}
	}

	mag1 := 0 // Magnitude of vector 1 (sum of squares of word counts)
	for _, v := range v1 {
		mag1 += v * v
	}

	mag2 := 0 // Magnitude of vector 2 (sum of squares of word counts)
	for _, v := range v2 {
		mag2 += v * v
	}

	// Avoid division by zero if either vector has zero magnitude (empty or no relevant words)
	if mag1 == 0 || mag2 == 0 {
		return 0
	}

	// Cosine similarity formula: dotProduct / (magnitude1 * magnitude2)
	return float64(dotProduct) / (math.Sqrt(float64(mag1)) * math.Sqrt(float64(mag2)))
}

// DetectMultiArticle returns true if the description appears to have multiple distinct articles.
// It compares the article suggestions from each comma-separated part. If parts suggest different
// high-confidence articles, it's a multi-article description.
func (dc *DescriptionClassifier) DetectMultiArticle(description string, threshold float64) bool {
	// No commas means single-article
	if !strings.Contains(description, ",") {
		return false
	}

	// Analyze each comma-separated part
	parts := strings.Split(description, ",")
	if len(parts) <= 1 {
		return false
	}

	var partArticleIDSets []map[string]bool

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}

		articleIDs := make(map[string]bool)
		for _, s := range dc.suggest(trimmedPart, threshold) {
			articleIDs[s.ArticleID] = true
		}

		if len(articleIDs) > 0 {
			partArticleIDSets = append(partArticleIDSets, articleIDs)
		}
	}

	// If we have fewer than 2 parts with suggestions, it's single-article
	if len(partArticleIDSets) < 2 {
		return false
	}

	// Check if any two parts have different article sets
	for i := range len(partArticleIDSets) - 1 {
		currentSet := partArticleIDSets[i]
		nextSet := partArticleIDSets[i+1]

		// If the sets are different (not identical), it's multi-article
		if !articlesMatch(currentSet, nextSet) {
			return true
		}
	}

	return false
}

// articlesMatch checks if two article ID sets are identical.
func articlesMatch(set1, set2 map[string]bool) bool {
	if len(set1) != len(set2) {
		return false
	}

	for id := range set1 {
		if !set2[id] {
			return false
		}
	}

	return true
}

// SuggestionBreakdown represents suggestions grouped by comma-separated parts.
type SuggestionBreakdown struct {
	Part        string       `json:"part"`
	Suggestions []Suggestion `json:"suggestions"`
}

// SuggestWithBreakdown returns suggestions grouped by comma-separated parts.
// This is useful for display to show which part maps to which articles.
func (dc *DescriptionClassifier) SuggestWithBreakdown(description string, threshold float64) []SuggestionBreakdown {
	// If no commas, return single breakdown for the whole description
	if !strings.Contains(description, ",") {
		return []SuggestionBreakdown{
			{
				Part:        description,
				Suggestions: dc.suggest(description, threshold),
			},
		}
	}

	// Analyze each comma-separated part
	parts := strings.Split(description, ",")

	breakdown := make([]SuggestionBreakdown, 0, len(parts))

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}

		breakdown = append(breakdown, SuggestionBreakdown{
			Part:        trimmedPart,
			Suggestions: dc.suggest(trimmedPart, threshold),
		})
	}

	return breakdown
}
