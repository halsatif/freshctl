package providers

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWithHeartbeatReportsQuietPeriodsUntilStopped(t *testing.T) {
	old := heartbeatInterval
	heartbeatInterval = 20 * time.Millisecond
	t.Cleanup(func() { heartbeatInterval = old })

	var mu sync.Mutex
	var lines []string
	log := func(line string) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, line)
	}
	count := func() int {
		mu.Lock()
		defer mu.Unlock()
		return len(lines)
	}

	wrapped, stop := withHeartbeat(log)
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	first := ""
	if len(lines) > 0 {
		first = lines[0]
	}
	mu.Unlock()
	if !strings.Contains(first, "still running, no output for") {
		t.Fatalf("expected a heartbeat line, got %q", first)
	}

	stop()
	time.Sleep(50 * time.Millisecond) // let a tick already in flight finish
	n := count()
	time.Sleep(100 * time.Millisecond)
	if count() != n {
		t.Fatal("heartbeat kept logging after stop")
	}

	wrapped("real output") // wrapped log still forwards lines
	if count() != n+1 {
		t.Fatal("wrapped log did not forward the line")
	}
}

func TestWithHeartbeatNilLog(t *testing.T) {
	wrapped, stop := withHeartbeat(nil)
	if wrapped != nil {
		t.Fatal("nil log should stay nil")
	}
	stop()
}
