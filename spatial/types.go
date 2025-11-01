// Copyright 2025 The ChapaUY Authors
//
// SPDX-License-Identifier: Apache-2.0
package spatial

import (
	"database/sql/driver"
	"fmt"
	"math"
)

const earthRadius = 6371e3 // meters

// Point represents a geographical point with latitude and longitude.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// String returns a string representation of the Point.
func (p Point) String() string {
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat)
}

// Value implements the driver.Valuer interface for database serialization.
func (p Point) Value() (driver.Value, error) {
	return p.String(), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		p.Lat, p.Lng = 0, 0

		return nil
	}

	switch v := value.(type) {
	case []byte:
		// The format from DuckDB is "POINT (lng lat)"
		_, err := fmt.Sscanf(string(v), "POINT (%f %f)", &p.Lng, &p.Lat)

		return err
	case map[string]interface{}:
		x, okX := v["x"].(float64)
		y, okY := v["y"].(float64)

		if !okX || !okY {
			return fmt.Errorf("spatial: invalid map for point: expected 'x' and 'y' float64 fields, got %+v", v)
		}

		p.Lng = x
		p.Lat = y

		return nil
	default:
		return fmt.Errorf("spatial: unsupported type for Point scan: %T", value)
	}
}

// HaversineDistance calculates the distance between two points on Earth in meters.
func (p *Point) HaversineDistance(other *Point) float64 {
	lat1 := p.Lat * math.Pi / 180
	lat2 := other.Lat * math.Pi / 180
	dLat := (other.Lat - p.Lat) * math.Pi / 180
	dLng := (other.Lng - p.Lng) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
