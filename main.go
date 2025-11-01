// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/jcodagnone/chapauy/cmd"
)

var Version = "development"

func main() {
	cmd.Execute(Version)
}
