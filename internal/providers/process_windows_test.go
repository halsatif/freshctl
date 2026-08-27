//go:build windows

package providers

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestWaitCommandCancelsProcessTreePromptly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.Command(os.Args[0], "-test.run=TestProcessTreeHelper", "--")
	cmd.Env = append(os.Environ(), "FRESHCTL_PROCESS_HELPER=parent")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	ready := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if scanner.Text() == "ready" {
				close(ready)
				return
			}
		}
	}()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("helper process did not start its child")
	}

	started := time.Now()
	cancel()
	err = waitCommand(ctx, cmd)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("process tree cancellation took too long: %s", elapsed)
	}
}

func TestProcessTreeHelper(t *testing.T) {
	switch os.Getenv("FRESHCTL_PROCESS_HELPER") {
	case "parent":
		child := exec.Command(os.Args[0], "-test.run=TestProcessTreeHelper", "--")
		child.Env = append(os.Environ(), "FRESHCTL_PROCESS_HELPER=child")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		fmt.Println("ready")
		_ = child.Wait()
		os.Exit(0)
	case "child":
		time.Sleep(2 * time.Minute)
		os.Exit(0)
	}
}
