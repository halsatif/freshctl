//go:build windows

package detection

import (
	"golang.org/x/sys/windows/registry"
)

func DetectRegistry(value string) bool {
	for _, root := range uninstallRegistryRoots() {
		if registryContainsDisplayName(root.key, root.path, value) {
			return true
		}
	}
	return false
}

type registryRoot struct {
	key  registry.Key
	path string
}

func uninstallRegistryRoots() []registryRoot {
	return []registryRoot{
		{key: registry.LOCAL_MACHINE, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{key: registry.LOCAL_MACHINE, path: `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{key: registry.CURRENT_USER, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{key: registry.CURRENT_USER, path: `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
	}
}

func registryContainsDisplayName(root registry.Key, path, value string) bool {
	key, err := registry.OpenKey(root, path, registry.READ)
	if err != nil {
		return false
	}
	defer key.Close()

	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return false
	}

	for _, name := range names {
		subKey, err := registry.OpenKey(key, name, registry.READ)
		if err != nil {
			continue
		}
		displayName, _, err := subKey.GetStringValue("DisplayName")
		subKey.Close()
		if err != nil {
			continue
		}
		if MatchRegistryDisplayName(displayName, value) {
			return true
		}
	}
	return false
}
