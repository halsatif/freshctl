//go:build !windows

package detection

func DetectRegistry(value string) bool {
	return false
}
