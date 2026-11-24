package providers

import (
	"fmt"
	"sync/atomic"
	"time"
)

var heartbeatInterval = 15 * time.Second

// withHeartbeat wraps log so that every line resets a quiet timer, and logs a
// notice each heartbeatInterval while nothing else is written. Silent installers
// and --no-progress downloads otherwise look hung. Call stop when the install ends.
func withHeartbeat(log func(string)) (wrapped func(string), stop func()) {
	if log == nil {
		return nil, func() {}
	}

	var last atomic.Int64
	last.Store(time.Now().UnixNano())
	wrapped = func(line string) {
		last.Store(time.Now().UnixNano())
		log(line)
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				quiet := time.Since(time.Unix(0, last.Load()))
				if quiet >= heartbeatInterval {
					log(fmt.Sprintf("[freshctl] still running, no output for %s", quiet.Round(time.Second)))
				}
			}
		}
	}()
	return wrapped, func() { close(done) }
}
