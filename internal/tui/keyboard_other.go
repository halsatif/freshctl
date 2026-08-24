//go:build !windows

package tui

func platformControlKeyPressed() bool {
	return false
}
