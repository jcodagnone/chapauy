// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/net/html"
)

func TestUnidadReajustables(t *testing.T) {
	tests := []struct {
		s        string
		wantVal  UR
		asString string
		wantErr  bool
	}{
		{"", 0, "0", false},
		{"1", 100, "1", false},
		{"0", 0, "0", false},
		{"0,0", 0, "0", false},
		{"0.0", 0, "0", false},
		{"0,25", 0.25 * urResolution, "0.25", false},
		{"0,5", 0.5 * urResolution, "0.5", false},
		{"0.5", 0.5 * urResolution, "0.5", false},
		{"0.50", 0.5 * urResolution, "0.5", false},
		{"0,-1", 0, "", true},
		{"0,9", 0.9 * urResolution, "0.9", false},
		{"0,123", 0, "", true},
		{"5", 5 * urResolution, "5", false},
		{"2XPERS", 2 * urResolution, "2", false},
	}

	for _, tc := range tests {
		t.Run(tc.s, func(t *testing.T) {
			val, err := parseUR(tc.s)
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("unexpected error: %s", err)
				}

				return
			}

			if int(val) != int(tc.wantVal) {
				t.Fatalf("val: expected %d got %d", tc.wantVal, val)
			}

			if expected, got := tc.asString, val.String(); expected != got {
				t.Fatalf("String() want %v, got %v", expected, got)
			}
		})
	}

	if expected, got := "12.3", UR(12.3*urResolution).String(); expected != got {
		t.Fatalf("want %v, got %v", expected, got)
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     time.Time
		wantZero bool
	}{
		{
			name:     "empty string",
			input:    "",
			wantZero: true,
		},
		{
			name:     "only spaces",
			input:    "   ",
			wantZero: true,
		},
		{
			name:     "incomplete time",
			input:    "25/09/2023 15:",
			wantZero: true,
		},
		{
			name:  "DD/MM/YYYY hh:mm",
			input: "25/09/2023 15:04",
			want:  time.Date(2023, time.September, 25, 15, 4, 0, 0, UruguayTimezone),
		},
		{
			name:  "YYYY-MM-DD hh:mm",
			input: "2023-09-25 15:04",
			want:  time.Date(2023, time.September, 25, 15, 4, 0, 0, UruguayTimezone),
		},
		{
			name:  "YYYY-MM-DD hh:mm  HS",
			input: "4/1/24 22:00 HS",
			want:  time.Date(2024, time.January, 4, 22, 0, 0, 0, UruguayTimezone),
		},
		{
			name:  "DD/MM/YYYY hh:mm - don't care about spaces",
			input: "25/09/2023 1 5:04",
			want:  time.Date(2023, time.September, 25, 15, 4, 0, 0, UruguayTimezone),
		},
		{
			name:  "single digit day and month",
			input: "1/2/2023 15:04",
			want:  time.Date(2023, time.February, 1, 15, 4, 0, 0, UruguayTimezone),
		},
		{
			name:  "DD/MM/YYYY",
			input: "29/10/2024",
			want:  time.Date(2024, time.October, 29, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			name:  "DD/MM/YY",
			input: "28/04/23",
			want:  time.Date(2023, time.April, 28, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			name:     "DD-MM-YYYY",
			input:    "25-09-2023",
			want:     time.Date(2023, time.September, 25, 0, 0, 0, 0, UruguayTimezone),
			wantZero: false,
		},
		{
			name:     "21/04/24 17:00HS",
			input:    "21/04/24 17:00HS",
			want:     time.Date(2024, time.April, 21, 17, 0, 0, 0, UruguayTimezone),
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDateTime(tt.input)

			if tt.wantZero {
				if !got.IsZero() {
					t.Errorf("parseDateTime wanted nil, got %v", got)
				}

				return
			}

			if got.IsZero() {
				t.Errorf("parseDateTime wanted %v, got nil", tt.want)

				return
			}

			if !got.Equal(tt.want) {
				t.Errorf("parseDateTime mismatch, got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOffenseValidate(t *testing.T) {
	tests := []struct {
		input       *TrafficOffense
		expectedErr bool
	}{
		{
			input: &TrafficOffense{
				Vehicle:     "ZME2015",
				Description: "Foo",
				Time:        time.Date(2025, 1, 1, 0, 0, 0, 0, UruguayTimezone),
				UR:          UR(1),
			},
			expectedErr: false,
		},
		{
			input: &TrafficOffense{
				Description: "Foo",
				Time:        time.Date(2025, 1, 1, 0, 0, 0, 0, UruguayTimezone),
				UR:          UR(1),
			},
			expectedErr: true,
		},
		{
			input: &TrafficOffense{
				Vehicle:     "ZME2015",
				Time:        time.Date(2025, 1, 1, 0, 0, 0, 0, UruguayTimezone),
				UR:          UR(1),
				Description: "",
			},
			expectedErr: true,
		},
		{
			input: &TrafficOffense{
				Vehicle:     "ZME2015",
				Description: "Foo",
				UR:          UR(1),
			},
			expectedErr: true,
		},
		{
			input: &TrafficOffense{
				Vehicle:     "ZME2015",
				Description: "Foo",
				Time:        time.Date(2025, 1, 1, 0, 0, 0, 0, UruguayTimezone),
				UR:          UR(1),
				Error:       "some error",
			},
			expectedErr: false,
		},
		{
			input: &TrafficOffense{
				Vehicle:     "ZME2015",
				Description: "Foo",
				Time:        time.Date(2014, 12, 31, 23, 59, 59, 0, UruguayTimezone),
				UR:          UR(1),
			},
			expectedErr: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectedErr && err == nil {
				t.Fatal("expected an error")
			}

			if !tt.expectedErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestTrafficOffensePropertyFrom(t *testing.T) {
	tests := []struct {
		input       string
		want        OffenseProperty
		expectedErr bool
	}{
		{
			input:       "Matrícula",
			want:        propVehicle,
			expectedErr: false,
		},
		{
			input:       "  matricula  ",
			want:        propVehicle,
			expectedErr: false,
		},

		{
			input:       "Fecha y Hora",
			want:        propTime,
			expectedErr: false,
		},

		{
			input:       "Intersección",
			want:        propLocation,
			expectedErr: false,
		},

		{
			input:       "Intervenido",
			want:        propID,
			expectedErr: false,
		},
		{
			input:       "Artículo",
			want:        propDescription,
			expectedErr: false,
		},

		{
			input:       "Valor en UR",
			want:        propUR,
			expectedErr: false,
		},
		{
			input:       "Valor en U.R.",
			want:        propUR,
			expectedErr: false,
		},
		{
			input:       "Localidad",
			want:        propLocalidad,
			expectedErr: false,
		},
		{
			input:       "Hora",
			want:        propHora,
			expectedErr: false,
		},
		// Error cases
		{
			input:       "SomethingUnknown",
			want:        0, // Default value
			expectedErr: true,
		},
		{
			input:       "",
			want:        0, // Default value
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			prop, err := documentPropertyFromString(tt.input)

			// Check error expectation
			if (err != nil) != tt.expectedErr {
				t.Errorf("trafficOffensePropertyFrom() error = %v, expectedErr = %v", err, tt.expectedErr)

				return
			}

			// If no error expected, check the property value
			if !tt.expectedErr && prop != tt.want {
				t.Errorf("trafficOffensePropertyFrom() got = %v, want = %v", prop, tt.want)
			}
		})
	}
}

func TestVisitHTML(t *testing.T) {
	tests := []struct {
		expected TrafficOffense
		input    string
		docID    string
		date     time.Time
	}{
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "ZME2015",
				Location:    "Ruta Interbalnearia y Rosa de los Vientos",
				Time:        time.Date(2025, 1, 1, 0, 0, 0, 0, UruguayTimezone),
				ID:          "IDM 0000000000",
				Description: "Exceso de velocidad hasta 20 km/h",
				UR:          UR(5 * urResolution),
			},
			`
			<html>
				<title>Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado N° 1/025</title>
				<h5>Fecha de Publicación:    01/02/2025    </h5>
				<table class="tabla_en_texto">
					<TR>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Matricula</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Fecha y Hora</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Interseccion</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Intervenido</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Articulo</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Valor en UR</pre></TD>
					</TR>
				 	<TR>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>ZME2015</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>01/01/2025 00:00</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Ruta Interbalnearia y Rosa de los Vientos</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>IDM 0000000000</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Exceso de velocidad hasta 20 km/h</pre></TD>
					  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>5</pre></TD>
				   </TR>
				</table>
			</html>
			`,
			"1/025",
			time.Date(2025, 2, 1, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "ZME2015",
				Location:    "WILLIMAN DR. CLAUDIO RBLA. Y PARIS AVDA.",
				Time:        time.Date(2024, 12, 18, 20, 5, 0, 0, UruguayTimezone),
				ID:          "FM14 1144",
				Description: "15.4 No respetar señales luminosas",
				UR:          6 * urResolution,
			},
			`
			<html>
			    <title>Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado N° 1/024</title>
				<h5>Fecha de Publicación: 27/12/2024 </h5>
				<table class="tabla_en_texto">
					<TR>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Matricula</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Fecha y Hora</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Interseccion</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Intervenido</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Articulo</pre></TD>
					  <TD style="text-align:center;vertical-align:bottom;border-width:1px 1px 1px 1px;" ><pre>Valor en UR</pre></TD>
					</TR>
					<TR>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>ZME2015</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>2024-12-18 20:05</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>WILLIMAN DR. CLAUDIO RBLA. Y PARIS AVDA.</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>FM14 1144</pre></TD>
					  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>15.4 No respetar señales luminosas</pre></TD>
					  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>6</pre></TD>
				   </TR>
				</table>
			</html>
			`,
			"1/024",
			time.Date(2024, 12, 27, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "ZME2015",
				Description: "No respetar señales luminosas",
				UR:          4 * urResolution,
				Time:        time.Date(2022, 0o5, 0o2, 0, 0, 0, 0, UruguayTimezone),
			},
			`
			<html>
			    <title>Resolución Dirección de Tránsito Intendencia de Lavalleja N° 917/021</title>
				<h5>Fecha de Publicación: 02/05/2022 </h5>
				<TABLE class="tabla_en_texto" style="width:100%;">
				 <TR>
				  <TD style="text-align:center;vertical-align:center;border-width:1px 1px 1px 1px;" ><pre>Matrícula</pre></TD>
				  <TD style="text-align:center;vertical-align:center;border-width:1px 1px 1px 1px;" ><pre>Artículo</pre></TD>
				  <TD style="text-align:center;vertical-align:center;border-width:1px 1px 1px 1px;" ><pre>Valor en UR</pre></TD>
				 </TR>
				 <TR>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>ZME   201 5</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>: No respetar señales luminosas</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>4</pre></TD>
				 </TR>
			</html>
			`,
			"917/021",
			time.Date(2022, 5, 2, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "ABF5416",
				Location:    "Ruta 7 y Km 36",
				Time:        time.Date(2025, 11, 5, 11, 48, 0, 0, UruguayTimezone),
				ID:          "DPC 9999000604",
				Description: "Exceso de velocidad de entre 21 km/h y 30 km/h",
				UR:          UR(8 * urResolution),
				VehicleInfo: &VehicleInfo{
					Country: ISOUruguay,
				},
			},
			`
			<html>
				<title>Resolución Policía Caminera N° 1000/025</title>
				<h5>Fecha de Publicación: 06/11/2025 </h5>
				<TABLE class="tabla_en_texto" style="width:100%;">
				 <TR>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Matrícula</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Pais</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Fecha y Hora</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Intersección</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Intervenido</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Artículo</pre></TD>
				  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Valor en UR</pre></TD>
				   </TR>
				 <TR>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>ABF5416</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>URUGUAY</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>05/11/2025 11:48</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Ruta 7 y Km 36</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>DPC 9999000604</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Exceso de velocidad de entre 21 km/h y 30 km/h</pre></TD>
				  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>8</pre></TD>
				   </TR>
				</TABLE>
			</html>
			`,
			"1000/025",
			time.Date(2025, 11, 6, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "SBU3238",
				Location:    "Avda. Gral. Fructuoso Rivera y Avda. Luis Alberto de Herrera",
				Time:        time.Date(2025, 5, 23, 0, 0, 0, 0, UruguayTimezone),
				ID:          "5042880",
				Description: suciveArt9Descr,
				UR:          UR(0),
			},
			`
			<html>
				<title>Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado N° 1/024</title>
				<h5>Fecha de Publicación: 29/05/2025 </h5>
				<pre>Notifícase a los propietarios de los vehículos cuyas matrículas se determinan, que se constató la contravención a lo dispuesto en el art. 9 del Texto Ordenado del Sucive.</pre>
				<TABLE class="tabla_en_texto" style="width:100%;">
				 <TR>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Matrícula</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Fecha</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Lugar</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Serie-Boleta</pre></TD>
				   </TR>
				 <TR>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>SBU3238</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>23/05/2025</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Avda. Gral. Fructuoso Rivera y Avda. Luis Alberto de Herrera</pre></TD>
				  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>5042880</pre></TD>
				   </TR>
				</TABLE>
			</html>
			`,
			"1/024",
			time.Date(2025, 5, 29, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID: 1,
				Vehicle:  "AAA1880",
				// No location or date logic change for this test, focusing on description
				// In 2.html, the table doesn't have date/place, wait, let me check 2.html content again.
				// 2.html table only has Matricula. But document header has date?
				// Looking at 2.html provided in prompt...
				// It has "Fecha de Publicación: 14/01/2019"
				// Table has ONLY "Matrícula".
				// So Time should default to DocDate?
				Time:        time.Date(2019, 1, 14, 0, 0, 0, 0, UruguayTimezone),
				Description: suciveArt9Descr,
				UR:          UR(0),
			},
			`
			<html>
				<title>Resolución Centro de Gestión de Movilidad N° 137/018</title>
				<h5>Fecha de Publicación: 14/01/2019 </h5>
				<pre>                            Notificación 248/018

Notifícase a los propietarios de los vehículos cuyas matrículas se determinan, que se constató la contravención a lo dispuesto en el art. 9 del Texto Ordenado del Sucive.
(5.637)
</pre><pre>
   El Director del Departamento de Movilidad de la Intendencia de Montevideo ha dispuesto por Expediente 2018-1504-98-000025, notificar a los propietarios de los vehículos cuya matrícula se detalla en planilla adjunta, que la Intendencia de Montevideo, constató la contravención a lo Dispuesto en el artículo 9 del Texto Ordenado del Sucive.
   Se dispone la presente notificación confiriendo vista de estas actuaciones a los titulares de los vehículos relacionados, conforme lo establece el Digesto Departamental. A partir de la publicación cuenta con un plazo de 10 días hábiles a efectos de realizar los descargos que entienda corresponden.
   En este acto se le hace saber que en caso de que no se regularicen los adeudos por concepto de patente de rodados dentro del plazo para adherir al Plan de Regularización de Adeudos de Sucive 2018 aprobado por el Congreso de Intendentes el 14/3/18, se le aplicará la sanción establecida en el artículo 9 del Texto Ordenado del Sucive (multa equivalente al 25% de la patente anual). El valor de la patente anual puede ser consultado en www.sucive.gub.uy.

    VISTO: Las detecciones efectuadas por los equipos de fiscalización, y/o inspectores de tránsito.
   RESULTANDO:
I) Que el cuerpo inspectivo constató la contravención a lo dispuesto en
   el Art 9 del Texto Ordenado del Sucive.
II)     Que se dió vista personal de estas actuaciones a los interesados
   propietarios de los vehículos;

<TABLE class="tabla_en_texto" style="width:100%;">
 <TR>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Matrícula</pre></TD>
   </TR>
 <TR>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>AAA1880</pre></TD>
   </TR>
</TABLE>
			</html>
			`,
			"137/018",
			time.Date(2019, 1, 14, 0, 0, 0, 0, UruguayTimezone),
		},
		{
			TrafficOffense{
				RecordID:    1,
				Vehicle:     "SAB5624",
				Time:        time.Date(2022, 4, 2, 8, 37, 0, 0, UruguayTimezone),
				Location:    "AV ITALIA y AV BOLIVIA",
				Description: suciveArt9Descr,
				UR:          UR(0),
				ID:          "PAT 6570012510",
			},
			`
			<html>
				<title>Notificación Centro de Gestión de Movilidad N° 1684/022</title>
				<h5>Fecha de Publicación: 17/06/2022 </h5>
				<h4><pre>GOBIERNOS DEPARTAMENTALES
INTENDENCIAS
INTENDENCIA DE MONTEVIDEO
CENTRO DE GESTIÓN DE MOVILIDAD
</pre></h4><pre>                           Notificación 1.684/022

Notifícase al propietario del vehículo cuya matrícula se determina, que se constató la contravención a las normas vigentes de tránsito, por lo que correspondería la aplicación de una multa.
(1.725)
</pre><pre>
   El Director del Departamento de Movilidad de la Intendencia de Montevideo ha dispuesto en la fecha 14/06/2022 10:26:51, notificar a los propietarios de los vehículos cuya matrícula se detalla en planilla adjunta, que la Intendencia de Montevideo, constató la contravención a lo Dispuesto en el artículo 9 del Texto Ordenado del Sucive.
   Se dispone la presente notificación confiriendo vista de estas actuaciones a los titulares de los vehículos relacionados, conforme lo establece el Digesto Departamental. A partir de la publicación cuenta con un plazo de 10 días hábiles a efectos de realizar los descargos que entienda corresponden.
   La sanción establecida en el artículo 9 del Texto Ordenado del Sucive, dispone una multa equivalente al 25% de la patente anual. El valor de la patente anual puede ser consultado en www.sucive.gub.uy

<TABLE class="tabla_en_texto" style="width:100%;">
 <TR>
  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Matrícula</pre></TD>
  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Fecha y Hora</pre></TD>
  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Intersección</pre></TD>
  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>Intervenido</pre></TD>
   </TR>
 <TR>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>SAB5624</pre></TD>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>2/4/2022 8:37</pre></TD>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>AV ITALIA y AV BOLIVIA</pre></TD>
  <TD style="vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>PAT 6570012510</pre></TD>
   </TR>
</TABLE>
			</html>
			`,
			"1684/022",
			time.Date(2022, 6, 17, 0, 0, 0, 0, UruguayTimezone),
		},
	}

	for _, test := range tests {
		node, err := html.Parse(strings.NewReader(test.input))
		if nil != err {
			t.Error(err)

			return
		}

		n, err := ExtractDocument(
			[]string{
				"dirección general de tránsito y transporte intendencia de maldonado",
				"dirección de tránsito intendencia de lavalleja",
				"policía caminera",
				"centro de gestión de movilidad",
			},
			"",
			node,
		)
		if err != nil {
			t.Fatalf("parseReader: `%s': %s", test.input, err)
		}

		if len(n) == 0 {
			t.Fatal("expected offenses, got none")
		}

		offenseDoc := n[0].Document
		if expected, actual := test.docID, offenseDoc.DocID; expected != actual {
			t.Errorf("docId - %q != %q", expected, actual)
		}

		if expected, actual := test.date, offenseDoc.DocDate; expected != actual {
			t.Errorf("date - %q != %q", expected, actual)
		}

		if expected, actual := 1, len(n); expected != actual {
			t.Errorf("len(Offenses) - %d != %d", expected, actual)
		}

		if diff := cmp.Diff(&test.expected, n[0], cmpopts.IgnoreFields(TrafficOffense{}, "Document")); diff != "" {
			t.Errorf("parse output mismatch (-expected +got):\n%s", diff)
		}
	}
}

func TestVisitOffensesTable_WithLocalidadAndHora(t *testing.T) {
	htmlInput := `
	<TABLE class="tabla_en_texto" style="width:100%;">
	 <TR>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>MATRICULA</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>INFRACCION</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>VALOR</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>FECHA</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>HORA</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>UBICACIÓN</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>LOCALIDAD</pre></TD>
	   </TR>
	 <TR>
	  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre> ABE 8658</pre></TD>
	  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>ADELANTAR POR LA DERECHA</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>3</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>31/03/2024</pre></TD>
	  <TD style="text-align:center;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>17:27</pre></TD>
	  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>L.A. DE HERRERA Y LAVALLEJA</pre></TD>
	  <TD style="text-align:left;vertical-align:top;border-width:1px 1px 1px 1px;" ><pre>MINAS</pre></TD>
	   </TR>
	</TABLE>
	`

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	var tableNode *html.Node

	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "tabla_en_texto" {
					tableNode = n

					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTable(c)

			if tableNode != nil {
				return
			}
		}
	}

	findTable(doc)

	if tableNode == nil {
		t.Fatal("could not find table node")
	}

	var offenses []*TrafficOffense

	defaultDate := time.Now()

	var tbodyNode *html.Node

	for c := tableNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			tbodyNode = c

			break
		}
	}

	if tbodyNode == nil {
		t.Fatal("could not find tbody node")
	}

	err = visitOffensesTable(tbodyNode, &offenses, &defaultDate, "", nil)
	if err != nil {
		t.Fatalf("visitOffensesTable returned an error: %v", err)
	}

	if len(offenses) != 1 {
		t.Fatalf("expected 1 offense, got %d", len(offenses))
	}

	if offenses[0].Error != "" {
		t.Fatalf("expected no error, but got: %s", offenses[0].Error)
	}

	expected := &TrafficOffense{
		RecordID:    1,
		Vehicle:     "ABE8658",
		Description: "ADELANTAR POR LA DERECHA",
		UR:          UR(3 * urResolution),
		Time:        time.Date(2024, time.March, 31, 17, 27, 0, 0, UruguayTimezone),
		Location:    "L.A. DE HERRERA Y LAVALLEJA, MINAS",
	}

	if diff := cmp.Diff(expected, offenses[0]); diff != "" {
		t.Errorf("mismatch (-expected +got):\n%s", diff)
	}
}

func TestVisitHTMLWithArt9(t *testing.T) {
	htmlInput := `
	<html>
		<title>Notificación Dirección General de Tránsito y Transporte Intendencia de Montevideo N° 3906/025</title>
		<h5>Fecha de Publicación: 10/12/2025</h5>
		<p>... que se constató la contravención a lo dispuesto en el art. 9 del Texto Ordenado del Sucive.</p>
		<table class="tabla_en_texto">
			<TR>
				<TD><pre>Matricula</pre></TD>
				<TD><pre>Fecha y Hora</pre></TD>
			</TR>
			<TR>
				<TD><pre>SBF1234</pre></TD>
				<TD><pre>10/12/2025 10:00</pre></TD>
			</TR>
		</table>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	offenses, err := ExtractDocument([]string{"intendencia de montevideo"}, "", doc)
	if err != nil {
		t.Fatalf("ExtractDocument failed: %v", err)
	}

	if len(offenses) != 1 {
		t.Fatalf("expected 1 offense, got %d", len(offenses))
	}

	if offenses[0].Description != suciveArt9Descr {
		t.Errorf("expected description '%s', got '%s'", suciveArt9Descr, offenses[0].Description)
	}
}

func TestVisitHTMLWithMissingHeaders(t *testing.T) {
	htmlInput := `
	<html>
		<title>Notificación Dirección General de Tránsito y Transporte Intendencia de Treinta y Tres N° 14/024</title>
		<h5>Fecha de Publicación: 10/12/2024</h5>
		<table class="tabla_en_texto">
			<TR>
				<TD><pre>SBF1234</pre></TD>
				<TD><pre>Exceso de velocidad</pre></TD>
				<TD><pre>5</pre></TD>
			</TR>
		</table>
	</html>
	`

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	offenses, err := ExtractDocument(
		[]string{"intendencia de treinta y tres"},
		"https://www.impo.com.uy/bases/notificaciones-transito-treintaytres/14-2024",
		doc,
	)
	if err != nil {
		t.Fatalf("ExtractDocument failed: %v", err)
	}

	if len(offenses) != 1 {
		t.Fatalf("expected 1 offense, got %d", len(offenses))
	}

	if offenses[0].Vehicle != "SBF1234" {
		t.Errorf("expected vehicle 'SBF1234', got '%s'", offenses[0].Vehicle)
	}

	if offenses[0].Description != "Exceso de velocidad" {
		t.Errorf("expected description 'Exceso de velocidad', got '%s'", offenses[0].Description)
	}

	if offenses[0].UR != 5*urResolution {
		t.Errorf("expected UR 5, got %v", offenses[0].UR)
	}
}
