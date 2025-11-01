// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

import (
	"errors"
	"fmt"
	"strings"
)

// validMethods contiene los métodos de geocodificación permitidos.
var validMethods = map[string]bool{
	"radares_rutas":     true,
	"google_maps":       true,
	"manual":            true,
	"manual_click":      true,
	"manual_adjustment": true,
	"manual_input":      true,
}

// validConfidence contiene los niveles de confianza permitidos.
var validConfidence = map[string]bool{
	"high":   true,
	"medium": true,
	"low":    true,
	"none":   true,
}

// validateCoordinates verifica que las coordenadas sean válidas.
func validateCoordinates(lat, lon float64) error {
	// Límites globales
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitud debe estar entre -90 y 90 (recibido: %f)", lat)
	}

	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitud debe estar entre -180 y 180 (recibido: %f)", lon)
	}

	// Límites razonables para Uruguay (con margen)
	// Uruguay: aproximadamente 30°S a 35°S, 53°W a 58°W
	// Usamos un margen de ~1 grado para errores de precisión
	const (
		uruguayMinLat = -36.0
		uruguayMaxLat = -29.0
		uruguayMinLon = -59.0
		uruguayMaxLon = -52.0
	)

	if lat < uruguayMinLat || lat > uruguayMaxLat {
		return fmt.Errorf("latitud fuera de los límites de Uruguay (%f a %f): %f", uruguayMinLat, uruguayMaxLat, lat)
	}

	if lon < uruguayMinLon || lon > uruguayMaxLon {
		return fmt.Errorf("longitud fuera de los límites de Uruguay (%f a %f): %f", uruguayMinLon, uruguayMaxLon, lon)
	}

	return nil
}

// validateJudgment verifica que un LocationJudgment tenga datos válidos.
func validateJudgment(j *Location) error {
	if j == nil {
		return errors.New("judgment no puede ser nil")
	}

	// Validar ubicación
	if strings.TrimSpace(j.Location) == "" {
		return errors.New("location no puede estar vacío")
	}

	if len(j.Location) > 500 {
		return errors.New("location demasiado largo (máximo 500 caracteres)")
	}

	// Validar coordenadas si están presentes
	if j.Point != nil {
		if err := validateCoordinates(j.Point.Lat, j.Point.Lng); err != nil {
			return fmt.Errorf("coordenadas inválidas: %w", err)
		}
	}

	// Validar método de geocodificación
	if j.GeocodingMethod != "" && !validMethods[j.GeocodingMethod] {
		return fmt.Errorf("método de geocodificación inválido: %s", j.GeocodingMethod)
	}

	// Validar nivel de confianza
	if j.Confidence != "" && !validConfidence[j.Confidence] {
		return fmt.Errorf("nivel de confianza inválido: %s", j.Confidence)
	}

	// Validar notas
	if len(j.Notes) > 1000 {
		return errors.New("notes demasiado largo (máximo 1000 caracteres)")
	}

	return nil
}

// sanitizeLocation limpia y normaliza una cadena de ubicación.
func sanitizeLocation(loc string) string {
	// Eliminar espacios al inicio y final
	loc = strings.TrimSpace(loc)

	// Limitar longitud
	if len(loc) > 500 {
		loc = loc[:500]
	}

	return loc
}
