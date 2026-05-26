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
	"time"

	"github.com/halsatif/freshctl/internal/catalog"
	"golang.org/x/sys/windows"
)

var ErrPackageManagerMissing = errors.New("chocolatey was not found")
var ErrBrokenPackageManagerInstall = errors.New("broken Chocolatey installation detected")
var ErrInstallSkipped = errors.New("install skipped")

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
	App     catalog.App
	Line    string
	Success bool
	Err     error
	Results []Result
}

type Result struct {
	App     catalog.App
	Success bool
	Skipped bool
	Err     error
}

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

func CommandFor(app catalog.App) string {
	return fmt.Sprintf("choco install %s -y --no-progress", app.ID)
}

func HasPackageManager() bool {
	if HasBrokenPackageManagerInstall() {
		return false
	}
	return chocoPath() != ""
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

func InstallApps(ctx context.Context, apps []catalog.App, events chan<- Event, skips <-chan struct{}) {
	defer close(events)

	if len(apps) == 0 {
		events <- Event{Kind: EventLog, Line: "No apps selected. Go back and choose at least one app."}
		events <- Event{Kind: EventSummary}
		return
	}

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
		events <- Event{Kind: EventAppStarted, App: app, Line: CommandFor(app)}
		err := installOneWithSkip(ctx, app, events, skips)
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

func installOneWithSkip(ctx context.Context, app catalog.App, events chan<- Event, skips <-chan struct{}) error {
	appCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	skipped := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		select {
		case <-skips:
			skipped <- struct{}{}
			cancel()
		case <-done:
		case <-appCtx.Done():
		}
	}()

	err := installOne(appCtx, app, events)
	close(done)

	select {
	case <-skipped:
		return ErrInstallSkipped
	default:
		return err
	}
}

func installOne(ctx context.Context, app catalog.App, events chan<- Event) error {
	cmd := exec.CommandContext(ctx, chocoPath(), "install", app.ID, "-y", "--no-progress")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go scanOutput(&wg, stdout, app, events)
	go scanOutput(&wg, stderr, app, events)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	return nil
}

func chocoPath() string {
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

func scanOutput(wg *sync.WaitGroup, r io.Reader, app catalog.App, events chan<- Event) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			events <- Event{Kind: EventLog, App: app, Line: line}
		}
	}
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
