package profiles

import (
	"fmt"
	"strings"

	"github.com/halsatif/freshctl/internal/catalog"
)

const Version = 1

type Profile struct {
	Version  int      `json:"version"`
	Name     string   `json:"name"`
	Packages []string `json:"packages"`
}

func Validate(profile Profile, packages []catalog.Package) error {
	if profile.Version != Version {
		return fmt.Errorf("unsupported profile version %d", profile.Version)
	}
	if strings.TrimSpace(profile.Name) != profile.Name {
		return fmt.Errorf("profile name should not have leading or trailing spaces")
	}
	if len(profile.Packages) == 0 {
		return fmt.Errorf("profile must include at least one package")
	}

	known := make(map[string]bool, len(packages))
	for _, pkg := range packages {
		known[pkg.PackageID] = true
	}

	seen := make(map[string]bool, len(profile.Packages))
	for _, packageID := range profile.Packages {
		id := strings.TrimSpace(packageID)
		if id == "" {
			return fmt.Errorf("profile contains an empty package id")
		}
		if id != packageID {
			return fmt.Errorf("package id %q should not have leading or trailing spaces", packageID)
		}
		if seen[id] {
			return fmt.Errorf("duplicate package id %q", id)
		}
		if !known[id] {
			return fmt.Errorf("unknown package id %q", id)
		}
		seen[id] = true
	}

	return nil
}

func PackageIDs(profile Profile) []string {
	return append([]string{}, profile.Packages...)
}
