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
	"github.com/halsatif/freshctl/internal/sources"
	"golang.org/x/sys/windows"
)

var ErrPackageManagerMissing = errors.New("chocolatey was not found")
var ErrBrokenPackageManagerInstall = errors.New("broken Chocolatey installation detected")
var ErrInstallSkipped = sources.ErrInstallSkipped

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
	App     catalog.Package
	Line    string
	Success bool
	Err     error
	Results []Result
}

type Result struct {
	App     catalog.Package
	Success bool
	Skipped bool
	Err     error
}

type InstallError = sources.InstallError

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

func CommandFor(app catalog.Package) string {
	source, ok := sources.GetSource(string(app.Source))
	if !ok {
		return fmt.Sprintf("unknown package source: %s", app.Source)
	}
	if commandSource, ok := source.(sources.CommandSource); ok {
		return commandSource.Command(app)
	}
	return fmt.Sprintf("%s install %s", app.Source, app.PackageID)
}

func HasPackageManager() bool {
	if HasBrokenPackageManagerInstall() {
		return false
	}
	return sources.ChocolateyPath() != ""
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

func InstallApps(ctx context.Context, apps []catalog.Package, events chan<- Event, skips <-chan struct{}) {
	defer close(events)

	if len(apps) == 0 {
		events <- Event{Kind: EventLog, Line: "No apps selected. Go back and choose at least one app."}
		events <- Event{Kind: EventSummary}
		return
	}

	resolvedSources := make(map[string]sources.Source, len(apps))
	needsChocolatey := false
	for _, app := range apps {
		source, ok := sources.GetSource(string(app.Source))
		if !ok {
			continue
		}
		resolvedSources[app.PackageID] = source
		if source.ID() == string(catalog.PackageSourceChocolatey) {
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
		source, ok := resolvedSources[app.PackageID]
		if !ok {
			err := fmt.Errorf("unknown package source: %s", app.Source)
			results = append(results, Result{App: app, Err: err})
			events <- Event{Kind: EventAppStarted, App: app, Line: err.Error()}
			events <- Event{Kind: EventAppFinished, App: app, Err: err}
			continue
		}

		events <- Event{Kind: EventAppStarted, App: app, Line: CommandFor(app)}
		err := source.Install(ctx, app, sources.InstallOptions{
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
