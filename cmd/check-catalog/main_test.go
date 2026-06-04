package main

import (
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestValidSourceAcceptsRegisteredSources(t *testing.T) {
	for _, source := range []catalog.PackageSource{
		catalog.PackageSourceChocolatey,
		catalog.PackageSourceWinget,
	} {
		if !validSource(source) {
			t.Fatalf("expected source %q to be valid", source)
		}
	}
}

func TestValidSourceRejectsUnknownSource(t *testing.T) {
	if validSource(catalog.PackageSource("UnknownSource")) {
		t.Fatal("unknown source should not be valid")
	}
}
