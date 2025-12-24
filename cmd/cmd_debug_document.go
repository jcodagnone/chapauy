// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jcodagnone/chapauy/impo"
	"github.com/jcodagnone/chapauy/utils/htmlutils"
	"github.com/spf13/cobra"
)

var debugDocumentCmd = &cobra.Command{
	Use:   "document [file]",
	Short: "Lee un documento HTML y extrae las ofensas en formato JSON.",
	Long: `Lee un documento HTML desde un archivo o desde la entrada estándar,
y extrae la información de las ofensas de tránsito, imprimiéndola en formato JSON.

Ejemplos:
  cat ./impo/testdata/canelones.html | go run main.go debug document
  go run main.go debug document ./impo/testdata/canelones.html`,
	Run: func(_ *cobra.Command, args []string) {
		var (
			r   io.Reader
			err error
		)

		if len(args) > 0 {
			r, err = os.Open(args[0])
			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
		} else {
			r = os.Stdin
			if isTerminal(os.Stdin) {
				fmt.Fprintln(os.Stderr, "Reading from stdin. Paste HTML and press Ctrl+D to finish.")
			}
		}

		node, err := htmlutils.AsNode(r)
		if err != nil {
			log.Fatalf("error parsing html: %v", err)
		}

		var issuers []string
		err = impo.Each(func(db impo.DbReference) error {
			issuers = append(issuers, db.Issuers...)

			return nil
		})
		if err != nil {
			log.Fatalf("error listing issuers: %v", err)
		}

		notification, err := impo.ExtractDocument(issuers, "", node)
		if err != nil {
			log.Fatalf("error extracting document: %v", err)
		}

		output, err := json.MarshalIndent(notification, "", "  ")
		if err != nil {
			log.Fatalf("error marshalling json: %v", err)
		}

		fmt.Println(string(output))
	},
}
