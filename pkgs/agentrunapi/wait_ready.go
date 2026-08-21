package agentrunapi

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultReadyTimeout      = 60 * time.Second
	defaultReadyPollInterval = 500 * time.Millisecond
)

// WaitReadyOpts polls an injectable status source until the session is ready.
type WaitReadyOpts struct {
	SessionID    string                         // required (non-empty after trim)
	StatusFn     func() (stdout string, err error) // required
	Timeout      time.Duration                  // 0 → 60s
	PollInterval time.Duration                  // 0 → 500ms
}

// WaitReady polls StatusFn until IsSessionReadyFromStatus or timeout.
// Timeout error mentions ready and/or timeout. No agent-run binary LookPath.
func WaitReady(opts WaitReadyOpts) error {
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if opts.StatusFn == nil {
		return fmt.Errorf("StatusFn is required")
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultReadyTimeout
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultReadyPollInterval
	}

	deadline := time.Now().Add(timeout)
	var lastStdout string
	var lastErr error
	for {
		stdout, err := opts.StatusFn()
		lastStdout, lastErr = stdout, err
		if err == nil && IsSessionReadyFromStatus(stdout) {
			return nil
		}
		if !time.Now().Before(deadline) {
			return readyTimeoutErr(timeout, sessionID, lastStdout, lastErr)
		}
		sleep := interval
		if rem := time.Until(deadline); rem < sleep {
			if rem <= 0 {
				return readyTimeoutErr(timeout, sessionID, lastStdout, lastErr)
			}
			sleep = rem
		}
		time.Sleep(sleep)
	}
}

func readyTimeoutErr(timeout time.Duration, sessionID, lastStdout string, lastErr error) error {
	screen, sendable := ParseTTYStatus(lastStdout)
	runnerID := ParseRunnerSessionIDFromStatus(lastStdout)
	return fmt.Errorf(
		"ready timeout after %s: session %q not idle/banner+sendable+bound (screen=%s sendable=%s runner_session_id=%q last status err=%v)",
		timeout, sessionID, screen, sendable, runnerID, lastErr,
	)
}
