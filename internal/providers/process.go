package providers

import (
	"context"
	"os/exec"
)

func waitCommand(ctx context.Context, cmd *exec.Cmd) error {
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		select {
		case err := <-done:
			return err
		default:
		}

		terminateProcessTree(cmd.Process)
		<-done
		return ctx.Err()
	}
}
