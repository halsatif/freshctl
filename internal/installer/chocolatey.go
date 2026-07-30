package installer

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

	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/providers"
	"golang.org/x/sys/windows"
)

var ErrPackageManagerMissing = errors.New("chocolatey was not found")
var ErrBrokenPackageManagerInstall = errors.New("broken Chocolatey installation detected")
var ErrInstallSkipped = providers.ErrInstallSkipped

const (
	chocolateyDir = `C:\ProgramData\chocolatey`
	chocolateyExe = `C:\ProgramData\chocolatey\bin\choco.exe`
)

type EventKind int

const (
	EventLog EventKind = iota
	EventAppStarted
	EventAppFinished
	EventSummary
)

type Event struct {
	Kind    EventKind
	App     catalog.Application
	Line    string
	Success bool
	Err     error
	Results []Result
}

type Result struct {
	App     catalog.Application
	Success bool
	Skipped bool
	Err     error
}

type InstallError = providers.InstallError

type BootstrapEventKind int

const (
	BootstrapLog BootstrapEventKind = iota
	BootstrapFinished
)

type BootstrapEvent struct {
	Kind  BootstrapEventKind
	Line  string
	Ready bool
	Err   error
}

func CommandFor(app catalog.Application) string {
	provider, providerInstaller, err := resolveApplicationProvider(app)
	if err != nil {
		return err.Error()
	}
	if commandProvider, ok := providerInstaller.(providers.CommandProvider); ok {
		return commandProvider.Command(app, provider)
	}
	return fmt.Sprintf("%s install %s", provider.Type, provider.PackageID)
}

func HasPackageManager() bool {
	if HasBrokenPackageManagerInstall() {
		return false
	}
	return providers.ChocolateyPath() != ""
}

func HasBrokenPackageManagerInstall() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	info, err := os.Stat(chocolateyDir)
	if err != nil || !info.IsDir() {
		return false
	}

	chocoInfo, err := os.Stat(chocolateyExe)
	return err != nil || chocoInfo.IsDir()
}

func RemoveBrokenPackageManagerInstall() error {
	if runtime.GOOS != "windows" {
		return errors.New("broken Chocolatey cleanup is only available on Windows")
	}
	if !HasBrokenPackageManagerInstall() {
		return nil
	}
	return os.RemoveAll(chocolateyDir)
}

func IsElevated() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	return windows.GetCurrentProcessToken().IsElevated()
}

func RelaunchElevated(args []string) error {
	if runtime.GOOS != "windows" {
		return errors.New("administrator relaunch is only available on Windows")
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	script := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", escapePowerShellSingleQuoted(exe))
	if len(args) > 0 {
		script += " -ArgumentList " + powerShellArray(args)
	}
	return exec.Command("powershell.exe", "-NoProfile", "-Command", script).Run()
}

func InstallApps(ctx context.Context, apps []catalog.Application, events chan<- Event, skips <-chan struct{}) {
	defer close(events)

	if len(apps) == 0 {
		events <- Event{Kind: EventLog, Line: "No apps selected. Go back and choose at least one app."}
		events <- Event{Kind: EventSummary}
		return
	}

	type resolvedProvider struct {
		metadata  catalog.Provider
		installer providers.Installer
		err       error
	}

	resolvedProviders := make(map[string]resolvedProvider, len(apps))
	needsChocolatey := false
	for _, app := range apps {
		provider, providerInstaller, err := resolveApplicationProvider(app)
		resolvedProviders[app.ID] = resolvedProvider{
			metadata:  provider,
			installer: providerInstaller,
			err:       err,
		}
		if err != nil {
			continue
		}
		if provider.Type == catalog.ProviderChocolatey {
			needsChocolatey = true
		}
	}

	if needsChocolatey {
		if HasBrokenPackageManagerInstall() {
			events <- Event{Kind: EventLog, Line: ErrBrokenPackageManagerInstall.Error()}
			events <- Event{Kind: EventSummary}
			return
		}
		if !HasPackageManager() {
			events <- Event{Kind: EventLog, Line: ErrPackageManagerMissing.Error()}
			events <- Event{Kind: EventSummary}
			return
		}
	}

	results := make([]Result, 0, len(apps))
	for _, app := range apps {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			results = append(results, Result{App: app, Err: err})
			events <- Event{Kind: EventAppFinished, App: app, Err: err}
			events <- Event{Kind: EventSummary, Results: results}
			return
		default:
		}

		drainSkipRequests(skips)
		resolved, ok := resolvedProviders[app.ID]
		if !ok || resolved.err != nil {
			err := resolved.err
			if err == nil {
				err = fmt.Errorf("provider resolution failed for %s", app.Name)
			}
			results = append(results, Result{App: app, Err: err})
			events <- Event{Kind: EventAppStarted, App: app, Line: err.Error()}
			events <- Event{Kind: EventAppFinished, App: app, Err: err}
			continue
		}

		events <- Event{Kind: EventAppStarted, App: app, Line: CommandFor(app)}
		err := resolved.installer.Install(ctx, app, resolved.metadata, providers.InstallOptions{
			Log: func(line string) {
				events <- Event{Kind: EventLog, App: app, Line: line}
			},
			Skip: skips,
		})
		result := Result{App: app, Success: err == nil, Err: err}
		if errors.Is(err, ErrInstallSkipped) {
			result.Skipped = true
		}
		results = append(results, result)
		events <- Event{Kind: EventAppFinished, App: app, Success: result.Success, Err: result.Err}
	}

	events <- Event{Kind: EventSummary, Results: results}
}

func resolveApplicationProvider(app catalog.Application) (catalog.Provider, providers.Installer, error) {
	provider, ok := app.PrimaryProvider()
	if !ok {
		return catalog.Provider{}, nil, fmt.Errorf("no install provider configured for %s", app.Name)
	}
	providerInstaller, ok := providers.Get(provider.Type)
	if !ok {
		return provider, nil, fmt.Errorf("unknown package provider: %s", provider.Type)
	}
	if err := providerInstaller.Validate(app, provider); err != nil {
		return provider, providerInstaller, fmt.Errorf("invalid %s provider for %s: %w", provider.Type, app.Name, err)
	}
	return provider, providerInstaller, nil
}

func BootstrapPackageManager(ctx context.Context, events chan<- BootstrapEvent) {
	defer close(events)

	if runtime.GOOS != "windows" {
		events <- BootstrapEvent{
			Kind: BootstrapFinished,
			Err:  errors.New("chocolatey bootstrap is only available on Windows"),
		}
		return
	}

	if HasBrokenPackageManagerInstall() {
		events <- BootstrapEvent{Kind: BootstrapFinished, Err: ErrBrokenPackageManagerInstall}
		return
	}
	if HasPackageManager() {
		events <- BootstrapEvent{Kind: BootstrapFinished, Ready: true}
		return
	}

	cmd := exec.CommandContext(
		ctx,
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", chocolateyBootstrapScript(),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		events <- BootstrapEvent{Kind: BootstrapFinished, Err: err}
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		events <- BootstrapEvent{Kind: BootstrapFinished, Err: err}
		return
	}
	if err := cmd.Start(); err != nil {
		events <- BootstrapEvent{Kind: BootstrapFinished, Err: err}
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scanBootstrapOutput(&wg, stdout, events)
	go scanBootstrapOutput(&wg, stderr, events)
	wg.Wait()

	err = cmd.Wait()
	events <- BootstrapEvent{
		Kind:  BootstrapFinished,
		Ready: hasDirectPackageManager(),
		Err:   err,
	}
}

func hasDirectPackageManager() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	info, err := os.Stat(chocolateyExe)
	return err == nil && !info.IsDir()
}

func chocolateyBootstrapScript() string {
	return strings.Join([]string{
		"Set-ExecutionPolicy Bypass -Scope Process -Force",
		"[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072",
		"iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))",
	}, "; ")
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func powerShellArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "'"+escapePowerShellSingleQuoted(value)+"'")
	}
	return "@(" + strings.Join(quoted, ",") + ")"
}

func scanBootstrapOutput(wg *sync.WaitGroup, r io.Reader, events chan<- BootstrapEvent) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			events <- BootstrapEvent{Kind: BootstrapLog, Line: line}
		}
	}
}

func drainSkipRequests(skips <-chan struct{}) {
	for {
		select {
		case <-skips:
		default:
			return
		}
	}
}
