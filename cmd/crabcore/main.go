package main

import (
	"os"

	"github.com/zerodotfive/crabcore/cmd/crabcore/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
