// Package logfile appends install logs to %LOCALAPPDATA%\freshctl\logs, one file per run.
package logfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu   sync.Mutex
	file *os.File
	dead bool // set after the first open failure so we do not retry on every line
)

// Write appends "2006-01-02 15:04:05.000 [source] line". Failures are ignored:
// logging must never break an install.
func Write(source, line string) {
	mu.Lock()
	defer mu.Unlock()
	if dead {
		return
	}
	if file == nil && !open() {
		dead = true
		return
	}
	fmt.Fprintf(file, "%s [%s] %s\r\n", time.Now().Format("2006-01-02 15:04:05.000"), source, line)
}

// Path returns the current log file path, or "" if nothing was written yet.
func Path() string {
	mu.Lock()
	defer mu.Unlock()
	if file == nil {
		return ""
	}
	return file.Name()
}

// Close releases the file; the next Write starts a new one.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
		file = nil
	}
	dead = false
}

func open() bool {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return false
	}
	dir := filepath.Join(base, "freshctl", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	name := "freshctl-" + time.Now().Format("2006-01-02_15-04-05") + ".log"
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	file = f
	return true
}
