package sources

import (
	"context"
	"errors"

	"github.com/halsatif/freshctl/internal/catalog"
)

var ErrWingetNotImplemented = errors.New("winget source not implemented yet")

type WingetSource struct{}

func (s *WingetSource) ID() string {
	return string(catalog.PackageSourceWinget)
}

func (s *WingetSource) Install(context.Context, catalog.Package, InstallOptions) error {
	return ErrWingetNotImplemented
}
