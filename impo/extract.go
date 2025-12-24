// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

// Package impo handles traffic offense notifications processing.
package impo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jcodagnone/chapauy/spatial"
	"github.com/jcodagnone/chapauy/utils/htmlutils"
	"github.com/mattn/go-isatty"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/html"
)

// UR represents Unidad Reajustable.
// We encode as an integer to avoid losing precision with 0.5 values.
// The value is stored as 10× the actual value (e.g., 5.5 UR is stored as 55).
type UR int

const urResolution = 100

// String formats the UR value back to a printable string.
func (ur UR) String() string {
	a, b := int(ur)/urResolution, int(ur)%urResolution
	if b == 0 {
		return strconv.Itoa(a)
	}

	return strconv.Itoa(a) + "." + strings.TrimRight(strconv.Itoa(b), "0")
}

// Converts from string representation `0,5' in the of UR to its internal representation.
func parseUR(s string) (UR, error) {
	if s == "" {
		return 0, nil
	}
	// motocycle some times is written like 2XPERS
	for _, expr := range []string{
		"XPERS",
		" X PERS",
		" X PERS.",
		" X PERS,",
		" X ASIEN",
		"UR", // https://www.impo.com.uy/bases/notificaciones-transito-treintaytres/5-2023
	} {
		s = strings.TrimRight(s, expr)
	}

	// Split integer and fractional parts
	parts := strings.SplitN(
		// sometimes is stored like 0,5
		strings.ReplaceAll(s, ",", "."),
		".",
		2,
	)

	// Parse the integer part
	a, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w %q: %w", errParseInt, parts[0], err)
	}

	ret := a * urResolution

	// If there is a fractional part, process it.
	if len(parts) == 2 {
		b := parts[1]
		// Pad the fractional part with a trailing zero if needed.
		n := len(b)
		if n > 2 {
			return 0, errors.New("only one decimal is supported")
		}

		switch n {
		case 1:
			b += "0"
		case 0:
			b = "00"
		}
		// Parse the fractional part
		fraction, err := strconv.ParseInt(b, 10, 64)
		if err != nil || fraction < 0 {
			return 0, fmt.Errorf("parsing fractional part %q: %w", b, err)
		}

		ret += fraction
	}

	return UR(ret), nil
}

// UruguayTimezone is the time location for Uruguay.
var UruguayTimezone = func() *time.Location {
	tz, err := time.LoadLocation("America/Montevideo")
	if err != nil {
		panic(err)
	}

	return tz
}()

// Some dates have bad spacing like "25/09/2023 1 2:02".
func parseDateTime(s string) time.Time {
	var ret time.Time

	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimRight(s, "HS")
	s = strings.TrimLeft(s, "/") // https://www.impo.com.uy/bases/notificaciones-cgm/2933-2024

	if s == "" {
		return ret
	}

	var err error

	formats := []string{
		"2/1/200615:04",
		"2/1/0615.04",
		"2/1/200615.04",
		"2/1/0615:04",
		"2/1/200615:4", // seen @ https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/80-2021
		"2006-01-0215:04",
		// date only
		"2/1/2006",
		"2/1/006",
		"2/1/06",
		"02-01-2006", // seen @ https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/533-2024
		"06-02-01",   // seen @ https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/416-2024
		"1/2/2006",   // seet @ https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/75-2024
	}
	for _, fmt := range formats {
		ret, err = time.ParseInLocation(fmt, s, UruguayTimezone)
		if err == nil {
			return ret
		}
	}

	return time.Time{}
}

// Document contains offenses.
type Document struct {
	DocSource string    `json:"doc_src,omitempty"`
	DocID     string    `json:"doc_id,omitempty"`
	DocDate   time.Time `json:"doc_date"`
}

// TrafficOffense represents a single traffic violation.
type TrafficOffense struct {
	*Document
	*VehicleInfo
	DbID            int            `json:"repo_id"`
	RecordID        int            `json:"record_id,omitempty"` // Position in the original document
	Vehicle         string         `json:"vehicle"`             // Vehicle identifier, e.g. 'ABC1234'
	Time            time.Time      `json:"time"`                // Time of offense with minute precision
	Location        string         `json:"location"`            // Location, e.g. 'Bvar. Artigas S y Frente a Municipio'
	DisplayLocation string         `json:"display_location,omitempty"`
	ID              string         `json:"id"`              // Authority ID, e.g. 'IDM 0000000000'
	Description     string         `json:"description"`     // Offense description, e.g. 'Exceso de velocidad hasta 20 km/h'
	UR              UR             `json:"ur"`              // Fine amount in UR
	Error           string         `json:"error,omitempty"` // The error that occurred
	Point           *spatial.Point `json:"point,omitempty"` // Geocoded point
	ArticleIDs      []string       `json:"article_id"`
	ArticleCodes    []int8         `json:"article_codes"`
	H3Res1          uint64         `json:"h3_res1"`
	H3Res2          uint64         `json:"h3_res2"`
	H3Res3          uint64         `json:"h3_res3"`
	H3Res4          uint64         `json:"h3_res4"`
	H3Res5          uint64         `json:"h3_res5"`
	H3Res6          uint64         `json:"h3_res6"`
	H3Res7          uint64         `json:"h3_res7"`
	H3Res8          uint64         `json:"h3_res8"`
}

// OffenseProperty represents a property of a traffic offense.
type OffenseProperty int

const (
	propVehicle = iota
	propTime
	propLocation
	propID
	propDescription
	propUR
	// specials.
	propLocalidad
	propHora
	propCountry
	// used to ignore columns.
	propIgnore
)

// phrases. This function maps these phrases to the concepts.
func documentPropertyFromString(s string) (OffenseProperty, error) {
	ns := normalize(s)

	for prop, names := range map[OffenseProperty][]string{
		propVehicle: {
			"Matrícula",
			"Matrícula y padrón",
			"ATRICULA", // https://www.impo.com.uy/bases/resoluciones-transito-rionegro/116-2023
			"MATRICLA", // https://www.impo.com.uy/bases/notificaciones-transito-treintaytres/38-2024
			"MAT.",     // https://www.impo.com.uy/bases/notificaciones-transito-colonia/78-2025
		},
		propTime: {
			"Fecha y Hora",
			"Fecha-Hora",
			"Fecha-Hola", // https://www.impo.com.uy/bases/notificaciones-transito-movilidad-maldonado/172-2025
			"Fecha",
			"Fecha Ingreso",
		},
		propLocation: {
			"Intersección",
			"ntersección", // https://www.impo.com.uy/bases/notificaciones-cgm/57-2017
			"Lugar",
			"Ubicación",
		},
		propID: {
			"Intervenido",
			"Serie-Boleta",
			"ID_BOLETA",
			"ID",
		},
		propDescription: {
			"Artículo",
			"INFRACCION",
			"Nom. Tributo",
			"Detalle",
			"Detalles",
			"Multa",
			"CONDUCTOR", // https://www.impo.com.uy/bases/notificaciones-transito-colonia/76-2025
		},
		propUR: {
			"Valor en UR",
			"Valor UR",
			"Valor Total",
			"Valor",
			"UR",
			"Monto",
		},
		// Lavalleja provee informacion adicional de localidad
		propLocalidad: {
			"Localidad",
		},
		// Lavalleja separa la hora del día https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/25-2025
		propHora: {
			"Hora",
		},
		// Caminera arrancó a exponerlo desde https://impo.com.uy/bases/resoluciones-policia-caminera/1000-2025
		// Esto viene de https://www.gub.uy/congreso-intendentes/comunicacion/noticias/multas-transito-vehiculos-matricula-extranjera
		// Esta instrucción se imparte porque el sistema informático no distingue matrículas nacionales de extranjeras. Por ese motivo
		// el dato de la procedencia debe ser preciso por constituir un factor central para su correcta visualización.
		// A título informativo, por ejemplo, las motos de Uruguay y los autos de origen argentino –con matrículas anteriores
		// a la del Mercosur-, comparten la misma estructura de “3 letras + 3 números”, por lo que, si al anotarse la infracción se
		//  la marca como “vehículo nacional”, la misma irá directamente al Sucive, y si lo marcan como “vehículo extranjero”
		//  irá al nuevo departamento “extranjeros”. De la forma en que se haga esta anotación en el sistema, dependerá
		//  la correcta visualización como vehículo extranjero desde las plataformas del Sucive.
		propCountry: {
			"Pais",
			"País",
		},
		propIgnore: {
			"Unidad",   // TODO(juan) caminera tengo que volver a esto
			"Cantidad", // TODO(juan) caminera tengo que volver a esto
			"CI.",      // Colonia desde https://www.impo.com.uy/bases/notificaciones-transito-colonia/76-2025 reporta cedula
		},
	} {
		for _, name := range names {
			normalizedName := normalize(name)
			if ns == normalizedName {
				return prop, nil
			}
		}
	}

	return 0, fmt.Errorf("unknown property for header %q", s)
}

// Assigns a value to the appropriate field based on the index.
func (record *TrafficOffense) set(i OffenseProperty, s string) error {
	switch i {
	case propVehicle:
		record.Vehicle = NormalizeVehicleID(s)
	case propTime:
		if s != "" {
			record.Time = parseDateTime(s)
			if record.Time.IsZero() {
				return fmt.Errorf("%w: %q", errParseDateTime, s)
			}
		}
	case propLocation:
		record.Location = s
	case propID:
		record.ID = s
	case propDescription:
		// lots of Maldonado starts with " : …"
		record.Description = strings.TrimLeft(s, ": ")
	case propUR:
		ur, err := parseUR(s)
		if err != nil {
			return fmt.Errorf("can't convert %q to UR: %w", s, err)
		}

		record.UR = ur
	case propCountry:
		country, err := normalizeCountryName(s)
		if err != nil {
			// preferimos cortar ejecución para poder cargar el País, en vez
			// de almacenarlo con error o ignorarlo
			panic(err)
		}

		if country != "" {
			if record.VehicleInfo == nil {
				record.VehicleInfo = &VehicleInfo{}
			}

			record.VehicleInfo.Country = country
		}
	case propIgnore:
		// skip
	default:
		return fmt.Errorf("don't know how to handle index %d", i)
	}

	return nil
}

var vehiclePattern = regexp.MustCompile("(?i)^[A-Z0-9]{4,10}$")
var (
	errInvalidVehicle = errors.New("matrícula inválida")
	errMissingTime    = errors.New("falta horario")
	errParseInt       = errors.New("parsing integer part")
	errParseDateTime  = errors.New("couldn't parse datetime")
)

const suciveArt9Descr = "Cobros por acciones, trámites o gestiones"

// Validate checks if the traffic offense record is valid.
func (record *TrafficOffense) Validate() error {
	if record.Error != "" {
		return nil
	}

	if record.Vehicle == "" || !vehiclePattern.MatchString(record.Vehicle) {
		return errInvalidVehicle
	}

	if record.Time.IsZero() {
		return errMissingTime
	}

	if record.Time.Before(time.Date(2015, 1, 1, 0, 0, 0, 0, UruguayTimezone)) {
		return fmt.Errorf("la fecha `%v' es anterior a 2015-01-01", record.Time)
	}

	if record.Description == "" {
		return errors.New("falta descripción")
	}

	return nil
}

// ExtractMetrics tracks statistics about the parsing process.
type ExtractMetrics struct {
	NewRecords     int
	NewErrors      int
	SuccessfulDocs int
	FailedDocs     int
}

// Merge combines two ParseMetrics.
func (m *ExtractMetrics) Merge(o *ExtractMetrics) *ExtractMetrics {
	m.NewRecords += o.NewRecords
	m.NewErrors += o.NewErrors
	m.SuccessfulDocs += o.SuccessfulDocs
	m.FailedDocs += o.FailedDocs

	return m
}

// Extracts offenses from the HTML table.
func visitOffensesTable(
	child *html.Node,
	offenses *[]*TrafficOffense,
	defaultDate *time.Time,
	defaultDescription string,
) error {
	nr := 0
	// Map to store the column index to property mapping
	columnMap := make(map[int]OffenseProperty)

	for child := child.FirstChild; child != nil; child = child.NextSibling {
		// We're interested in <tr> elements
		if child.Type != html.ElementNode || !strings.EqualFold("tr", child.Data) {
			continue
		}

		sb := strings.Builder{}

		if nr == 0 {
			// Process header row to determine column mapping
			i := 0

			for child := child.FirstChild; child != nil; child = child.NextSibling {
				if child.Type != html.ElementNode || !strings.EqualFold("td", child.Data) {
					continue
				}

				sb.Reset()

				err := htmlutils.Node2string(child, &sb)
				if err != nil {
					continue
				}

				columnMap[i], err = documentPropertyFromString(sb.String())
				if err != nil {
					return err
				}

				i++
			}

			hasDescriptionCol := false

			for _, prop := range columnMap {
				if prop == propDescription {
					hasDescriptionCol = true

					break
				}
			}

			if !hasDescriptionCol && defaultDescription == "" {
				return errors.New("tabla sin columna descripción")
			}

			nr++

			continue
		}

		hasDateCol := false

		for _, prop := range columnMap {
			if propTime == prop {
				hasDateCol = true

				break
			}
		}

		record := TrafficOffense{}
		record.RecordID = nr

		if !hasDateCol {
			// some documents like https://www.impo.com.uy/bases/notificaciones-transito-colonia/1-2023 don't
			// have an infraction date available. To avoid discarting the records, we assume that the record
			// is the one of the document
			record.Time = *defaultDate
		}

		if defaultDescription != "" {
			record.Description = defaultDescription
		}

		var lastErr error // Track the first error for each record

		i := 0

		// casos especiales de Lavalleja que envia la fecha y el lugar separado
		// recolectamos los valores parciales mientras recorremos las columnas
		// para luega intentar usarlos
		var hora, fecha, localidad string

		for child := child.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != html.ElementNode || !strings.EqualFold("td", child.Data) {
				continue
			}

			sb.Reset()

			err := htmlutils.Node2string(child, &sb)
			if err == nil {
				s := sb.String()
				// Get the property for this column index
				if prop, exists := columnMap[i]; exists {
					switch prop {
					case propHora:
						hora = s
					case propLocalidad:
						localidad = s
					case propTime:
						fecha = s
						err = record.set(prop, s)
					default:
						err = record.set(prop, s)
					}
				} else {
					err = fmt.Errorf("no property for index %d", i)
				}
			}

			if err != nil && lastErr == nil {
				lastErr = err
			}

			i++
		}

		// merge special split columns
		if localidad != "" && record.Location != "" {
			record.Location = fmt.Sprintf("%s, %s", record.Location, localidad)
		}

		if !record.Time.IsZero() && fecha != "" {
			if t := parseDateTime(fmt.Sprintf("%s %s", fecha, hora)); !t.IsZero() {
				record.Time = t
			}
		}

		if lastErr == nil {
			lastErr = record.Validate()
		}

		if lastErr == nil && !record.Time.IsZero() && record.Time.After(*defaultDate) {
			// ver PAV1450 en https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/16-2024
			lastErr = fmt.Errorf("la fecha `%v' es más nueva que la fecha de publicación `%v'", record.Time, *defaultDate)
		}

		if lastErr != nil {
			record.Error = lastErr.Error()
		}

		*offenses = append(*offenses, &record)

		nr++
	}

	return nil
}

// Traverses the HTML document searching for offenses and metadata.
func visitDocument(
	issuers []string,
	doc *Document,
	offenses *[]*TrafficOffense,
	defaultDescription *string,
	n *html.Node,
) error {
	// Look for a table with class="tabla_en_texto"
	var isTable bool

	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "table":
			for _, attr := range n.Attr {
				isTable = isTable || (strings.EqualFold("class", attr.Key) && attr.Val == "tabla_en_texto")
			}
		case "title":
			sb := strings.Builder{}

			err := htmlutils.Node2string(n, &sb)
			if err != nil {
				return err
			}

			// Title: 'Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado N° 1/025'
			title := strings.ToLower(sb.String())

			// Detect the issuer
			for _, issuer := range issuers {
				if len(issuer) > 0 {
					if idx := strings.Index(title, strings.ToLower(issuer)); idx > -1 {
						title = strings.TrimSpace(title[idx+len(issuer):])
						// Extract notification ID (e.g., "N° 1/025" -> "1/025")
						idx := strings.LastIndex(title, " ")

						if idx >= 0 && idx < len(title)-1 {
							title = title[idx+1:]
							doc.DocID = title

							break
						} else if title == "s/n" {
							doc.DocID = title

							break
						}
					}
				}
			}
		case "h5":
			// Extract publication date: "Fecha de Publicación: 08/04/2025"
			sb := strings.Builder{}

			err := htmlutils.Node2string(n, &sb)
			if err != nil {
				return err
			}

			title := strings.ToLower(sb.String())

			const expected = "fecha de publicación:"

			if idx := strings.LastIndex(title, expected); idx > -1 {
				title = strings.TrimSpace(title[idx+len(expected):])

				var err error

				doc.DocDate, err = time.ParseInLocation("02/01/2006", title, UruguayTimezone)
				if err != nil {
					return err
				}
			}
		case "p", "pre", "div":
			if *defaultDescription != "" {
				break
			}

			// normalizeText normalizes the text by squashing multiple spaces into one and lowercasing.
			normalizeText := func(s string) string {
				return strings.Join(strings.Fields(strings.ToLower(s)), " ")
			}

			sb := strings.Builder{}
			if err := htmlutils.Node2string(n, &sb); err == nil {
				text := normalizeText(sb.String())
				// Phrases to search for.
				phrases := []string{
					"que se constató la contravención a lo dispuesto en el art. 9 del texto ordenado del sucive",
					"que el cuerpo inspectivo constató la contravención a lo dispuesto en el art 9 del texto ordenado del sucive",
					"que la intendencia de montevideo, constató la contravención a lo dispuesto en el artículo 9 del texto ordenado del sucive",
				}

				for _, phrase := range phrases {
					if strings.Contains(text, phrase) {
						*defaultDescription = suciveArt9Descr

						break
					}
				}
			}
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		var err error
		if isTable {
			err = visitOffensesTable(
				child,
				offenses,
				&doc.DocDate,
				*defaultDescription,
			)
		} else {
			err = visitDocument(issuers, doc, offenses, defaultDescription, child)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// ExtractDocument extracts traffic offense information from HTML.
func ExtractDocument(issuers []string, n *html.Node) ([]*TrafficOffense, error) {
	doc := &Document{}
	offenses := make([]*TrafficOffense, 0, 800)

	var defaultDescription string

	if err := visitDocument(issuers, doc, &offenses, &defaultDescription, n); err != nil {
		return nil, err
	}

	// Assign the document to each offense
	for _, offense := range offenses {
		offense.Document = doc
	}

	return offenses, nil
}

// Converts HTML document to JSON extracting notifications.
func (c *Client) extractDocument(id string) (*ExtractMetrics, error) {
	failedMetrics := &ExtractMetrics{
		FailedDocs: 1,
	}
	r, err := c.store.GetDocument(id)

	if err != nil {
		return failedMetrics, fmt.Errorf("opening document %s: %w", id, err)
	}

	node, err := htmlutils.AsNode(r)

	if closeErr := r.Close(); closeErr != nil {
		return failedMetrics, fmt.Errorf("closing document: %w", closeErr)
	}

	if err != nil {
		return failedMetrics, fmt.Errorf("parsing document: %w", err)
	}

	offenses, err := ExtractDocument(c.dbRef.Issuers, node)
	if err != nil {
		return failedMetrics, fmt.Errorf("parsing document: %w", err)
	}

	if len(offenses) > 0 {
		offenses[0].Document.DocSource = id
	}

	for _, o := range offenses {
		o.DbID = c.dbRef.ID
	}

	var errorsCount int

	var firstError error

	for _, offense := range offenses {
		if offense.Error != "" {
			errorsCount++

			if firstError == nil {
				firstError = errors.New(offense.Error)
			}
		}
	}

	successCount := len(offenses) - errorsCount

	failedMetrics = &ExtractMetrics{
		NewRecords: successCount,
		NewErrors:  errorsCount,
		FailedDocs: 1,
	}
	if len(offenses) > 0 && offenses[0].DocID == "" {
		return failedMetrics, errors.New("document ID not found")
	}

	if n := float64(successCount); n > 0 {
		// we have a failsafe that fail to save documents with more than 5% of errors
		// this allows us to catch extraction errors
		if pct := float64(errorsCount) / n * 100.0; pct > 5.0 {
			switch id {
			// and all these case are cases with more than 5% but that had been reviewed as
			// ok. usually they have low number of total records
			case
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/6-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-colonia/18-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-colonia/19-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-colonia/104-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/2211-2023",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/7-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/14-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/31-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/17-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/11-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/12-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/13-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/15-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/20-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/22-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/25-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/33-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/34-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-lavalleja/37-2025",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/52-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/93-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/231-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/244-2025",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/257-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/425-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/551-2024",
				"https://www.impo.com.uy/bases/resoluciones-transito-lavalleja/334-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-soriano/204-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-tacuarembo/7-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-tacuarembo/9-2024",
				"https://www.impo.com.uy/bases/notificaciones-transito-tacuarembo/37-2025_A",
				"https://www.impo.com.uy/bases/notificaciones-transito-tacuarembo/41-2025",
				"https://www.impo.com.uy/bases/notificaciones-transito-treintaytres/14-2024",
				"https://www.impo.com.uy/bases/notificaciones-cgm/1709-2022",
				"https://www.impo.com.uy/bases/notificaciones-cgm/3183-2024",
				"https://www.impo.com.uy/bases/notificaciones-cgm/3458-2025",
				"https://www.impo.com.uy/bases/resoluciones-transito-mtop/207-2025":
				// these documents where reviewed, and aren't issues related to the scrapper
			default:
				return failedMetrics, fmt.Errorf("parsing document - too many errors - %2.f%%: for example: %w", pct, firstError)
			}
		}
	}

	if !c.options.DryRun && (errorsCount == 0 || !c.options.SkipErrDocs) {
		if err := c.repo.SaveTrafficOffenses(offenses); err != nil {
			return failedMetrics, fmt.Errorf("storing document: %w", err)
		}
	}

	if errorsCount > 0 && c.options.SkipErrDocs {
		var tmp []error

		for _, offense := range offenses {
			if offense.Error != "" {
				tmp = append(tmp, errors.New(offense.Error))
			}
		}

		err = errors.Join(tmp...)

		return failedMetrics, err
	}

	return &ExtractMetrics{
		NewRecords:     successCount,
		NewErrors:      errorsCount,
		SuccessfulDocs: 1,
	}, nil
}

// Extracts JSON from downloaded HTML documents.
func (c *Client) extractDocuments() error {
	var docs []string

	var err error

	if c.options.ExtractFull {
		docs, err = c.store.ExistingDocuments()
	} else {
		// get all local HTML documents
		allDocs, err := c.store.ExistingDocuments()
		if err != nil {
			return fmt.Errorf("getting all local documents: %w", err)
		}

		// get all extracted documents from the database
		extractedDocs, err := c.repo.GetExtractedDocuments(c.dbRef)
		if err != nil {
			return fmt.Errorf("getting extracted documents: %w", err)
		}

		// find the documents that have not been extracted yet
		for _, doc := range allDocs {
			if _, ok := extractedDocs[doc]; !ok {
				docs = append(docs, doc)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("getting documents to extract: %w", err)
	}

	slices.Sort(docs)
	n := len(docs)

	maxProcs := c.options.ExtractMaxProcs
	if maxProcs == 0 {
		maxProcs = runtime.NumCPU()
	}

	var bar *progressbar.ProgressBar
	if isatty.IsTerminal(os.Stderr.Fd()) {
		bar = progressbar.NewOptions(n,
			progressbar.OptionSetDescription("Extracting "+c.dbRef.Name),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionClearOnFinish(),
		)
	}

	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxProcs)
	errChan := make(chan error, n)
	metricsChan := make(chan *ExtractMetrics, n)

	for _, id := range docs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()
			semaphore <- struct{}{}

			defer func() { <-semaphore }()

			metrics, err := c.extractDocument(id)
			if err != nil {
				errChan <- fmt.Errorf("extracting %s - %w", id, err)
			}

			if metrics != nil {
				metricsChan <- metrics
			}

			if bar == nil {
				log.Printf("Extracting %s", id)
			} else {
				if err := bar.Add(1); err != nil {
					errChan <- fmt.Errorf("updating progress bar for %s: %w", id, err)
				}
			}
		}(id)
	}

	wg.Wait()
	close(errChan)
	close(metricsChan)

	for err := range errChan {
		log.Printf("Extraction failed - %s", err)
	}

	for metrics := range metricsChan {
		c.Metrics.ExtractMetrics.Merge(metrics)
	}

	log.Printf(
		"Extraction phase complete - %d new records, %d errors from %d documents, %d successful and %d failed.",
		c.Metrics.NewRecords,
		c.Metrics.NewErrors,
		c.Metrics.SuccessfulDocs+c.Metrics.FailedDocs,
		c.Metrics.SuccessfulDocs,
		c.Metrics.FailedDocs,
	)

	return nil
}
