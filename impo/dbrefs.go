// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"errors"
	"fmt"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	errMultipleMatches  = errors.New("multiple matches")
	errDatabaseNotFound = errors.New("database not found")
)

// DbReference represents a reference to an IMPO database. See:
// https://www.impo.com.uy/directorio-bases-institucionales/
type DbReference struct {
	Name     string                           // Name of the database
	ID       int                              // ID of the database
	TodosID  int                              // ID of the document type to search
	SeedURL  string                           // Initial URL from where we get the anonymous credentials
	QueryURL string                           // URL used for querying the database
	BaseURL  string                           // Base URL for each documents, it isn't always the same domain as the query
	Issuers  []string                         // List of issuing organizations
	id2file  []func(string) ([]string, error) // Functions that transform the URL to a filesystem path for storage
}

// Validate checks if the DbReference has all required fields.
// Returns an error if any required field is missing.
func (d *DbReference) Validate() error {
	if d.Name == "" {
		return errors.New("database reference: name must not be empty")
	}

	if d.SeedURL == "" {
		return fmt.Errorf("database reference %q: seed URL must not be empty", d.Name)
	}

	return nil
}

// GetDBName returns the name of the database with the given ID.
func GetDBName(id int) (string, error) {
	for _, db := range databases {
		if db.ID == id {
			return db.Name, nil
		}
	}

	return "", fmt.Errorf("database with ID %d not found", id)
}

// All available databases.
var databases = func() []DbReference {
	ret := []DbReference{
		{
			ID:       65,
			Name:     "Caminera",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multascaminera",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=65",
			BaseURL:  "https://impo.com.uy/",
			TodosID:  799,
			Issuers: []string{
				"Policía Caminera",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-policia-caminera/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       40,
			Name:     "Canelones",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multascanelones",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=40",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  709,
			Issuers: []string{
				"Dirección General de Tránsito y Transporte Intendencia de Canelones",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-canelones/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       48,
			Name:     "Colonia",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multascolonia",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=48",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  876,
			Issuers: []string{
				"Dirección de Tránsito y Transporte Intendencia de Colonia",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-colonia/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       26,
			Name:     "Lavalleja",
			SeedURL:  "https://impo.com.uy/base-institucional/multaslavalleja",
			QueryURL: "https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=26",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  600,
			Issuers: []string{
				"Dirección de Tránsito Intendencia de Lavalleja",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-lavalleja/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       45,
			Name:     "Maldonado",
			SeedURL:  "https://impo.com.uy/base-institucional/multasmaldonado",
			QueryURL: "https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=45",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  802,
			Issuers: []string{
				"Dirección General de Tránsito y Transporte Intendencia de Maldonado",
				"Departamento de Movilidad Intendencia de Maldonado",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-maldonado/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-movilidad-maldonado/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       6,
			Name:     "Montevideo",
			SeedURL:  "https://www.impo.com.uy/base-institucional/cgm",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=6",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  383,
			Issuers: []string{
				"Centro de Gestión de Movilidad",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-cgm/([\dA-Za-z]+)-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       43,
			Name:     "Paysandu",
			SeedURL:  "https://impo.com.uy/base-institucional/multaspaysandu",
			QueryURL: "https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=43",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  777,
			Issuers: []string{
				"Dirección de Tránsito Intendencia de Paysandú",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-paysandu/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       55,
			Name:     "Rio Negro",
			SeedURL:  "https://impo.com.uy/base-institucional/multasrionegro",
			QueryURL: "https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=55",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  815,
			Issuers: []string{
				"Dirección de Tránsito Intendencia de Río Negro",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-rionegro/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       49,
			Name:     "Soriano",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multassoriano",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=49",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  879,
			Issuers: []string{
				"Departamento de Tránsito y Transporte Intendencia de Soriano",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-soriano/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       56,
			Name:     "Tacuarembó",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multastacuarembo",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=56",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  891,
			Issuers: []string{
				"Dirección General de Tránsito Intendencia de Tacuarembó",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(notificaciones)-transito-tacuarembo/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       52,
			Name:     "Treinta y Tres",
			SeedURL:  "https://impo.com.uy/base-institucional/multastreintaytres",
			QueryURL: "https://impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=52",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  818,
			Issuers: []string{
				"Dirección de Tránsito Intendencia de Treinta y Tres",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(notificaciones)-transito-treintaytres/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
		{
			ID:       68,
			Name:     "Vialidad",
			SeedURL:  "https://www.impo.com.uy/base-institucional/multasmtop",
			QueryURL: "https://www.impo.com.uy/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=68",
			BaseURL:  "https://www.impo.com.uy/",
			TodosID:  867,
			Issuers: []string{
				"Tránsito MTOP",
			},
			id2file: []func(string) ([]string, error){
				makeID2PathFunc(
					regexp.MustCompile(`^/bases/(resoluciones|notificaciones)-transito-mtop/([\dA-Za-z]+)\-(\d+)(?:_([A-Z]))?$`),
					typeNumberYearOptional,
				),
			},
		},
	}

	// Validate and prepare databases
	for i := range ret {
		if err := ret[i].Validate(); err != nil {
			panic(err)
		}
		// Convert issuers to lowercase to ease later matching
		for j := range ret[i].Issuers {
			ret[i].Issuers[j] = strings.ToLower(ret[i].Issuers[j])
		}
	}

	return ret
}()

// Extracts type, year, and number from regexp matches.
// It handles the optional suffix in document numbers.
func typeNumberYearOptional(matches []string) []string {
	// Document number
	number := matches[2]
	// Add optional suffix (A, B, etc.)
	if matches[4] != "" {
		number = number + "_" + matches[4]
	}

	return []string{
		matches[1], // type (resoluciones|notificaciones)
		matches[3], // year
		number,     // number
	}
}

// Creates a function that transforms a URL path into
// filesystem path components using the provided regex and transformer.
func makeID2PathFunc(
	re *regexp.Regexp,
	transformer func([]string) []string,
) func(string) ([]string, error) {
	return func(id string) ([]string, error) {
		url, err := neturl.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("parsing id as URL %q: %w", id, err)
		}

		matches := re.FindStringSubmatch(url.Path)
		if matches == nil {
			return nil, fmt.Errorf("failed to match path %q with pattern %s", url.Path, re)
		}

		return transformer(matches), nil
	}
}

// Find locates a database by its ID or name.
// If q represents a number, it searches by ID; otherwise, it searches by name.
// Returns an error if no match or multiple matches are found.
func Find(q string) (*DbReference, error) {
	if q == "" {
		return nil, errors.New("empty search query")
	}

	var predicate func(d *DbReference) bool
	if n, err := strconv.Atoi(q); err == nil {
		predicate = func(db *DbReference) bool {
			return n == db.ID
		}
	} else {
		predicate = func(db *DbReference) bool {
			// Case insensitive prefix match
			return len(db.Name) >= len(q) &&
				strings.EqualFold(db.Name[:len(q)], q)
		}
	}

	var found *DbReference

	for i := range databases {
		if predicate(&databases[i]) {
			if found == nil {
				// Create a copy to avoid returning a reference to the slice element
				dbCopy := databases[i]
				found = &dbCopy
			} else {
				return nil, fmt.Errorf("%w for %q: %q, %q", errMultipleMatches, q, found.Name, databases[i].Name)
			}
		}
	}

	if found == nil {
		return nil, fmt.Errorf("%w: %q", errDatabaseNotFound, q)
	}

	return found, nil
}

// Each applies the given callback function to each database reference.
// It stops iteration and returns the error if the callback returns an error.
func Each(callback func(DbReference) error) error {
	for i := range databases {
		if err := callback(databases[i]); err != nil {
			return err
		}
	}

	return nil
}
