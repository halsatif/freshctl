package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/halsatif/freshctl/internal/catalog"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	outputPath := filepath.Join("site", "catalog.generated.js")
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	content, err := catalog.GeneratedCatalogJS(catalog.Default())
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, content, 0o644)
}
