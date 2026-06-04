package sources

import (
	"context"
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestChocolateySourceIsRegistered(t *testing.T) {
	source, ok := GetSource(string(catalog.PackageSourceChocolatey))
	if !ok {
		t.Fatal("Chocolatey source should be registered")
	}
	if source.ID() != string(catalog.PackageSourceChocolatey) {
		t.Fatalf("expected Chocolatey source ID, got %q", source.ID())
	}
}

func TestUnknownSourceLookup(t *testing.T) {
	if _, ok := GetSource("MissingSource"); ok {
		t.Fatal("unexpected source registered for MissingSource")
	}
}

func TestChocolateyCommand(t *testing.T) {
	source := &ChocolateySource{}
	command := source.Command(catalog.Package{
		Name:      "Git",
		PackageID: "git",
		Source:    catalog.PackageSourceChocolatey,
	})

	if command != "choco install git -y --no-progress" {
		t.Fatalf("unexpected Chocolatey command %q", command)
	}
}

func TestChocolateyCommandIncludesPrerelease(t *testing.T) {
	source := &ChocolateySource{}
	command := source.Command(catalog.Package{
		Name:       "Zen Browser",
		PackageID:  "zen-browser",
		Source:     catalog.PackageSourceChocolatey,
		Prerelease: true,
	})

	if !strings.Contains(command, "--pre") {
		t.Fatalf("expected prerelease command to include --pre, got %q", command)
	}
}

type noopSource struct{}

func (s noopSource) ID() string {
	return "NoopSource"
}

func (s noopSource) Install(context.Context, catalog.Package, InstallOptions) error {
	return nil
}

func TestRegisterSource(t *testing.T) {
	Register(noopSource{})

	source, ok := GetSource("NoopSource")
	if !ok {
		t.Fatal("registered source should be available")
	}
	if source.ID() != "NoopSource" {
		t.Fatalf("unexpected source ID %q", source.ID())
	}
}
