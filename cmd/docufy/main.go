package main

import (
	"os"

	"github.com/mlihgenel/docufy/v2/internal/entrypoint"
)

var (
	// ldflags ile override edilebilir:
	//   -X main.version=1.5.0 -X main.buildDate=2026-02-21T12:00:00Z
	version   = "dev"
	buildDate = ""
)

func main() {
	os.Exit(entrypoint.Run(version, buildDate))
}
