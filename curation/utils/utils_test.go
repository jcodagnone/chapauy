// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLowerAsciiFolding(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello world"},
		{"  Spaces  ", "spaces"},
		{"Áéíóú", "aeiou"},
		{"Ñandú", "nandu"},
		{"Crème Brûlée", "creme brulee"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, LowerASCIIFolding(tc.input))
		})
	}
}

func TestAnyToInt8Slice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []int8
		ok       bool
	}{
		{"nil", nil, nil, true},
		{"[]int8", []int8{1, 2, 3}, []int8{1, 2, 3}, true},
		{"[]int64 valid", []int64{1, 2, 3}, []int8{1, 2, 3}, true},
		{"[]int64 overflow", []int64{128}, nil, false},
		{"[]int64 underflow", []int64{-129}, nil, false},
		{"[]any int8", []any{int8(1), int8(2)}, []int8{1, 2}, true},
		{"[]any int64 valid", []any{int64(1), int64(2)}, []int8{1, 2}, true},
		{"[]any int64 overflow", []any{int64(128)}, nil, false},
		{"[]any mixed invalid", []any{int8(1), "string"}, nil, false},
		{"not a slice", "string", nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, ok := AnyToInt8Slice(tc.input)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestAnyToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
		ok       bool
	}{
		{"nil", nil, nil, true},
		{"[]string", []string{"a", "b"}, []string{"a", "b"}, true},
		{"[]any string", []any{"a", "b"}, []string{"a", "b"}, true},
		{"[]any mixed invalid", []any{"a", 1}, nil, false},
		{"not a slice", 123, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, ok := AnyToStringSlice(tc.input)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestResolveMultiArticle(t *testing.T) {
	mockClassifier := func(part string) (Classification, bool, error) {
		switch part {
		case "part1":
			return Classification{ArticleIDs: []string{"1"}, ArticleCodes: []int8{1}}, true, nil
		case "part2":
			return Classification{ArticleIDs: []string{"2"}, ArticleCodes: []int8{2}}, true, nil
		case "error":
			return Classification{}, false, assert.AnError
		default:
			return Classification{}, false, nil
		}
	}

	tests := []struct {
		name        string
		description string
		expected    Classification
		found       bool
		hasError    bool
	}{
		{
			name:        "Single part match",
			description: "part1",
			expected:    Classification{ArticleIDs: []string{"1"}, ArticleCodes: []int8{1}},
			found:       true,
		},
		{
			name:        "Multi part match",
			description: "part1, part2",
			expected:    Classification{ArticleIDs: []string{"1", "2"}, ArticleCodes: []int8{1, 2}},
			found:       true,
		},
		{
			name:        "Partial match",
			description: "part1, unknown",
			expected:    Classification{},
			found:       false,
		},
		{
			name:        "Empty description",
			description: "",
			expected:    Classification{},
			found:       false,
		},
		{
			name:        "Empty parts",
			description: "part1, , part2",
			expected:    Classification{ArticleIDs: []string{"1", "2"}, ArticleCodes: []int8{1, 2}},
			found:       true,
		},
		{
			name:        "Classifier error",
			description: "part1, error",
			expected:    Classification{},
			found:       false,
			hasError:    true,
		},
		{
			name:        "Whitespace trimming",
			description: "  part1  ,  part2  ",
			expected:    Classification{ArticleIDs: []string{"1", "2"}, ArticleCodes: []int8{1, 2}},
			found:       true,
		},
		{
			name:        "Only commas and spaces",
			description: " , , ",
			expected:    Classification{},
			found:       false,
		},
		{
			name:        "Case preservation",
			description: "Part1",          // Mock classifier needs to handle "Part1" if we want it to match, or we verify it passes "Part1"
			expected:    Classification{}, // Mock only handles "part1", so "Part1" should fail if not normalized by ResolveMultiArticle
			found:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, found, err := ResolveMultiArticle(tc.description, mockClassifier)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.found, found)

			if found {
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{12, "12"},
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{-1, "-1"},
		{-12, "-12"},
		{-123, "-123"},
		{-1234, "-1,234"},
		{-12345, "-12,345"},
		{-123456, "-123,456"},
		{-1234567, "-1,234,567"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, FormatInt(tc.input))
		})
	}
}
