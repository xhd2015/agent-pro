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
// A zero or negative timeout waits indefinitely.
func WaitUntilWritable(check CheckWritableFunc, listenAddr, sessionID string, timeout time.Duration) WritableStatus {
	var last WritableStatus
	poll := func() WritableStatus {
		text, err := SnapshotText(listenAddr, sessionID)
		if err == nil && check != nil {
			last = check([]byte(text))
		}
		return last
	}
	if timeout <= 0 {
		for {
			if poll().Ready {
				return last
			}
			time.Sleep(150 * time.Millisecond)
		}
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if poll().Ready {
			return last
		}
		time.Sleep(150 * time.Millisecond)
	}
	if last.Reason == "" {
		last.Reason = "timed out waiting for writable prompt"
	}
	return last
}