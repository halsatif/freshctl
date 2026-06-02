package profiles

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/halsatif/freshctl/internal/catalog"
)

func ReadJSON(path string, catalogPackages []catalog.Package) (Profile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Profile{}, fmt.Errorf("%s not found", path)
		}
		return Profile{}, fmt.Errorf("read profile: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(content, &profile); err != nil {
		return Profile{}, fmt.Errorf("invalid JSON")
	}
	if err := Validate(profile, catalogPackages); err != nil {
		return Profile{}, err
	}
	return profile, nil
}
