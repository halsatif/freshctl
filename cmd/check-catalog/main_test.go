package main

import (
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestValidProviderTypeAcceptsKnownTypes(t *testing.T) {
	for _, providerType := range []catalog.ProviderType{
		catalog.ProviderChocolatey,
		catalog.ProviderWinget,
		catalog.ProviderDirect,
		catalog.ProviderCommunity,
	} {
		if !validProviderType(providerType) {
			t.Fatalf("expected provider type %q to be valid", providerType)
		}
	}
}

func TestValidProviderTypeRejectsUnknownProvider(t *testing.T) {
	if validProviderType(catalog.ProviderType("UnknownProvider")) {
		t.Fatal("unknown provider type should not be valid")
	}
}
