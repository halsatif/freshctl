package providers

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/halsatif/freshctl/internal/catalog"
)

var ErrInstallSkipped = errors.New("install skipped")

const (
	chocolateyExe      = `C:\ProgramData\chocolatey\bin\choco.exe`
	installTimeout     = 30 * time.Minute
	defaultBufferLines = 64 * 1024
)

type Chocolatey struct{}

type InstallError struct {
	App      catalog.Application
	Provider catalog.Provider
	ExitCode int
	Err      error
}

func (e InstallError) Error() string {
	if e.ExitCode >= 0 {
		return fmt.Sprintf("Chocolatey failed for %s (%s) with exit code %d", e.App.Name, e.Provider.PackageID, e.ExitCode)
	}
	return fmt.Sprintf("Chocolatey failed for %s (%s): %v", e.App.Name, e.Provider.PackageID, e.Err)
}

func (e InstallError) Unwrap() error {
	return e.Err
}

func (c *Chocolatey) Type() catalog.ProviderType {
	return catalog.ProviderChocolatey
}

func (c *Chocolatey) Validate(_ catalog.Application, provider catalog.Provider) error {
	if provider.Type != catalog.ProviderChocolatey {
		return fmt.Errorf("Chocolatey provider has unexpected type %q", provider.Type)
	}
	if provider.Strategy != catalog.InstallStrategyPackageManager {
		return fmt.Errorf("Chocolatey provider has unsupported install strategy %q", provider.Strategy)
	}
	if strings.TrimSpace(provider.PackageID) == "" {
		return errors.New("Chocolatey provider package ID is required")
	}
	if provider.Metadata.Direct != nil {
		return errors.New("Chocolatey provider contains unsupported direct installer metadata")
	}
	return nil
}

func (c *Chocolatey) Command(_ catalog.Application, provider catalog.Provider) string {
	return "choco " + strings.Join(installArgs(provider), " ")
}

func (c *Chocolatey) Install(ctx context.Context, app catalog.Application, provider catalog.Provider, opts InstallOptions) error {
	if err := c.Validate(app, provider); err != nil {
		return err
	}

	appCtx, cancel := context.WithTimeout(ctx, installTimeout)
	defer cancel()

	skipped := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		select {
		case <-opts.Skip:
			skipped <- struct{}{}
			cancel()
		case <-done:
		case <-appCtx.Done():
		}
	}()

	err := c.install(appCtx, app, provider, opts)
	close(done)

	select {
	case <-skipped:
		return ErrInstallSkipped
	default:
		return err
	}
}

func (c *Chocolatey) install(ctx context.Context, app catalog.Application, provider catalog.Provider, opts InstallOptions) error {
	choco := ChocolateyPath()
	if choco == "" {
		return InstallError{App: app, Provider: provider, ExitCode: -1, Err: errors.New("chocolatey executable was not found")}
	}

	cmd := exec.CommandContext(ctx, choco, installArgs(provider)...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return InstallError{App: app, Provider: provider, ExitCode: -1, Err: err}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return InstallError{App: app, Provider: provider, ExitCode: -1, Err: err}
	}
	if err := cmd.Start(); err != nil {
		return InstallError{App: app, Provider: provider, ExitCode: -1, Err: err}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scanOutput(&wg, stdout, opts.Log)
	go scanOutput(&wg, stderr, opts.Log)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		exitCode := -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return InstallError{App: app, Provider: provider, ExitCode: exitCode, Err: err}
	}
	return nil
}

func installArgs(provider catalog.Provider) []string {
	args := []string{"install", provider.PackageID, "-y", "--no-progress"}
	if provider.Metadata.Prerelease {
		args = append(args, "--pre")
	}
	return args
}

func ChocolateyPath() string {
	if path, err := exec.LookPath("choco"); err == nil {
		return path
	}
	if runtime.GOOS == "windows" {
		if info, err := os.Stat(chocolateyExe); err == nil && !info.IsDir() {
			return chocolateyExe
		}
	}
	return ""
}

func scanOutput(wg *sync.WaitGroup, r io.Reader, log func(string)) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), defaultBufferLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && log != nil {
			log(line)
		}
	}
}
