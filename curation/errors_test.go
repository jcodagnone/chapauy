// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"errors"
	"testing"
)

type errorCheckTestCase struct {
	name string
	err  error
	want bool
}

func runErrorCheckTest(t *testing.T, tests []errorCheckTestCase, checkFunc func(error) bool) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkFunc(tt.err); got != tt.want {
				t.Errorf("checkFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "rate limit error type",
			err: &GeocodingError{
				Type:    ErrorTypeRateLimit,
				Message: "rate limit exceeded",
			},
			want: true,
		},
		{
			name: "error message contains rate limit",
			err:  errors.New("rate limit exceeded"),
			want: true,
		},
		{
			name: "error message contains too many requests",
			err:  errors.New("too many requests"),
			want: true,
		},
		{
			name: "error message contains 429",
			err:  errors.New("nominatim returned status 429"),
			want: true,
		},
		{
			name: "other error type",
			err: &GeocodingError{
				Type:    ErrorTypeNotFound,
				Message: "not found",
			},
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimitError(tt.err); got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsQuotaExceededError(t *testing.T) {
	tests := []errorCheckTestCase{
		{
			name: "quota exceeded error type",
			err: &GeocodingError{
				Type:    ErrorTypeQuotaExceeded,
				Message: "quota exceeded",
			},
			want: true,
		},
		{
			name: "error message contains over_query_limit",
			err:  errors.New("google maps status: OVER_QUERY_LIMIT"),
			want: true,
		},
		{
			name: "error message contains quota exceeded",
			err:  errors.New("quota exceeded"),
			want: true,
		},
		{
			name: "other error type",
			err: &GeocodingError{
				Type:    ErrorTypeRateLimit,
				Message: "rate limit",
			},
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	runErrorCheckTest(t, tests, IsQuotaExceededError)
}

func TestIsTimeoutError(t *testing.T) {
	tests := []errorCheckTestCase{
		{
			name: "timeout error type",
			err: &GeocodingError{
				Type:    ErrorTypeTimeout,
				Message: "timeout",
			},
			want: true,
		},
		{
			name: "error message contains timeout",
			err:  errors.New("request timeout after 10 seconds"),
			want: true,
		},
		{
			name: "error message contains deadline exceeded",
			err:  errors.New("context deadline exceeded"),
			want: true,
		},
		{
			name: "other error type",
			err: &GeocodingError{
				Type:    ErrorTypeNotFound,
				Message: "not found",
			},
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("some other error"),
			want: false,
		},
	}

	runErrorCheckTest(t, tests, IsTimeoutError)
}

func TestClassifyHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantType   ErrorType
	}{
		{
			name:       "429 too many requests",
			statusCode: 429,
			wantType:   ErrorTypeRateLimit,
		},
		{
			name:       "403 forbidden",
			statusCode: 403,
			wantType:   ErrorTypeQuotaExceeded,
		},
		{
			name:       "400 bad request",
			statusCode: 400,
			wantType:   ErrorTypeInvalidRequest,
		},
		{
			name:       "404 not found",
			statusCode: 404,
			wantType:   ErrorTypeNotFound,
		},
		{
			name:       "503 service unavailable",
			statusCode: 503,
			wantType:   ErrorTypeNetworkError,
		},
		{
			name:       "502 bad gateway",
			statusCode: 502,
			wantType:   ErrorTypeNetworkError,
		},
		{
			name:       "504 gateway timeout",
			statusCode: 504,
			wantType:   ErrorTypeNetworkError,
		},
		{
			name:       "500 internal server error",
			statusCode: 500,
			wantType:   ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyHTTPError(tt.statusCode, tt.body)
			if got.Type != tt.wantType {
				t.Errorf("ClassifyHTTPError() type = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}

func TestGeocodingErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	geoErr := &GeocodingError{
		Type:    ErrorTypeNotFound,
		Message: "location not found",
		Err:     innerErr,
	}

	if !errors.Is(geoErr, innerErr) {
		t.Error("errors.Is should find wrapped error")
	}

	if !errors.Is(geoErr.Unwrap(), innerErr) {
		t.Error("Unwrap should return inner error")
	}
}
