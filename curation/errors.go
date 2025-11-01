// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// GeocodingError representa errores específicos de geocodificación.
type GeocodingError struct {
	Type    ErrorType
	Message string
	Err     error
}

// ErrorType define tipos de errores de geocodificación.
type ErrorType int

const (
	// ErrorTypeUnknown error desconocido.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeRateLimit límite de tasa alcanzado.
	ErrorTypeRateLimit
	// ErrorTypeQuotaExceeded cuota excedida.
	ErrorTypeQuotaExceeded
	// ErrorTypeTimeout timeout de conexión.
	ErrorTypeTimeout
	// ErrorTypeNotFound ubicación no encontrada.
	ErrorTypeNotFound
	// ErrorTypeInvalidRequest request inválido.
	ErrorTypeInvalidRequest
	// ErrorTypeNetworkError error de red.
	ErrorTypeNetworkError
)

func (e *GeocodingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *GeocodingError) Unwrap() error {
	return e.Err
}

// IsRateLimitError verifica si el error es por límite de tasa.
func IsRateLimitError(err error) bool {
	var geoErr *GeocodingError
	if errors.As(err, &geoErr) {
		return geoErr.Type == ErrorTypeRateLimit
	}

	// Detectar por mensaje de error común
	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "429")
}

// IsQuotaExceededError verifica si el error es por cuota excedida.
func IsQuotaExceededError(err error) bool {
	var geoErr *GeocodingError
	if errors.As(err, &geoErr) {
		return geoErr.Type == ErrorTypeQuotaExceeded
	}

	// Detectar por mensaje de error común (Google Maps)
	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "over_query_limit") ||
		strings.Contains(errStr, "quota exceeded")
}

// IsTimeoutError verifica si el error es por timeout.
func IsTimeoutError(err error) bool {
	var geoErr *GeocodingError
	if errors.As(err, &geoErr) {
		return geoErr.Type == ErrorTypeTimeout
	}

	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded")
}

// ClassifyHTTPError clasifica un error HTTP en un tipo de error de geocodificación.
func ClassifyHTTPError(statusCode int, _ string) *GeocodingError {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return &GeocodingError{
			Type:    ErrorTypeRateLimit,
			Message: "límite de tasa alcanzado",
		}
	case http.StatusForbidden: // 403
		return &GeocodingError{
			Type:    ErrorTypeQuotaExceeded,
			Message: "cuota excedida o acceso denegado",
		}
	case http.StatusBadRequest: // 400
		return &GeocodingError{
			Type:    ErrorTypeInvalidRequest,
			Message: "request inválido",
		}
	case http.StatusNotFound: // 404
		return &GeocodingError{
			Type:    ErrorTypeNotFound,
			Message: "ubicación no encontrada",
		}
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return &GeocodingError{
			Type:    ErrorTypeNetworkError,
			Message: fmt.Sprintf("servicio no disponible (código %d)", statusCode),
		}
	default:
		return &GeocodingError{
			Type:    ErrorTypeUnknown,
			Message: fmt.Sprintf("error HTTP %d", statusCode),
		}
	}
}
