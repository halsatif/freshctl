//go:build windows

package providers

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func terminateProcessTree(process *os.Process) {
	if process == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "taskkill.exe", "/PID", strconv.Itoa(process.Pid), "/T", "/F")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		_ = process.Kill()
	}
}
