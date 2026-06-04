package installer

import (
	"context"
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/sources"
)

type fakeSource struct {
	called bool
}

func (s *fakeSource) ID() string {
	return "TestSource"
}

func (s *fakeSource) Install(_ context.Context, _ catalog.Package, opts sources.InstallOptions) error {
	s.called = true
	if opts.Log != nil {
		opts.Log("fake source installed")
	}
	return nil
}

func TestInstallAppsUsesPackageSource(t *testing.T) {
	source := &fakeSource{}
	sources.Register(source)

	app := catalog.Package{
		Name:      "Fake App",
		PackageID: "fake-app",
		Source:    catalog.PackageSource(source.ID()),
	}

	events := collectInstallEvents(app)

	if !source.called {
		t.Fatal("expected install flow to call package source")
	}
	if !hasEventLine(events, "fake source installed") {
		t.Fatal("expected source log line to be forwarded")
	}
	if !hasSuccessfulResult(events, app.PackageID) {
		t.Fatal("expected fake source install to succeed")
	}
}

func TestInstallAppsHandlesUnknownSource(t *testing.T) {
	app := catalog.Package{
		Name:      "Unknown App",
		PackageID: "unknown-app",
		Source:    catalog.PackageSource("UnknownSource"),
	}

	events := collectInstallEvents(app)

	if !hasEventLine(events, "unknown package source: UnknownSource") {
		t.Fatal("expected readable unknown source error")
	}
	if !hasFailedResult(events, app.PackageID, "unknown package source: UnknownSource") {
		t.Fatal("expected unknown source to be reported as failed result")
	}
}

func TestCommandForResolvesSource(t *testing.T) {
	app := catalog.Package{
		Name:      "Git",
		PackageID: "git",
		Source:    catalog.PackageSourceChocolatey,
	}

	if got := CommandFor(app); got != "choco install git -y --no-progress" {
		t.Fatalf("unexpected command %q", got)
	}
}

func collectInstallEvents(app catalog.Package) []Event {
	events := make(chan Event)
	go InstallApps(context.Background(), []catalog.Package{app}, events, nil)

	collected := make([]Event, 0)
	for event := range events {
		collected = append(collected, event)
	}
	return collected
}

func hasEventLine(events []Event, want string) bool {
	for _, event := range events {
		if strings.Contains(event.Line, want) {
			return true
		}
	}
	return false
}

func hasSuccessfulResult(events []Event, packageID string) bool {
	for _, event := range events {
		for _, result := range event.Results {
			if result.App.PackageID == packageID && result.Success {
				return true
			}
		}
	}
	return false
}

func hasFailedResult(events []Event, packageID, message string) bool {
	for _, event := range events {
		for _, result := range event.Results {
			if result.App.PackageID != packageID || result.Success || result.Err == nil {
				continue
			}
			if strings.Contains(result.Err.Error(), message) {
				return true
			}
		}
	}
	return false
}
