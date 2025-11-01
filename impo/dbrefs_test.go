// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"errors"
	"strings"
	"testing"
)

func TestFind(t *testing.T) {
	// table-driven test cases
	tests := []struct {
		name         string
		query        string
		expectedName string
		expectErr    string
	}{
		{
			name:         "NumericMatch",
			query:        "65",
			expectedName: "Caminera",
		},
		{
			name:         "StringExactMatch",
			query:        "Montevideo",
			expectedName: "Montevideo",
		},
		{
			name:         "CaseSensitiveMatch",
			query:        "monteVIdEO",
			expectedName: "Montevideo",
		},
		{
			name:         "CasePrefixMatch",
			query:        "MONTE",
			expectedName: "Montevideo",
		},
		{
			name:      "NoMatch",
			query:     "xxx",
			expectErr: "not found",
		},
		{
			name:      "MultipleMatches",
			query:     "C", // Colonia, Canelones
			expectErr: "multiple matches",
		},
	}

	// We can use any instance to call the method because Find does not use the receiver content.
	// Using DbCaminera to obtain a pointer receiver.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Find(tc.query)
			// Check error conditions
			if tc.expectErr != "" {
				// Even if got error, the returned reference might be non-nil,
				// but we care that the error indicates multiple matches.
				switch {
				case got != nil:
					t.Errorf("Find(%q) expected nil db", tc.query)
				case err == nil:
					t.Errorf("Find(%q) expected error but got none", tc.query)
				case !strings.Contains(err.Error(), tc.expectErr):
					t.Errorf("Find(%q) expecting %v but got : %v", tc.query, tc.expectErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Find(%q) unexpected error: %v", tc.query, err)
				}

				if tc.expectedName == "" {
					if got != nil {
						t.Errorf("Find(%q) expected nil but got %+v", tc.query, got)
					}
				} else {
					if got == nil {
						t.Errorf("Find(%q) expected %q but got nil", tc.query, tc.expectedName)
					} else if got.Name != tc.expectedName {
						t.Errorf("Find(%q) expected database name %q, got %q", tc.query, tc.expectedName, got.Name)
					}
				}
			}
		})
	}
}

func TestEach_Ok(t *testing.T) {
	var found []string

	err := Each(func(db DbReference) error {
		found = append(found, db.Name)

		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if expected, got := "Caminera", found[0]; expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	} else if expected, got := "Vialidad", found[len(found)-1]; expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEach_Err(t *testing.T) {
	var found []string

	i := 0

	err := Each(func(db DbReference) (err error) {
		if i >= 2 {
			err = errors.New("fail")
		} else {
			found = append(found, db.Name)
		}

		i++

		return err
	})
	if err == nil {
		t.Error("expecting an  error")
	} else if expected, got := "Caminera", found[0]; expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	} else if expected, got := "Canelones", found[len(found)-1]; expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEachURI2file(t *testing.T) {
	tests := []struct {
		db       string
		id       string
		expected []string
	}{
		{
			db:       "Caminera",
			id:       "https://www.impo.com.uy/bases/notificaciones-policia-caminera/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
		{
			db:       "Caminera",
			id:       "https://www.impo.com.uy/bases/resoluciones-policia-caminera/3701-2023",
			expected: []string{"resoluciones", "2023", "3701"},
		},
		{
			db:       "Colonia",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-colonia/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
		{
			db:       "Canelones",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-canelones/99-2023",
			expected: []string{"notificaciones", "2023", "99"},
		},
		{
			db:       "Lavalleja",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/1-2021",
			expected: []string{"notificaciones", "2021", "1"},
		},
		{
			db:       "Lavalleja",
			id:       "https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/1-2021",
			expected: []string{"resoluciones", "2021", "1"},
		},
		{
			db:       "Maldonado",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-maldonado/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
		{
			db:       "Maldonado",
			id:       "https://www.impo.com.uy/bases/resoluciones-transito-maldonado/31-2023_A",
			expected: []string{"resoluciones", "2023", "31_A"},
		},
		{
			db:       "Maldonado",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/170-2025",
			expected: []string{"notificaciones", "2025", "170"},
		},
		{
			db:       "Montevideo",
			id:       "https://www.impo.com.uy/bases/notificaciones-cgm/999-2022",
			expected: []string{"notificaciones", "2022", "999"},
		},
		{
			db:       "Montevideo",
			id:       "https://www.impo.com.uy/bases/resoluciones-cgm/999-2022",
			expected: []string{"resoluciones", "2022", "999"},
		},
		{
			db:       "Paysandu",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-paysandu/1-2022",
			expected: []string{"notificaciones", "2022", "1"},
		},
		{
			db:       "Paysandu",
			id:       "https://www.impo.com.uy/bases/resoluciones-transito-paysandu/1-2022",
			expected: []string{"resoluciones", "2022", "1"},
		},
		{
			db:       "Rio Negro",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-rionegro/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
		{
			db:       "Rio Negro",
			id:       "https://www.impo.com.uy/bases/resoluciones-transito-rionegro/1-2023",
			expected: []string{"resoluciones", "2023", "1"},
		},
		{
			db:       "Soriano",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-soriano/1-2024",
			expected: []string{"notificaciones", "2024", "1"},
		},
		{
			db:       "Soriano",
			id:       "https://www.impo.com.uy/bases/resoluciones-transito-soriano/1-2024",
			expected: []string{"resoluciones", "2024", "1"},
		},
		{
			db:       "Tacuarembó",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-tacuarembo/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
		{
			db:       "Treinta y Tres",
			id:       "https://www.impo.com.uy/bases/notificaciones-transito-treintaytres/1-2023",
			expected: []string{"notificaciones", "2023", "1"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			db, err := Find(tc.db)
			if err != nil {
				t.Error(err)

				return
			}

			var file []string

			// Try each extraction function until one succeeds
			for _, extractFunc := range db.id2file {
				file, err = extractFunc(tc.id)
				if err == nil {
					break
				}
			}

			if err != nil {
				t.Error(err)

				return
			}

			if strings.Join(tc.expected, "/") != strings.Join(file, "/") {
				t.Fatalf("expected %q, but got %q", tc.expected, file)

				return
			}
		})
	}
}
