// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// LowerASCIIFolding normalizes a string by removing accents, lowercasing, and trimming spaces.
func LowerASCIIFolding(s string) string {
	s, _, _ = transform.String(
		transform.Chain(
			norm.NFD,
			runes.Remove(runes.In(unicode.Mn)),
			norm.NFC,
		),
		strings.TrimSpace(strings.ToLower(s)),
	)

	return s
}

// AnyToInt8Slice converts an interface{} to []int8 safely.
func AnyToInt8Slice(v any) ([]int8, bool) {
	if v == nil {
		return nil, true
	}

	if i, ok := v.([]int8); ok {
		return i, true
	}

	if i, ok := v.([]int64); ok {
		s := make([]int8, len(i))

		for j, e := range i {
			if e < -128 || e > 127 { // Check for int8 overflow
				return nil, false // Value out of int8 range
			}

			s[j] = int8(e)
		}

		return s, true
	}

	if i, ok := v.([]any); ok {
		s := make([]int8, len(i))

		for j, e := range i {
			val, ok := e.(int8)
			if !ok {
				if val64, ok := e.(int64); ok {
					if val64 < -128 || val64 > 127 {
						return nil, false // Value out of int8 range
					}

					s[j] = int8(val64)

					continue
				}

				return nil, false
			}

			s[j] = val
		}

		return s, true
	}

	return nil, false
}

// AnyToStringSlice converts an interface{} to []string safely.
func AnyToStringSlice(v any) ([]string, bool) {
	if v == nil {
		return nil, true
	}

	if i, ok := v.([]string); ok {
		return i, true
	}

	if i, ok := v.([]any); ok {
		s := make([]string, len(i))

		for j, e := range i {
			val, ok := e.(string)
			if !ok {
				return nil, false
			}

			s[j] = val
		}

		return s, true
	}

	return nil, false
}

// Classification represents the article IDs and codes associated with a description.
type Classification struct {
	ArticleIDs   []string
	ArticleCodes []int8
}

// ClassifierFunc returns the classification for a description part.
// It returns the classification, a boolean indicating if it was found, and an error if the lookup failed.
type ClassifierFunc func(part string) (Classification, bool, error)

// ResolveMultiArticle checks if all parts of a description are classified and returns the aggregated classification.
// It splits the description by comma and checks each part using the provided classifier function.
func ResolveMultiArticle(description string, classify ClassifierFunc) (Classification, bool, error) {
	parts := strings.Split(description, ",")

	var result Classification

	allFound := true
	hasParts := false

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		hasParts = true

		info, found, err := classify(part)
		if err != nil {
			return Classification{}, false, err
		}

		if !found {
			allFound = false

			break
		}

		result.ArticleIDs = append(result.ArticleIDs, info.ArticleIDs...)
		result.ArticleCodes = append(result.ArticleCodes, info.ArticleCodes...)
	}

	if !hasParts || !allFound {
		return Classification{}, false, nil
	}

	return result, true, nil
}

// FormatInt formats an integer with commas for human readability.
func FormatInt(n int64) string {
	in := strconv.FormatInt(n, 10)

	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}

	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}

		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

// ToRoman converts an integer to a Roman numeral.
func ToRoman(num int) string {
	if num <= 0 {
		return ""
	}

	val := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syb := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	roman := ""
	i := 0

	for num > 0 {
		for num >= val[i] {
			roman += syb[i]
			num -= val[i]
		}

		i++
	}

	return roman
}
