package profiles

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/halsatif/freshctl/internal/catalog"
)

const DefaultExportPath = "freshctl-profile.json"
const DefaultProfileName = "freshctl profile"

func FromPackages(name string, packages []catalog.Application) Profile {
	ids := make([]string, 0, len(packages))
	for _, pkg := range packages {
		ids = append(ids, pkg.ID)
	}
	return Profile{
		Version:  Version,
		Name:     name,
		Packages: ids,
	}
}

func WriteJSON(path string, profile Profile, catalogPackages []catalog.Application) error {
	if err := Validate(profile, catalogPackages); err != nil {
		return err
	}

	content, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	content = append(content, '\n')

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}
	return nil
}
