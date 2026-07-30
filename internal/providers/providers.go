package providers

import (
	"context"

	"github.com/halsatif/freshctl/internal/catalog"
)

type InstallOptions struct {
	Log  func(string)
	Skip <-chan struct{}
}

type Installer interface {
	Type() catalog.ProviderType
	Install(ctx context.Context, app catalog.Application, provider catalog.Provider, opts InstallOptions) error
}

type CommandProvider interface {
	Command(app catalog.Application, provider catalog.Provider) string
}

var registry = map[catalog.ProviderType]Installer{}

func Register(installer Installer) {
	if installer == nil || installer.Type() == "" {
		return
	}
	registry[installer.Type()] = installer
}

func Get(providerType catalog.ProviderType) (Installer, bool) {
	installer, ok := registry[providerType]
	return installer, ok
}

func init() {
	Register(&Chocolatey{})
}
