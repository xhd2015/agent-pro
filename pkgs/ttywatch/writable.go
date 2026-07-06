package ttywatch

import "time"

// WritableStatus reports whether a TTY session is ready to receive injected input.
type WritableStatus struct {
	Ready  bool
	Reason string
	State  string
}

// CheckWritableFunc inspects scrollback bytes for prompt readiness.
type CheckWritableFunc func(scrollback []byte) WritableStatus

// WaitUntilWritable polls check until ready or timeout.
func WaitUntilWritable(check CheckWritableFunc, listenAddr, sessionID string, timeout time.Duration) WritableStatus {
	deadline := time.Now().Add(timeout)
	var last WritableStatus
	for time.Now().Before(deadline) {
		text, err := SnapshotText(listenAddr, sessionID)
		if err == nil && check != nil {
			last = check([]byte(text))
		}
		if last.Ready {
			return last
		}
		time.Sleep(150 * time.Millisecond)
	}
	if last.Reason == "" {
		last.Reason = "timed out waiting for writable prompt"
	}
	return last
}