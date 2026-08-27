//go:build !windows

package providers

import "os"

func terminateProcessTree(process *os.Process) {
	if process != nil {
		_ = process.Kill()
	}
}
