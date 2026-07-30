package catalog

import "testing"

func TestPrimaryProviderReturnsFirstDeclaredProvider(t *testing.T) {
	app := Application{
		ID: "example-app",
		Providers: []Provider{
			{Type: ProviderChocolatey, PackageID: "example-choco", Strategy: InstallStrategyPackageManager},
			{Type: ProviderWinget, PackageID: "Example.App", Strategy: InstallStrategyPackageManager},
		},
	}

	provider, ok := app.PrimaryProvider()
	if !ok {
		t.Fatal("expected primary provider")
	}
	if provider.Type != ProviderChocolatey || provider.PackageID != "example-choco" {
		t.Fatalf("unexpected primary provider %#v", provider)
	}
}

func TestApplicationWithoutProvidersHasNoPrimaryProvider(t *testing.T) {
	if _, ok := (Application{ID: "no-provider"}).PrimaryProvider(); ok {
		t.Fatal("application without providers should not have a primary provider")
	}
}

func TestProviderByType(t *testing.T) {
	app := Application{
		ID: "example-app",
		Providers: []Provider{
			{Type: ProviderChocolatey, PackageID: "example-choco", Strategy: InstallStrategyPackageManager},
			{Type: ProviderWinget, PackageID: "Example.App", Strategy: InstallStrategyPackageManager},
		},
	}

	provider, ok := app.ProviderByType(ProviderWinget)
	if !ok || provider.PackageID != "Example.App" {
		t.Fatalf("expected Winget provider, got %#v ok=%v", provider, ok)
	}
}
