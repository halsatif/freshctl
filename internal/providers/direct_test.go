package providers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halsatif/freshctl/internal/catalog"
)

func TestDirectProviderIsRegistered(t *testing.T) {
	installer, ok := Get(catalog.ProviderDirect)
	if !ok {
		t.Fatal("Direct provider should be registered")
	}
	if installer.Type() != catalog.ProviderDirect {
		t.Fatalf("expected Direct provider type, got %q", installer.Type())
	}
}

func TestDirectMetadataValidation(t *testing.T) {
	direct := NewDirect()
	app, provider := validDirectTestMetadata()

	if err := direct.Validate(app, provider); err != nil {
		t.Fatalf("expected valid Direct metadata, got %v", err)
	}
}

func TestDirectRejectsUnsupportedMetadata(t *testing.T) {
	app, provider := validDirectTestMetadata()
	tests := []struct {
		name      string
		mutateApp func(*catalog.Application)
		mutate    func(*catalog.Provider)
	}{
		{
			name: "missing metadata",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct = nil
			},
		},
		{
			name: "unsupported installer type",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.InstallerType = "Archive"
			},
		},
		{
			name: "missing filename",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.Filename = ""
			},
		},
		{
			name: "filename extension mismatch",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.Filename = "TestSetup.msi"
			},
		},
		{
			name: "unsupported architecture",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.Downloads[0].Architecture = "x86"
			},
		},
		{
			name: "insecure download URL",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.Downloads[0].URL = "http://example.test/setup.exe"
			},
		},
		{
			name: "checksum before verification support",
			mutate: func(provider *catalog.Provider) {
				provider.Metadata.Direct.ChecksumSHA256 = strings.Repeat("a", 64)
			},
		},
		{
			name: "missing detection metadata",
			mutateApp: func(app *catalog.Application) {
				app.DetectMethod = catalog.DetectNone
				app.DetectValue = ""
			},
			mutate: func(*catalog.Provider) {},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidateApp := app
			candidate := cloneProvider(provider)
			if test.mutateApp != nil {
				test.mutateApp(&candidateApp)
			}
			test.mutate(&candidate)
			err := NewDirect().Validate(candidateApp, candidate)
			if !errors.Is(err, ErrUnsupportedDirectMetadata) {
				t.Fatalf("expected unsupported metadata error, got %v", err)
			}
		})
	}
}

func TestDirectInstallerErrorPropagation(t *testing.T) {
	app, provider := validDirectTestMetadata()
	runnerErr := errors.New("runner failed")
	direct := testDirectProvider(t)
	direct.downloadFile = func(context.Context, string, string, func(string)) error {
		return nil
	}
	direct.runInstaller = func(context.Context, catalog.InstallerType, string, []string, func(string)) error {
		return runnerErr
	}
	direct.detectInstalled = func(catalog.Application) bool {
		t.Fatal("detection should not run after installer failure")
		return false
	}

	err := direct.Install(context.Background(), app, provider, InstallOptions{})
	if !errors.Is(err, runnerErr) {
		t.Fatalf("expected installer error to be preserved, got %v", err)
	}
	if !strings.Contains(err.Error(), "installer execution failed for Test App") {
		t.Fatalf("expected readable installer error, got %v", err)
	}
}

func TestDirectDownloadErrorIsReadable(t *testing.T) {
	app, provider := validDirectTestMetadata()
	downloadErr := errors.New("network unavailable")
	direct := testDirectProvider(t)
	direct.downloadFile = func(context.Context, string, string, func(string)) error {
		return downloadErr
	}
	direct.runInstaller = func(context.Context, catalog.InstallerType, string, []string, func(string)) error {
		t.Fatal("installer should not run after download failure")
		return nil
	}

	err := direct.Install(context.Background(), app, provider, InstallOptions{})
	if !errors.Is(err, downloadErr) {
		t.Fatalf("expected download error to be preserved, got %v", err)
	}
	if !strings.Contains(err.Error(), "download failed for Test App") {
		t.Fatalf("expected readable download error, got %v", err)
	}
}

func TestDirectDetectionFailureAfterSuccessfulInstaller(t *testing.T) {
	app, provider := validDirectTestMetadata()
	direct := testDirectProvider(t)
	direct.downloadFile = func(context.Context, string, string, func(string)) error {
		return nil
	}
	direct.runInstaller = func(context.Context, catalog.InstallerType, string, []string, func(string)) error {
		return nil
	}
	direct.detectInstalled = func(catalog.Application) bool {
		return false
	}

	err := direct.Install(context.Background(), app, provider, InstallOptions{})
	if err == nil || err.Error() != "installation completed but Test App was not detected" {
		t.Fatalf("expected detection failure, got %v", err)
	}
}

func TestDirectUsesArchitectureDownloadAndSilentArguments(t *testing.T) {
	app, provider := validDirectTestMetadata()
	direct := testDirectProvider(t)
	direct.architecture = "arm64"

	var downloadedURL string
	direct.downloadFile = func(_ context.Context, downloadURL, _ string, _ func(string)) error {
		downloadedURL = downloadURL
		return nil
	}
	var gotArgs []string
	var gotInstallerType catalog.InstallerType
	direct.runInstaller = func(_ context.Context, installerType catalog.InstallerType, _ string, args []string, _ func(string)) error {
		gotInstallerType = installerType
		gotArgs = append([]string{}, args...)
		return nil
	}
	direct.detectInstalled = func(catalog.Application) bool {
		return true
	}

	if err := direct.Install(context.Background(), app, provider, InstallOptions{}); err != nil {
		t.Fatal(err)
	}
	if downloadedURL != "https://example.test/arm64.exe" {
		t.Fatalf("expected arm64 download, got %q", downloadedURL)
	}
	if strings.Join(gotArgs, " ") != "/VERYSILENT /NORESTART" {
		t.Fatalf("unexpected silent arguments %q", gotArgs)
	}
	if gotInstallerType != catalog.InstallerTypeExecutable {
		t.Fatalf("unexpected installer type %q", gotInstallerType)
	}
}

func TestDirectInstallerCommandSupportsExecutableAndMSI(t *testing.T) {
	tests := []struct {
		name          string
		installerType catalog.InstallerType
		path          string
		args          []string
		wantCommand   string
		wantArgs      string
	}{
		{
			name:          "executable",
			installerType: catalog.InstallerTypeExecutable,
			path:          `C:\\Temp\\setup.exe`,
			args:          []string{"/S"},
			wantCommand:   `C:\\Temp\\setup.exe`,
			wantArgs:      "/S",
		},
		{
			name:          "msi",
			installerType: catalog.InstallerTypeMSI,
			path:          `C:\\Temp\\setup.msi`,
			args:          []string{"/qn", "/norestart"},
			wantCommand:   "msiexec.exe",
			wantArgs:      `/i C:\\Temp\\setup.msi /qn /norestart`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, args, err := directInstallerCommand(test.installerType, test.path, test.args)
			if err != nil {
				t.Fatal(err)
			}
			if command != test.wantCommand || strings.Join(args, " ") != test.wantArgs {
				t.Fatalf("got %q %q, want %q %q", command, args, test.wantCommand, test.wantArgs)
			}
		})
	}
}

func TestDirectMSIMetadataValidation(t *testing.T) {
	direct := NewDirect()
	app, provider := validDirectTestMetadata()
	provider.Metadata.Direct.InstallerType = catalog.InstallerTypeMSI
	provider.Metadata.Direct.Filename = "TestSetup.msi"

	if err := direct.Validate(app, provider); err != nil {
		t.Fatalf("expected valid MSI metadata, got %v", err)
	}
}

func TestDirectCleansTemporaryDirectoryAfterInstall(t *testing.T) {
	app, provider := validDirectTestMetadata()
	direct := testDirectProvider(t)
	var tempDir string
	direct.makeTempDir = func(_ string, _ string) (string, error) {
		tempDir = filepath.Join(t.TempDir(), "direct")
		return tempDir, os.MkdirAll(tempDir, 0o755)
	}
	direct.downloadFile = func(_ context.Context, _ string, destination string, _ func(string)) error {
		return os.WriteFile(destination, []byte("installer"), 0o600)
	}
	direct.runInstaller = func(context.Context, catalog.InstallerType, string, []string, func(string)) error {
		return nil
	}
	direct.detectInstalled = func(catalog.Application) bool { return true }

	if err := direct.Install(context.Background(), app, provider, InstallOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Fatalf("temporary directory should be removed, stat error: %v", err)
	}
}

func TestDownloadDirectInstallerReportsProgress(t *testing.T) {
	payload := strings.Repeat("x", 100)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Length", "100")
		_, _ = writer.Write([]byte(payload))
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "installer.exe")
	var logs []string
	if err := downloadDirectInstaller(context.Background(), server.URL, destination, func(line string) {
		logs = append(logs, line)
	}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != payload {
		t.Fatal("downloaded installer content does not match response")
	}
	if !containsLine(logs, "downloading: 100%") || !containsLine(logs, "download complete") {
		t.Fatalf("expected download progress and completion logs, got %#v", logs)
	}
}

func validDirectTestMetadata() (catalog.Application, catalog.Provider) {
	app := catalog.Application{
		ID:           "test-app",
		Name:         "Test App",
		DetectMethod: catalog.DetectRegistry,
		DetectValue:  "Test App",
	}
	provider := catalog.Provider{
		Type:      catalog.ProviderDirect,
		PackageID: "test-app-direct",
		Strategy:  catalog.InstallStrategyDirectInstaller,
		Metadata: catalog.ProviderMetadata{Direct: &catalog.DirectInstallerMetadata{
			Downloads: []catalog.DirectDownload{
				{URL: "https://example.test/x64.exe", Architecture: catalog.InstallerArchitectureX64},
				{URL: "https://example.test/arm64.exe", Architecture: catalog.InstallerArchitectureARM64},
			},
			Filename:      "TestSetup.exe",
			SilentArgs:    []string{"/VERYSILENT", "/NORESTART"},
			InstallerType: catalog.InstallerTypeExecutable,
		}},
	}
	app.Providers = []catalog.Provider{provider}
	return app, provider
}

func cloneProvider(provider catalog.Provider) catalog.Provider {
	cloned := provider
	if provider.Metadata.Direct != nil {
		direct := *provider.Metadata.Direct
		direct.Downloads = append([]catalog.DirectDownload{}, provider.Metadata.Direct.Downloads...)
		direct.SilentArgs = append([]string{}, provider.Metadata.Direct.SilentArgs...)
		cloned.Metadata.Direct = &direct
	}
	return cloned
}

func testDirectProvider(t *testing.T) *Direct {
	t.Helper()
	root := t.TempDir()
	direct := NewDirect()
	direct.makeTempDir = func(_ string, pattern string) (string, error) {
		return os.MkdirTemp(root, pattern)
	}
	return direct
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}
