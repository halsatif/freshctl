package providers

import (
	"context"
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestChocolateyIsRegistered(t *testing.T) {
	installer, ok := Get(catalog.ProviderChocolatey)
	if !ok {
		t.Fatal("Chocolatey provider should be registered")
	}
	if installer.Type() != catalog.ProviderChocolatey {
		t.Fatalf("expected Chocolatey provider type, got %q", installer.Type())
	}
}

func TestUnimplementedProvidersAreNotRegistered(t *testing.T) {
	for _, providerType := range []catalog.ProviderType{
		catalog.ProviderWinget,
		catalog.ProviderDirect,
		catalog.ProviderCommunity,
	} {
		if _, ok := Get(providerType); ok {
			t.Fatalf("%s should not be registered before its installer exists", providerType)
		}
	}
}

func TestUnknownProviderLookup(t *testing.T) {
	if _, ok := Get("MissingProvider"); ok {
		t.Fatal("unexpected installer registered for MissingProvider")
	}
}

func TestChocolateyCommand(t *testing.T) {
	installer := &Chocolatey{}
	command := installer.Command(catalog.Application{Name: "Git", ID: "git"}, catalog.Provider{
		Type:      catalog.ProviderChocolatey,
		PackageID: "git",
		Strategy:  catalog.InstallStrategyPackageManager,
	})

	if command != "choco install git -y --no-progress" {
		t.Fatalf("unexpected Chocolatey command %q", command)
	}
}

func TestChocolateyCommandIncludesPrerelease(t *testing.T) {
	installer := &Chocolatey{}
	command := installer.Command(catalog.Application{Name: "Zen Browser", ID: "zen-browser"}, catalog.Provider{
		Type:      catalog.ProviderChocolatey,
		PackageID: "zen-browser",
		Strategy:  catalog.InstallStrategyPackageManager,
		Metadata:  catalog.ProviderMetadata{Prerelease: true},
	})

	if !strings.Contains(command, "--pre") {
		t.Fatalf("expected prerelease command to include --pre, got %q", command)
	}
}

type noopInstaller struct{}

func (noopInstaller) Type() catalog.ProviderType {
	return "NoopProvider"
}

func (noopInstaller) Install(context.Context, catalog.Application, catalog.Provider, InstallOptions) error {
	return nil
}

func TestRegisterProvider(t *testing.T) {
	Register(noopInstaller{})

	installer, ok := Get("NoopProvider")
	if !ok {
		t.Fatal("registered provider should be available")
	}
	if installer.Type() != "NoopProvider" {
		t.Fatalf("unexpected provider type %q", installer.Type())
	}
}
