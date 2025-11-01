// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type logWriter struct {
	writer io.Writer
}

func (w *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Fprintf(w.writer, "%s %s", time.Now().Format("2006-01-02 15:04:05"), string(bytes))
}

func init() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{writer: os.Stderr})
}

var rootCmd = &cobra.Command{
	Use:   "chapa",
	Short: "infracciones y multas de tránsito uruguayas",
	Long: `
chapa permite acceder de forma programática a la información contenida en las
Notificaciones y Resoluciones publicadas en el Diario Oficial del Centro de
Información Oficial.
`,
}

var Version = "dev"

func Execute(version string) {
	Version = version

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
