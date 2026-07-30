package installer

import (
	"context"
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/providers"
)

const testProviderType catalog.ProviderType = "TestProvider"

type fakeProvider struct {
	called            bool
	applicationID     string
	providerPackageID string
}

func (p *fakeProvider) Type() catalog.ProviderType {
	return testProviderType
}

func (p *fakeProvider) Validate(catalog.Application, catalog.Provider) error {
	return nil
}

func (p *fakeProvider) Install(_ context.Context, app catalog.Application, provider catalog.Provider, opts providers.InstallOptions) error {
	p.called = true
	p.applicationID = app.ID
	p.providerPackageID = provider.PackageID
	if opts.Log != nil {
		opts.Log("fake provider installed")
	}
	return nil
}

func TestInstallAppsUsesApplicationProvider(t *testing.T) {
	providerInstaller := &fakeProvider{}
	providers.Register(providerInstaller)

	app := catalog.Application{
		Name: "Fake App",
		ID:   "fake-app",
		Providers: []catalog.Provider{{
			Type:      testProviderType,
			PackageID: "provider-fake-app",
			Strategy:  catalog.InstallStrategyPackageManager,
		}},
	}

	events := collectInstallEvents(app)

	if !providerInstaller.called {
		t.Fatal("expected install flow to call application provider")
	}
	if providerInstaller.applicationID != "fake-app" || providerInstaller.providerPackageID != "provider-fake-app" {
		t.Fatalf("installer should receive separate application and provider IDs, got app=%q provider=%q", providerInstaller.applicationID, providerInstaller.providerPackageID)
	}
	if !hasEventLine(events, "fake provider installed") {
		t.Fatal("expected provider log line to be forwarded")
	}
	if !hasSuccessfulResult(events, app.ID) {
		t.Fatal("expected fake provider install to succeed")
	}
}

func TestInstallAppsHandlesUnknownProvider(t *testing.T) {
	app := catalog.Application{
		Name: "Unknown App",
		ID:   "unknown-app",
		Providers: []catalog.Provider{{
			Type:      "UnknownProvider",
			PackageID: "unknown-app",
			Strategy:  catalog.InstallStrategyPackageManager,
		}},
	}

	events := collectInstallEvents(app)

	if !hasEventLine(events, "unknown package provider: UnknownProvider") {
		t.Fatal("expected readable unknown provider error")
	}
	if !hasFailedResult(events, app.ID, "unknown package provider: UnknownProvider") {
		t.Fatal("expected unknown provider to be reported as failed result")
	}
}

func TestInstallAppsHandlesMissingProvider(t *testing.T) {
	app := catalog.Application{
		Name: "Providerless App",
		ID:   "providerless-app",
	}

	events := collectInstallEvents(app)

	if !hasEventLine(events, "no install provider configured for Providerless App") {
		t.Fatal("expected readable missing provider error")
	}
	if !hasFailedResult(events, app.ID, "no install provider configured for Providerless App") {
		t.Fatal("expected missing provider to be reported as failed result")
	}
}

func TestCommandForResolvesProvider(t *testing.T) {
	app := catalog.Application{
		Name: "Git",
		ID:   "git-application",
		Providers: []catalog.Provider{{
			Type:      catalog.ProviderChocolatey,
			PackageID: "git",
			Strategy:  catalog.InstallStrategyPackageManager,
		}},
	}

	if got := CommandFor(app); got != "choco install git -y --no-progress" {
		t.Fatalf("unexpected command %q", got)
	}
}

func collectInstallEvents(app catalog.Application) []Event {
	events := make(chan Event)
	go InstallApps(context.Background(), []catalog.Application{app}, events, nil)

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
			if result.App.ID == packageID && result.Success {
				return true
			}
		}
	}
	return false
}

func hasFailedResult(events []Event, packageID, message string) bool {
	for _, event := range events {
		for _, result := range event.Results {
			if result.App.ID != packageID || result.Success || result.Err == nil {
				continue
			}
			if strings.Contains(result.Err.Error(), message) {
				return true
			}
		}
	}
	return false
}
