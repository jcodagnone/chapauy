// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var normalizeRegex = regexp.MustCompile(`[^\pL]`)

// normalize removes diacritics, non-letters and uppercases the string.
func normalize(s string) string {
	// Remove diacritics and non-letters
	s = normalizeRegex.ReplaceAllString(s, "")

	// Normalize to remove accents
	s, _, _ = transform.String(
		transform.Chain(
			norm.NFD,
			runes.Remove(runes.In(unicode.Mn)),
			norm.NFC,
		),
		s,
	)

	return strings.ToLower(s)
}

// NormalizeVehicleID removes any space and makes sure it is uppercase.
func NormalizeVehicleID(s string) string {
	if strings.IndexFunc(
		s,
		func(r rune) bool { return unicode.IsSpace(r) || r == '-' },
	) > -1 {
		// we don't want to have any space
		// nor we want `-' sometimes used in Montevideo they use it a
		// https://www.impo.com.uy/bases/notificaciones-cgm/1426-2022 - AAH-6052
		s = strings.Map(
			func(r rune) rune {
				if unicode.IsSpace(r) || r == '-' {
					return -1
				}

				return r
			},
			s,
		)
	}

	ret := strings.ToUpper(s)

	// Rio Negro usually mark foreign VehicleID with (E) - Extrajero
	ret = strings.TrimRight(ret, "(E)")

	return ret
}

// VehicleSpecialType represents a special plate combination and its category.
type VehicleSpecialType struct {
	Value    string
	Category string
}

// Vehicle category constants.
const (
	CatOfficial      = "Oficial"
	CatTaxi          = "Taxi"
	CatRemise        = "Remise"
	CatAmbulance     = "Ambulancia"
	CatJudicial      = "Poder Judicial"
	CatDiplomatic    = "Cuerpo Diplomático"
	CatConsular      = "Cuerpo Consular"
	CatSpecialMision = "Misión Especial"
)

var (
	uruguaySpecialTypes = []VehicleSpecialType{
		{Value: "DI", Category: "Discapacidad"},
		{Value: "ME", Category: "Medico"},
		{Value: "CC", Category: CatConsular},
		{Value: "CD", Category: CatDiplomatic},
		{Value: "OI", Category: CatSpecialMision},
		{Value: "AT", Category: "Admisión Temporaria - Administrativos y Técnicos"},
		{Value: "NG", Category: CatSpecialMision},
		{Value: "OF", Category: CatOfficial},
		{Value: "PJ", Category: CatJudicial},
		{Value: "AM", Category: CatAmbulance},
		{Value: "RE", Category: CatRemise},
		{Value: "TX", Category: CatTaxi},
	}

	chileSpecialTypes = []VehicleSpecialType{
		{Value: "CC", Category: CatConsular},
		{Value: "CD", Category: CatDiplomatic},
		{Value: "EE", Category: CatSpecialMision},
	}
)

// Vehicle format constants.
const (
	FormatMercosur = "Mercosur"
)

// Uruguay department constants.
const (
	DeptMontevideo  = "S"
	DeptCanelones   = "A"
	DeptMaldonado   = "B"
	DeptRocha       = "C"
	DeptTreintaTres = "D"
	DeptCerroLargo  = "E"
	DeptRivera      = "F"
	DeptArtigas     = "G"
	DeptSalto       = "H"
	DeptPaysandu    = "I"
	DeptRioNegro    = "J"
	DeptSoriano     = "K"
	DeptColonia     = "L"
	DeptSanJose     = "M"
	DeptFlores      = "N"
	DeptFlorida     = "O"
	DeptLavalleja   = "P"
	DeptDurazno     = "Q"
	DeptTacuarembo  = "R"
)

// Map first letter to department in Uruguay.
var uruguayDepartments = map[string]bool{
	DeptMontevideo:  true,
	DeptCanelones:   true,
	DeptMaldonado:   true,
	DeptRocha:       true,
	DeptTreintaTres: true,
	DeptCerroLargo:  true,
	DeptRivera:      true,
	DeptArtigas:     true,
	DeptSalto:       true,
	DeptPaysandu:    true,
	DeptRioNegro:    true,
	DeptSoriano:     true,
	DeptColonia:     true,
	DeptSanJose:     true,
	DeptFlores:      true,
	DeptFlorida:     true,
	DeptLavalleja:   true,
	DeptDurazno:     true,
	DeptTacuarembo:  true,
}

// VehicleInfo contains the information extracted from a vehicle's license plate.
type VehicleInfo struct {
	Country        string `json:"country,omitempty"`      // ISO country code
	AdmDivision    string `json:"adm_division,omitempty"` // Department/Province if available
	VehicleType    string `json:"vehicle_type,omitempty"` // Vehicle type (Car, Motorcycle, etc.)
	Category       string `json:"category,omitempty"`     // Official, Private, etc.
	MercosurFormat bool   `json:"mercosur_format"`        // License plate format (Mercosur)
}

// PlatePattern defines a license plate pattern for a specific type/category.
type PlatePattern struct {
	Regex          *regexp.Regexp
	MercosurFormat bool
	Category       string
	VehicleType    string
	Department     string
}

// Vehicle type constants.
const (
	TypeAuto       = "Auto"
	TypeMoto       = "Moto"
	TypeAutoOrMoto = ""
)

// Country code constants.
const (
	ISOUruguay   = "UY"
	ISOArgentina = "AR"
	ISOBrasil    = "BR"
	ISOParaguay  = "PY"
	ISOChile     = "CL"
)

var (
	// Motorcycle prefixes in Mercosur format in Uruguay.
	uruguayMercosurMotoPrefixes = map[string]struct{}{
		// Canelones
		"AV": {},
		// BF es Garzón, pero desde la M se usa para identificar motos
		"BFM": {},
		"BFN": {},
		"BFO": {},
		"BFP": {},
		"BFQ": {}, // no visto hasta ahora
		// Colonia
		"LS": {},
		// Minas
		"PAV": {},
	}

	countryPatterns = []struct {
		ISO      string
		Patterns []PlatePattern
	}{
		{
			ISO: ISOUruguay,
			Patterns: []PlatePattern{
				{
					// Current Mercosur format (3 letters + 4 numbers)
					Regex:          regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`),
					MercosurFormat: true,
					VehicleType:    TypeAutoOrMoto,
				},
				{
					Regex:       regexp.MustCompile(`^[A-S][A-Z]{2}[0-9]{3}$`),
					VehicleType: TypeMoto,
				},
				{
					// Old Montevideo format (6 digits)
					Regex:       regexp.MustCompile(`^[0-9]{6}$`),
					VehicleType: TypeAutoOrMoto,
					Department:  DeptMontevideo,
				},
				{
					// Old format from interior (3 letters + 3 numbers or variations)
					Regex:       regexp.MustCompile(`^[A-Z][0-9]{3,6}$`),
					VehicleType: TypeAutoOrMoto,
				},
			},
		},
		{
			ISO: ISOArgentina,
			Patterns: []PlatePattern{
				{
					// AA000AA - Mercosur - Cars
					Regex:          regexp.MustCompile(`^[A-Z]{2}[0-9]{3}[A-Z]{2}$`),
					MercosurFormat: true,
					VehicleType:    TypeAuto,
				},
				{
					// A000AAA - Mercosur motorcycles
					Regex:          regexp.MustCompile(`^[A-Z][0-9]{3}[A-Z]{3}$`),
					MercosurFormat: true,
					VehicleType:    TypeMoto,
				},
				{
					// AAA000 - 1995-2016
					Regex:       regexp.MustCompile(`^[A-Z]{3}[0-9]{3}$`),
					VehicleType: TypeAuto,
				},
			},
		},
		{
			ISO: ISOBrasil,
			Patterns: []PlatePattern{
				{
					// AAA0A00 - Mercosur
					Regex:          regexp.MustCompile(`^[A-Z]{3}[0-9][A-Z][0-9]{2}$`),
					MercosurFormat: true,
					VehicleType:    TypeAuto,
				},
				{
					// AAA0000 - Old
					Regex:       regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`),
					VehicleType: TypeAuto,
				},
			},
		},
		{
			ISO: ISOParaguay,
			Patterns: []PlatePattern{
				{
					// AAAA000 - Mercosur
					Regex:          regexp.MustCompile(`^[A-Z]{4}[0-9]{3}$`),
					MercosurFormat: true,
					VehicleType:    TypeAuto,
				},
				{
					// 000AAAA - Mercosur
					Regex:          regexp.MustCompile(`^[0-9]{3}[A-Z]{4}$`),
					MercosurFormat: true,
					VehicleType:    TypeMoto,
				},
				{
					// 000-000 - Old format
					Regex:       regexp.MustCompile(`^[0-9]{6}$`),
					VehicleType: TypeAuto,
				},
			},
		},
		{
			ISO: ISOChile,
			Patterns: []PlatePattern{
				{
					// BBBB00 or BBBB00-0
					Regex:       regexp.MustCompile(`^[A-Z]{2,4}[0-9]{2,4}$`),
					VehicleType: TypeAuto,
				},
			},
		},
	}
)

func analyzeCountry(
	plate string,
	countryISO string,
	patterns []PlatePattern,
) (*VehicleInfo, bool) {
	info := &VehicleInfo{
		Country: countryISO,
	}

	// Check country patterns
	for _, pattern := range patterns {
		if pattern.Regex.MatchString(plate) {
			info.MercosurFormat = pattern.MercosurFormat
			info.Category = pattern.Category
			info.VehicleType = pattern.VehicleType
			info.AdmDivision = pattern.Department

			switch countryISO {
			case ISOUruguay:
				firstLetter := string(plate[0])
				if _, exists := uruguayDepartments[firstLetter]; exists {
					info.AdmDivision = firstLetter
				} else {
					info.AdmDivision = DeptMontevideo
				}

				if info.MercosurFormat {
					// The first two letters of Mercosur plates can identify the
					// vehicle type.
					_, isMoto2 := uruguayMercosurMotoPrefixes[plate[0:2]]
					_, isMoto3 := uruguayMercosurMotoPrefixes[plate[0:3]]

					if isMoto2 || isMoto3 {
						info.VehicleType = TypeMoto
					}

					for _, suffixInfo := range uruguaySpecialTypes {
						if plate[1:3] == suffixInfo.Value {
							info.Category = suffixInfo.Category

							return info, true
						}
					}
				}

			case ISOChile:
				// Check special prefixes for Chile
				for _, prefixInfo := range chileSpecialTypes {
					if strings.HasPrefix(plate, prefixInfo.Value) {
						info.Category = prefixInfo.Category

						return info, true
					}
				}

				// Check if it's a motorcycle by plate format
				if len(plate) >= 6 && regexp.MustCompile(`^[A-Z]{4}[0-9]{2}$`).MatchString(plate) {
					info.VehicleType = TypeMoto
				}
			}

			return info, true
		}
	}

	return &VehicleInfo{}, false
}

// normalizeCountryName normalizes a country name to its ISO code.
func normalizeCountryName(name string) (string, error) {
	switch normalize(name) {
	case "argentina":
		return ISOArgentina, nil
	case "uruguay":
		return ISOUruguay, nil
	case "brasil":
		return ISOBrasil, nil
	case "paraguay":
		return ISOParaguay, nil
	case "chile":
		return ISOChile, nil
	case "otro":
		return "", nil
	}

	return "", fmt.Errorf("unknown country: %q", name)
}

// AnalyzeVehicleID infers information from a license plate. On error returns blank + error.
func AnalyzeVehicleID(plate string, countryHint string) (*VehicleInfo, error) {
	plate = NormalizeVehicleID(plate)

	for _, countryCheck := range countryPatterns {
		if countryHint != "" && countryCheck.ISO != countryHint {
			continue
		}

		if info, matched := analyzeCountry(
			plate,
			countryCheck.ISO,
			countryCheck.Patterns,
		); matched {
			return info, nil
		}

		if countryHint != "" && countryCheck.ISO == countryHint {
			return &VehicleInfo{
				Country: countryHint,
			}, nil
		}
	}

	return &VehicleInfo{}, errors.New("no info available")
}
