// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package impo

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
)

func TestVisitSearchTable(t *testing.T) {
	tests := []struct {
		expected SearchResults
		input    string
	}{
		{
			SearchResults{
				Next: "tipoServicio=45&realizarconsulta=SI&idconsulta=abcdefgh&nrodocdesdehasta=201-250",
				Entries: []SearchResultEntry{
					{
						Title:    "Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 320/025",
						Subtitle: "",
						Href:     "/bases/notificaciones-transito-maldonado/320-2025",
					},
					{
						Title:    "Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 319/025",
						Subtitle: "NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO",
						Href:     "/bases/notificaciones-transito-maldonado/319-2025",
					},
				},
			},
			`<html lang="es">
  <body>
    <table class="table table-hover">
      <tbody id="resultadoConsulta" class="rowlink" data-link="row">
        <tr class="sitem">
          <td width="60px">1.</td>
          <td>
            <a class="visitado" href="/bases/notificaciones-transito-maldonado/320-2025">
              <strong>Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 320/025 </strong>
              <font color="#014895">(Documento original)</font>
            </a>
            <font size="2"></font>
          </td>
        </tr>
        <tr class="sitem">
          <td width="60px">2.</td>
          <td>
            <a class="visitado" href="/bases/notificaciones-transito-maldonado/319-2025" >
              <strong>Notificación Dirección General de Tránsito y Transporte Intendencia de Maldonado 319/025 </strong>
              <font color="#014895">(Documento original)</font>
            </a>
            <font size="2"><br />NOTIFICACION POR CONTRAVENCION A NORMAS DE TRANSITO</font>
          </td>
        </tr>
      </tbody>
    </table>
	<ul class="pager">
		<li><a class="nextPage"
	 		href="/cgi-bin/bases/consultaBasesBS.cgi?tipoServicio=45&realizarconsulta=SI&idconsulta=abcdefgh&nrodocdesdehasta=201-250" onMouseOver="status='Ir a la página siguiente';return true" onMouseOut="status='';return true"  onClick="setValoresSubmit();">
				Siguiente
			</a>
		</li>
	</ul>
  </body>
</html>`,
		},
	}

	for _, test := range tests {
		node, err := html.Parse(strings.NewReader(test.input))
		if nil != err {
			t.Error(err)

			return
		}

		r, err := parseSearches(
			node,
		)
		if err != nil {
			t.Fatalf("parseReader: `%s': %s", test.input, err)
		}

		if diff := cmp.Diff(test.expected, *r); diff != "" {
			t.Errorf("parse output  missmatch (-expected +got):\n%s", diff)
		}
	}
}
