package debuglog

import "sync"

// Entry is a structured debug log record passed to an optional host logger.
type Entry struct {
	Event  string
	Labels map[string]string
	Fields map[string]any
}

var (
	mu     sync.RWMutex
	logFn  func(Entry)
)

// SetLogger registers a host-provided debug logger. Nil clears the hook.
func SetLogger(fn func(Entry)) {
	mu.Lock()
	defer mu.Unlock()
	logFn = fn
}

// Log invokes the registered logger when set. Safe to call from any goroutine.
func Log(e Entry) {
	mu.RLock()
	fn := logFn
	mu.RUnlock()
	if fn == nil {
		return
	}
	if e.Labels == nil {
		e.Labels = map[string]string{}
	}
	fn(e)
}