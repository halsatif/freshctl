package detection

import (
	"os/exec"
	"strings"

	"github.com/halsatif/freshctl/internal/catalog"
)

func DetectInstalled(pkg catalog.Package) bool {
	switch pkg.DetectMethod {
	case catalog.DetectRegistry:
		return DetectRegistry(pkg.DetectValue)
	case catalog.DetectPath:
		return DetectPath(pkg.DetectValue)
	default:
		return false
	}
}

func HasDetectionMetadata(pkg catalog.Package) bool {
	return pkg.DetectMethod != catalog.DetectNone && strings.TrimSpace(pkg.DetectValue) != ""
}

func DetectPath(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	_, err := exec.LookPath(value)
	return err == nil
}

func MatchRegistryDisplayName(displayName, detectValue string) bool {
	displayName = strings.ToLower(strings.TrimSpace(displayName))
	detectValue = strings.ToLower(strings.TrimSpace(detectValue))
	if displayName == "" || detectValue == "" {
		return false
	}
	return strings.Contains(displayName, detectValue)
}
