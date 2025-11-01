// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"AAA 0000", "AAA0000"},
		{"AAA-000", "AAA000"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if s := NormalizeVehicleID(tt.id); s != tt.want {
				t.Errorf("Format = %v, want %v", s, tt.want)
			}
		})
	}
}

func TestAnalyzeMatricula(t *testing.T) {
	tests := []struct {
		name            string
		plate           string
		wantCountryISO  string
		wantVehicleType string
		wantCategory    string
		wantDepartment  string
		wantMercosurFmt bool
		wantError       bool
	}{
		{
			name:            "Uruguay - Maldonado",
			plate:           "BEC0000",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantCategory:    "",
			wantDepartment:  DeptMaldonado,
			wantMercosurFmt: true,
		},
		{
			name:            "Uruguay - Oficial de Montevideo",
			plate:           "SOF0000",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantCategory:    CatOfficial,
			wantDepartment:  DeptMontevideo,
			wantMercosurFmt: true,
		},
		{
			name:            "Uruguay - Taxi de Colonia",
			plate:           "LTX0000",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantCategory:    CatTaxi,
			wantDepartment:  DeptColonia,
			wantMercosurFmt: true,
		},
		{
			name:            "Uruguay - Formato antiguo de Montevideo",
			plate:           "123456",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantDepartment:  DeptMontevideo,
		},
		{
			name:            "Uruguay - Formato antiguo del interior",
			plate:           "B000000",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantDepartment:  DeptMaldonado,
		},
		{
			name:            "Argentina - Formato Mercosur",
			plate:           "AA000AA",
			wantCountryISO:  ISOArgentina,
			wantVehicleType: TypeAuto,
			wantMercosurFmt: true,
		},
		{
			name:            "Argentina - Moto formato Mercosur",
			plate:           "A000AAA",
			wantCountryISO:  ISOArgentina,
			wantVehicleType: TypeMoto,
			wantMercosurFmt: true,
		},
		{
			name:            "Uruguay - Moto LLLNNN (AAA000) - old format",
			plate:           "AAA000",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeMoto,
			wantDepartment:  DeptCanelones,
		},
		{
			name:            "Uruguay - Moto LLLNNN (SNW235)  - old format",
			plate:           "SNW235",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeMoto,
			wantDepartment:  DeptMontevideo,
		},
		{
			name:            "Uruguay - Moto LLLNNN (BFM643)",
			plate:           "BFM643",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeMoto,
			wantDepartment:  DeptMaldonado,
		},
		{
			name:            "Uruguay - Moto Mercosur (SBF1234)",
			plate:           "SBF1234",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeAutoOrMoto,
			wantDepartment:  DeptMontevideo,
			wantMercosurFmt: true,
		},
		{
			name:            "Uruguay - Moto Mercosur (BFM2698)",
			plate:           "BFM2698",
			wantCountryISO:  ISOUruguay,
			wantVehicleType: TypeMoto,
			wantDepartment:  DeptMaldonado,
			wantMercosurFmt: true,
		},
		{
			name:            "Brasil - Formato Mercosur",
			plate:           "AAA0A00",
			wantCountryISO:  ISOBrasil,
			wantVehicleType: TypeAuto,
			wantMercosurFmt: true,
		},
		{
			name:            "Paraguay - Formato Mercosur",
			plate:           "AAAA000",
			wantCountryISO:  ISOParaguay,
			wantVehicleType: TypeAuto,
			wantMercosurFmt: true,
		},
		{
			name:            "Paraguay - Formato Mercosur - MOTO",
			plate:           "000AAAA",
			wantCountryISO:  ISOParaguay,
			wantVehicleType: TypeMoto,
			wantMercosurFmt: true,
		},
		{
			name:            "Chile Auto",
			plate:           "AA0000",
			wantCountryISO:  ISOChile,
			wantVehicleType: TypeAuto,
			wantMercosurFmt: false,
		},
		{
			name:      "Matrícula inválida",
			plate:     "!@#$%",
			wantError: true,
		},
		{
			name:      "Matrícula vacía",
			plate:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := AnalyzeVehicleID(tt.plate, "")

			// Check if error when expected
			if (err != nil) != tt.wantError {
				t.Errorf("AnalyzeVehicleID() error = %v, wantError %v", err, tt.wantError)

				return
			}

			// If expecting error, don't check more fields
			if tt.wantError {
				return
			}

			// Check all expected fields
			if info.Country != tt.wantCountryISO {
				t.Errorf("CountryISO = %v, want %v", info.Country, tt.wantCountryISO)
			}

			if info.VehicleType != tt.wantVehicleType {
				t.Errorf("VehicleType = %v, want %v", info.VehicleType, tt.wantVehicleType)
			}

			if info.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", info.Category, tt.wantCategory)
			}

			if info.AdmDivision != tt.wantDepartment {
				t.Errorf("Department = %v, want %v", info.AdmDivision, tt.wantDepartment)
			}

			if info.MercosurFormat != tt.wantMercosurFmt {
				t.Errorf("Format = %v, want %v", info.MercosurFormat, tt.wantMercosurFmt)
			}
		})
	}
}

// TestSpecificCases checks specific examples mentioned.
func TestSpecificCases(t *testing.T) {
	// BEC0000 - Uruguay, Punta del Este
	if info, err := AnalyzeVehicleID("BEC0000", ""); err != nil {
		t.Errorf("analyzing BEC0000: %v", err)
	} else if info.Country != ISOUruguay || info.Category != "" ||
		info.AdmDivision != DeptMaldonado {
		t.Errorf("BEC0000 expected Uruguay, Maldonado but got: %+v", info)
	}

	// SOF0000 - Uruguay, Oficial de Montevideo
	if info, err := AnalyzeVehicleID("SOF0000", ""); err != nil {
		t.Errorf("analyzing SOF0000: %v", err)
	} else if info.Country != ISOUruguay || info.Category != CatOfficial ||
		info.AdmDivision != DeptMontevideo {
		t.Errorf("SOF0000 expected Uruguay, Oficial, Montevideo but got: %+v", info)
	}

	// LTX0000 - Uruguay, Taxi de Colonia
	if info, err := AnalyzeVehicleID("LTX0000", ""); err != nil {
		t.Errorf("analyzing LTX0000: %v", err)
	} else if info.Country != ISOUruguay || info.Category != CatTaxi ||
		info.AdmDivision != DeptColonia {
		t.Errorf("LTX0000 expected Uruguay, Taxi, Colonia, but got: %+v", info)
	}
}

func TestNormalizeCountryName(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"Uruguay", ISOUruguay, false},
		{" URUGUAY ", ISOUruguay, false},
		{"Argentina", ISOArgentina, false},
		{"Brasil", ISOBrasil, false},
		{"Paraguay", ISOParaguay, false},
		{"Chile", ISOChile, false},
		{"OTRO", "", false},
		{"Unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := normalizeCountryName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeCountryName() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("normalizeCountryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzeVehicleID_WithHint(t *testing.T) {
	// Case where hint disambiguates
	// Assume a plate that could be valid in multiple countries if we had overlapping patterns
	// For now, let's test that providing a hint works for a valid plate of that country
	// Uruguay plate with Uruguay hint
	info, err := AnalyzeVehicleID("ABC1234", ISOUruguay)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info.Country != ISOUruguay {
		t.Errorf("expected UY, got %v", info.Country)
	}

	// Uruguay plate with Argentina hint - should fail to match pattern (assuming ABC1234 is not valid AR)
	// Actually ABC1234 matches "Old format from interior" of Uruguay: ^[A-Z][0-9]{3,6}$ -> No, ABC1234 is 3 letters 4 numbers
	// Uruguay Mercosur: ^[A-Z]{3}[0-9]{4}$ -> Matches ABC1234

	// Argentina Mercosur: ^[A-Z]{2}[0-9]{3}[A-Z]{2}$ -> No
	// Argentina Old: ^[A-Z]{3}[0-9]{3}$ -> No

	// So ABC1234 is NOT valid in Argentina.
	// If we hint Argentina, it should return the hint but maybe empty other fields?
	// Our implementation returns:
	// return &VehicleInfo{Country: countryHint}, nil

	info, err = AnalyzeVehicleID("ABC1234", ISOArgentina)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if info.Country != ISOArgentina {
		t.Errorf("expected AR (from hint), got %v", info.Country)
	}
	// It shouldn't have matched any pattern, so other fields should be empty/default
	if info.VehicleType != "" {
		t.Errorf("expected empty vehicle type, got %v", info.VehicleType)
	}
}
