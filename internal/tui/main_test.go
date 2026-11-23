package tui

import (
	"os"
	"testing"

	"github.com/halsatif/freshctl/internal/logfile"
)

// Keep tests from writing into the real %LOCALAPPDATA%\freshctl\logs.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "freshctl-tui-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("LOCALAPPDATA", dir)
	code := m.Run()
	logfile.Close()
	os.RemoveAll(dir)
	os.Exit(code)
}
