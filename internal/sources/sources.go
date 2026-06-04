package sources

import (
	"context"

	"github.com/halsatif/freshctl/internal/catalog"
)

type InstallOptions struct {
	Log  func(string)
	Skip <-chan struct{}
}

type Source interface {
	ID() string
	Install(ctx context.Context, pkg catalog.Package, opts InstallOptions) error
}

type CommandSource interface {
	Command(pkg catalog.Package) string
}

var registry = map[string]Source{}

func Register(source Source) {
	if source == nil || source.ID() == "" {
		return
	}
	registry[source.ID()] = source
}

func GetSource(id string) (Source, bool) {
	source, ok := registry[id]
	return source, ok
}

func init() {
	Register(&ChocolateySource{})
	Register(&WingetSource{})
}
