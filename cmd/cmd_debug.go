// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jcodagnone/chapauy/impo"
	"github.com/spf13/cobra"
)

// we say that it isn't.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Dev tools",
}

var debugMatriculasCmd = &cobra.Command{
	Use:   "matriculas",
	Short: "Interacuar con el módulo de extracción de información de mátriculas",
	Long: `Lee una matrícula por línea, e imprime en stdout la matrícula seguida de la
información inferida.

$ echo BTX1234 | chapauy  debug matriculas
BTX1234		{Country:UY AdmDivision:B … Category:Taxi MercosurFormat:true}
	`,
	Run: func(_ *cobra.Command, _ []string) {
		input := os.Stdin
		if isTerminal(input) {
			fmt.Fprintln(os.Stderr, "Ingrese mátriculas a analizar, una por línea…")
		}
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			plate := scanner.Text()
			info, err := impo.AnalyzeVehicleID(plate, "")
			if err != nil {
				fmt.Printf("%s\t%q\n", plate, err)
			} else {
				if s, err := json.Marshal(info); err == nil {
					fmt.Printf("%s\t\t%s\n", plate, s)
				} else {
					log.Fatal(err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(debugMatriculasCmd)
	debugCmd.AddCommand(debugDocumentCmd)
}
