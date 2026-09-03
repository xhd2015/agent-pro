package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

const (
	defaultWaitTimeout = 30 * time.Minute

	WaitReasonTurnCompleted = "turn_completed"
	WaitReasonOutsideTurn   = "outside_turn"

	turnKindUser      = "user_message_chunk"
	turnKindCompleted = "turn_completed"
)

// WaitOpts drives Wait.
type WaitOpts struct {
	Timeout time.Duration // 0 → 30m
	Live    *LiveOptions  // nil → production Status probes

	// StatusInterval rechecks liveness and turn state while watching. 0 → 2s.
	StatusInterval time.Duration

	// ScanLinesFromTail injects Phase A. nil → logs.ScanLinesFromTail.
	ScanLinesFromTail func(path string, opts logs.ScanLinesFromTailOptions, fn func(line string) (stop bool, err error)) error

	// WatchLine injects Phase B. nil → logs.WatchLine.
	// Note: logs.WatchLine swallows callback errors; Wait signals completion
	// via cancel + re-classify, not via callback error returns.
	WatchLine func(ctx context.Context, path string, opts logs.WatchLineOptions, fn func(line string) error) error
}

// WaitResult is the outcome of a successful Wait.
type WaitResult struct {
	SessionID string
	Reason    string // turn_completed | outside_turn
}

// Wait blocks until the Grok session's current turn is finished, or errors if
// the session is not running. Turn state comes from updates.jsonl
// (user_message_chunk vs turn_completed), not screen idle.
func Wait(grokHome, sessionID string, opts WaitOpts) (*WaitResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}
	statusEvery := opts.StatusInterval
	if statusEvery <= 0 {
		statusEvery = 2 * time.Second
	}

	st, err := Status(grokHome, sessionID, true, opts.Live)
	if err != nil {
		return nil, err
	}
	if st.State != "running" {
		return nil, fmt.Errorf("session not running: %s (state=%s)", sessionID, st.State)
	}

	updatesPath := filepath.Join(filepath.Dir(st.Path), "updates.jsonl")
	scan := opts.ScanLinesFromTail
	if scan == nil {
		scan = logs.ScanLinesFromTail
	}
	watch := opts.WatchLine
	if watch == nil {
		watch = logs.WatchLine
	}

	cls, err := classifyTurnDetailed(updatesPath, scan)
	if err != nil {
		return nil, err
	}
	if !cls.inProgress {
		return &WaitResult{SessionID: sessionID, Reason: cls.reason}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// WatchLine starts at EOF and swallows callback errors, so a turn_completed
	// that lands between Phase A and watch-attach would be missed without
	// re-classify. Signal + periodic ScanLinesFromTail cover that race.
	done := make(chan struct{}, 1)
	signalDone := func() {
		select {
		case done <- struct{}{}:
		default:
		}
		cancel()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watch(ctx, updatesPath, logs.WatchLineOptions{DisableDebounce: true}, func(line string) error {
			kind, ok := sessionUpdateKind(line)
			if ok && kind == turnKindCompleted {
				signalDone()
			}
			return nil
		})
	}()

	ticker := time.NewTicker(statusEvery)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return &WaitResult{SessionID: sessionID, Reason: WaitReasonTurnCompleted}, nil
		case <-ctx.Done():
			if res, ok, err := finishIfOutside(sessionID, updatesPath, scan); err != nil {
				return nil, err
			} else if ok {
				return res, nil
			}
			return nil, fmt.Errorf("timeout waiting for session %s after %s (state=in_progress)", sessionID, timeout)
		case err := <-errCh:
			select {
			case <-done:
				return &WaitResult{SessionID: sessionID, Reason: WaitReasonTurnCompleted}, nil
			default:
			}
			if res, ok, cerr := finishIfOutside(sessionID, updatesPath, scan); cerr != nil {
				return nil, cerr
			} else if ok {
				return res, nil
			}
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout waiting for session %s after %s (state=in_progress)", sessionID, timeout)
			}
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timeout waiting for session %s after %s (state=in_progress)", sessionID, timeout)
		case <-ticker.C:
			st, err := Status(grokHome, sessionID, true, opts.Live)
			if err != nil {
				cancel()
				return nil, err
			}
			if st.State != "running" {
				cancel()
				return nil, fmt.Errorf("session not running: %s (state=%s)", sessionID, st.State)
			}
			if res, ok, err := finishIfOutside(sessionID, updatesPath, scan); err != nil {
				cancel()
				return nil, err
			} else if ok {
				signalDone()
				return res, nil
			}
		}
	}
}

func finishIfOutside(sessionID, updatesPath string, scan func(string, logs.ScanLinesFromTailOptions, func(string) (bool, error)) error) (*WaitResult, bool, error) {
	cls, err := classifyTurnDetailed(updatesPath, scan)
	if err != nil {
		return nil, false, err
	}
	if cls.inProgress {
		return nil, false, nil
	}
	return &WaitResult{SessionID: sessionID, Reason: cls.reason}, true, nil
}

type turnClassifyResult struct {
	inProgress bool
	reason     string // set when !inProgress
}

func classifyTurnDetailed(path string, scan func(string, logs.ScanLinesFromTailOptions, func(string) (bool, error)) error) (turnClassifyResult, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return turnClassifyResult{inProgress: false, reason: WaitReasonOutsideTurn}, nil
		}
		return turnClassifyResult{}, err
	}

	var found string
	err := scan(path, logs.ScanLinesFromTailOptions{}, func(line string) (bool, error) {
		kind, ok := sessionUpdateKind(line)
		if !ok {
			return false, nil
		}
		if kind == turnKindUser || kind == turnKindCompleted {
			found = kind
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return turnClassifyResult{}, err
	}
	switch found {
	case turnKindUser:
		return turnClassifyResult{inProgress: true}, nil
	case turnKindCompleted:
		return turnClassifyResult{inProgress: false, reason: WaitReasonTurnCompleted}, nil
	default:
		return turnClassifyResult{inProgress: false, reason: WaitReasonOutsideTurn}, nil
	}
}

func sessionUpdateKind(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return "", false
	}
	update := extractSessionUpdate(raw)
	if update == nil {
		return "", false
	}
	kind, _ := update["sessionUpdate"].(string)
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "", false
	}
	return kind, true
}
