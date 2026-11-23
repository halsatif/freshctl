package logfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestWriteCreatesTimestampedFile(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	t.Cleanup(Close) // Windows cannot delete an open file

	Write("Git", "first line")
	Write("Git", "second line")

	path := Path()
	if !strings.HasPrefix(path, filepath.Join(os.Getenv("LOCALAPPDATA"), "freshctl", "logs")) {
		t.Fatalf("unexpected log path %q", path)
	}
	if ok, _ := regexp.MatchString(`freshctl-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.log$`, path); !ok {
		t.Fatalf("unexpected log file name %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	line := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} \[Git\] (first|second) line$`)
	lines := strings.Split(strings.TrimSpace(string(data)), "\r\n")
	if len(lines) != 2 || !line.MatchString(lines[0]) || !line.MatchString(lines[1]) {
		t.Fatalf("unexpected log content %q", data)
	}
}
