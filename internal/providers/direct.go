package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/halsatif/freshctl/internal/catalog"
	"github.com/halsatif/freshctl/internal/detection"
)

var ErrUnsupportedDirectMetadata = errors.New("unsupported direct installer metadata")

const (
	directDownloadTimeout = 30 * time.Minute
	directInstallTimeout  = 30 * time.Minute
)

type Direct struct {
	architecture    string
	makeTempDir     func(string, string) (string, error)
	removeAll       func(string) error
	downloadFile    func(context.Context, string, string, func(string)) error
	runInstaller    func(context.Context, string, []string, func(string)) error
	detectInstalled func(catalog.Application) bool
}

func NewDirect() *Direct {
	return &Direct{
		architecture:    runtime.GOARCH,
		makeTempDir:     os.MkdirTemp,
		removeAll:       os.RemoveAll,
		downloadFile:    downloadDirectInstaller,
		runInstaller:    runDirectInstaller,
		detectInstalled: detection.DetectInstalled,
	}
}

func (d *Direct) Type() catalog.ProviderType {
	return catalog.ProviderDirect
}

func (d *Direct) Validate(app catalog.Application, provider catalog.Provider) error {
	if provider.Type != catalog.ProviderDirect {
		return directMetadataError("unexpected provider type %q", provider.Type)
	}
	if provider.Strategy != catalog.InstallStrategyDirectInstaller {
		return directMetadataError("unsupported install strategy %q", provider.Strategy)
	}
	if strings.TrimSpace(provider.PackageID) == "" {
		return directMetadataError("package ID is required")
	}
	metadata := provider.Metadata.Direct
	if metadata == nil {
		return directMetadataError("metadata is required")
	}
	if metadata.InstallerType != catalog.InstallerTypeExecutable {
		return directMetadataError("installer type %q is not supported", metadata.InstallerType)
	}
	if len(metadata.Downloads) == 0 {
		return directMetadataError("at least one download is required")
	}
	if len(metadata.SilentArgs) == 0 {
		return directMetadataError("silent installer arguments are required")
	}
	if metadata.ChecksumSHA256 != "" {
		return directMetadataError("checksum verification is not implemented yet")
	}
	if metadata.Filename != "" {
		if filepath.Base(metadata.Filename) != metadata.Filename || metadata.Filename == "." {
			return directMetadataError("installer filename must not contain a path")
		}
	}
	if !detection.HasDetectionMetadata(app) {
		return directMetadataError("installed detection metadata is required")
	}

	architectures := make(map[catalog.InstallerArchitecture]bool, len(metadata.Downloads))
	for _, download := range metadata.Downloads {
		if architectures[download.Architecture] {
			return directMetadataError("duplicate download architecture %q", download.Architecture)
		}
		architectures[download.Architecture] = true
		switch download.Architecture {
		case catalog.InstallerArchitectureAny, catalog.InstallerArchitectureX64, catalog.InstallerArchitectureARM64:
		default:
			return directMetadataError("architecture %q is not supported", download.Architecture)
		}
		parsedURL, err := url.Parse(download.URL)
		if err != nil || parsedURL.Host == "" || parsedURL.Scheme != "https" {
			return directMetadataError("download URL %q is invalid", download.URL)
		}
	}
	return nil
}

func (d *Direct) Command(app catalog.Application, _ catalog.Provider) string {
	return "direct install " + app.ID
}

func (d *Direct) Install(ctx context.Context, app catalog.Application, provider catalog.Provider, opts InstallOptions) error {
	if err := d.Validate(app, provider); err != nil {
		return err
	}

	installCtx, cancel := context.WithTimeout(ctx, directInstallTimeout)
	defer cancel()

	skipped := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		select {
		case <-opts.Skip:
			skipped <- struct{}{}
			cancel()
		case <-done:
		case <-installCtx.Done():
		}
	}()

	err := d.install(installCtx, app, provider, opts)
	close(done)

	select {
	case <-skipped:
		return ErrInstallSkipped
	default:
		return err
	}
}

func (d *Direct) install(ctx context.Context, app catalog.Application, provider catalog.Provider, opts InstallOptions) error {
	metadata := provider.Metadata.Direct
	download, err := selectDirectDownload(metadata.Downloads, d.architecture)
	if err != nil {
		return err
	}

	tempDir, err := d.makeTempDir("", "freshctl-direct-*")
	if err != nil {
		return fmt.Errorf("prepare temporary directory for %s: %w", app.Name, err)
	}
	defer func() {
		if cleanupErr := d.removeAll(tempDir); cleanupErr != nil {
			logLine(opts.Log, "temporary file cleanup failed: "+cleanupErr.Error())
		}
	}()

	filename := directInstallerFilename(metadata.Filename, download.URL)
	installerPath := filepath.Join(tempDir, filename)

	logLine(opts.Log, "downloading "+app.Name)
	if err := d.downloadFile(ctx, download.URL, installerPath, opts.Log); err != nil {
		return fmt.Errorf("download failed for %s: %w", app.Name, err)
	}

	logLine(opts.Log, "installing "+app.Name)
	if err := d.runInstaller(ctx, installerPath, metadata.SilentArgs, opts.Log); err != nil {
		return fmt.Errorf("installer execution failed for %s: %w", app.Name, err)
	}

	if !d.detectInstalled(app) {
		return fmt.Errorf("installation completed but %s was not detected", app.Name)
	}
	return nil
}

func selectDirectDownload(downloads []catalog.DirectDownload, goarch string) (catalog.DirectDownload, error) {
	architecture, err := installerArchitecture(goarch)
	if err != nil {
		return catalog.DirectDownload{}, err
	}

	var generic catalog.DirectDownload
	for _, download := range downloads {
		if download.Architecture == architecture {
			return download, nil
		}
		if download.Architecture == catalog.InstallerArchitectureAny {
			generic = download
		}
	}
	if generic.URL != "" {
		return generic, nil
	}
	return catalog.DirectDownload{}, fmt.Errorf("no direct installer is available for %s", architecture)
}

func installerArchitecture(goarch string) (catalog.InstallerArchitecture, error) {
	switch goarch {
	case "amd64":
		return catalog.InstallerArchitectureX64, nil
	case "arm64":
		return catalog.InstallerArchitectureARM64, nil
	default:
		return "", fmt.Errorf("direct installers do not support architecture %s", goarch)
	}
}

func directInstallerFilename(filename, downloadURL string) string {
	if filename != "" {
		return filename
	}
	parsedURL, err := url.Parse(downloadURL)
	if err == nil {
		candidate := filepath.Base(parsedURL.Path)
		if candidate != "." && candidate != "/" && strings.HasSuffix(strings.ToLower(candidate), ".exe") {
			return candidate
		}
	}
	return "installer.exe"
}

func directMetadataError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedDirectMetadata, fmt.Sprintf(format, args...))
}

func downloadDirectInstaller(ctx context.Context, downloadURL, destination string, log func(string)) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: directDownloadTimeout}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("server returned %s", response.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	reporter := &downloadProgress{
		total: response.ContentLength,
		log:   log,
		next:  10,
	}
	if _, err := io.Copy(file, io.TeeReader(response.Body, reporter)); err != nil {
		return err
	}
	logLine(log, "download complete")
	return nil
}

func runDirectInstaller(ctx context.Context, path string, args []string, _ func(string)) error {
	cmd := exec.CommandContext(ctx, path, args...)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("installer exited with code %d", exitErr.ExitCode())
		}
		return err
	}
	return nil
}

type downloadProgress struct {
	total   int64
	written int64
	next    int64
	log     func(string)
}

func (p *downloadProgress) Write(data []byte) (int, error) {
	written := len(data)
	p.written += int64(written)
	if p.total <= 0 {
		return written, nil
	}

	percent := p.written * 100 / p.total
	for percent >= p.next && p.next <= 100 {
		logLine(p.log, fmt.Sprintf("downloading: %d%%", p.next))
		p.next += 10
	}
	return written, nil
}

func logLine(log func(string), line string) {
	if log != nil {
		log(line)
	}
}
