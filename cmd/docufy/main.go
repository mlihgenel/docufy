package main

import (
	"os"

	"github.com/mlihgenel/docufy/v2/internal/entrypoint"
)

var (
	// ldflags ile override edilebilir:
	//   -X main.version=2.1.0 -X main.buildDate=2026-03-07T12:00:00Z
	version   = "dev"
	buildDate = ""
)

func main() {
	os.Exit(entrypoint.Run(version, buildDate))
}
