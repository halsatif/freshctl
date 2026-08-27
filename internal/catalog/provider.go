package catalog

type ProviderType string

const (
	ProviderChocolatey ProviderType = "Chocolatey"
	ProviderWinget     ProviderType = "Winget"
	ProviderDirect     ProviderType = "Direct"
	ProviderCommunity  ProviderType = "Community"
)

type InstallStrategy string

const (
	InstallStrategyPackageManager  InstallStrategy = "PackageManager"
	InstallStrategyDirectInstaller InstallStrategy = "DirectInstaller"
)

type InstallerType string

const (
	InstallerTypeExecutable InstallerType = "Executable"
	InstallerTypeMSI        InstallerType = "MSI"
)

type InstallerArchitecture string

const (
	InstallerArchitectureAny   InstallerArchitecture = ""
	InstallerArchitectureX64   InstallerArchitecture = "x64"
	InstallerArchitectureARM64 InstallerArchitecture = "arm64"
)

type DirectDownload struct {
	URL                string
	GitHubRepository   string
	GitHubAssetPattern string
	Architecture       InstallerArchitecture
}

type DirectInstallerMetadata struct {
	Downloads      []DirectDownload
	Filename       string
	SilentArgs     []string
	InstallerType  InstallerType
	ChecksumSHA256 string
}

type ProviderMetadata struct {
	Prerelease bool
	Direct     *DirectInstallerMetadata
}

type Provider struct {
	Type      ProviderType
	PackageID string
	Strategy  InstallStrategy
	Metadata  ProviderMetadata
}

func (app Application) PrimaryProvider() (Provider, bool) {
	if len(app.Providers) == 0 {
		return Provider{}, false
	}
	return app.Providers[0], true
}

func (app Application) ProviderByType(providerType ProviderType) (Provider, bool) {
	for _, provider := range app.Providers {
		if provider.Type == providerType {
			return provider, true
		}
	}
	return Provider{}, false
}

func chocolateyProvider(packageID string) Provider {
	return Provider{
		Type:      ProviderChocolatey,
		PackageID: packageID,
		Strategy:  InstallStrategyPackageManager,
	}
}

func directProvider(packageID string, metadata DirectInstallerMetadata) Provider {
	return Provider{
		Type:      ProviderDirect,
		PackageID: packageID,
		Strategy:  InstallStrategyDirectInstaller,
		Metadata:  ProviderMetadata{Direct: &metadata},
	}
}
