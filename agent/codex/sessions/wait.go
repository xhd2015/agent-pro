package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

const (
	defaultWaitTimeout = 30 * time.Minute

	WaitReasonTurnCompleted = "turn_completed"
	WaitReasonOutsideTurn   = "outside_turn"

	boundaryTaskStarted  = "task_started"
	boundaryTaskComplete = "task_complete"
	boundaryTurnAborted  = "turn_aborted"
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

	// WatchCreateMatch injects rollout-file create watch during readiness.
	// nil → logs.WatchCreateMatch.
	WatchCreateMatch func(ctx context.Context, rootDir string, opts logs.WatchCreateMatchOptions, match func(path string) bool, callback func(path string) error) error

	// WaitExclusiveLock blocks until an exclusive flock is acquired on path
	// (or ctx cancels). nil → syscall flock. Used to detect abandoned sessions
	// when the early thread-writer lock is released without a rollout.
	WaitExclusiveLock func(ctx context.Context, lockPath string) error

	// WaitProcessesExit blocks until any pid exits (or ctx cancels).
	// nil → kqueue NOTE_EXIT on Darwin; no-op wait on other platforms.
	// Used so closing the iTerm window (agent-run run exit) wakes ASAP even
	// if flock release lags.
	WaitProcessesExit func(ctx context.Context, pids []int) error

	// ReadinessPIDs injects PIDs to watch during readiness. nil → lsof lock
	// holders + parent chain via ListLiveProcs.
	ReadinessPIDs func(lockPath string) []int
}

// WaitResult is the outcome of a successful Wait.
type WaitResult struct {
	SessionID string
	Reason    string // turn_completed | outside_turn
}

// Wait blocks until the Codex session's current turn is finished, or errors if
// the session is not running. Turn state comes from the rollout JSONL
// (event_msg task_started vs task_complete/turn_aborted), not screen idle.
//
// When Find misses but ~/.codex/thread-writer-locks/<id>.lock exists (common
// right after `kck codex new`), Wait treats the session as pending and races
// rollout create (fsnotify) against lock release (blocking flock). Lock
// release without a rollout → "session never created".
func Wait(codexHome, sessionID string, opts WaitOpts) (*WaitResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	codexHome = strings.TrimSpace(codexHome)
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultWaitTimeout
	}
	statusEvery := opts.StatusInterval
	if statusEvery <= 0 {
		statusEvery = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	st, err := Status(codexHome, sessionID, true, opts.Live)
	if err != nil {
		if !isCodexSessionNotFound(err) {
			return nil, err
		}
		lockPath := threadWriterLockPath(codexHome, sessionID)
		if !fileExists(lockPath) {
			return nil, err
		}
		if _, rerr := waitRolloutReady(ctx, codexHome, sessionID, lockPath, opts); rerr != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout waiting for rollout for session %s", sessionID)
			}
			return nil, rerr
		}
		st, err = Status(codexHome, sessionID, true, opts.Live)
		if err != nil {
			return nil, err
		}
	}
	if st.State != "running" {
		return nil, fmt.Errorf("session not running: %s (state=%s)", sessionID, st.State)
	}

	rolloutPath := strings.TrimSpace(st.Path)
	if rolloutPath == "" {
		return nil, fmt.Errorf("session %s has empty rollout path", sessionID)
	}

	scan := opts.ScanLinesFromTail
	if scan == nil {
		scan = logs.ScanLinesFromTail
	}
	watch := opts.WatchLine
	if watch == nil {
		watch = logs.WatchLine
	}

	cls, err := classifyTurnDetailed(rolloutPath, scan)
	if err != nil {
		return nil, err
	}
	if !cls.inProgress {
		return &WaitResult{SessionID: sessionID, Reason: cls.reason}, nil
	}

	// WatchLine starts at EOF and swallows callback errors, so a close event
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
		errCh <- watch(ctx, rolloutPath, logs.WatchLineOptions{DisableDebounce: true}, func(line string) error {
			kind, ok := rolloutBoundaryKind(line)
			if ok && (kind == boundaryTaskComplete || kind == boundaryTurnAborted) {
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
			if res, ok, err := finishIfOutside(sessionID, rolloutPath, scan); err != nil {
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
			if res, ok, cerr := finishIfOutside(sessionID, rolloutPath, scan); cerr != nil {
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
			st, err := Status(codexHome, sessionID, true, opts.Live)
			if err != nil {
				cancel()
				return nil, err
			}
			if st.State != "running" {
				cancel()
				return nil, fmt.Errorf("session not running: %s (state=%s)", sessionID, st.State)
			}
			if res, ok, err := finishIfOutside(sessionID, rolloutPath, scan); err != nil {
				cancel()
				return nil, err
			} else if ok {
				signalDone()
				return res, nil
			}
		}
	}
}

func finishIfOutside(sessionID, rolloutPath string, scan func(string, logs.ScanLinesFromTailOptions, func(string) (bool, error)) error) (*WaitResult, bool, error) {
	cls, err := classifyTurnDetailed(rolloutPath, scan)
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
		kind, ok := rolloutBoundaryKind(line)
		if !ok {
			return false, nil
		}
		found = kind
		return true, nil
	})
	if err != nil {
		return turnClassifyResult{}, err
	}
	switch found {
	case boundaryTaskStarted:
		return turnClassifyResult{inProgress: true}, nil
	case boundaryTaskComplete, boundaryTurnAborted:
		return turnClassifyResult{inProgress: false, reason: WaitReasonTurnCompleted}, nil
	default:
		return turnClassifyResult{inProgress: false, reason: WaitReasonOutsideTurn}, nil
	}
}

func rolloutBoundaryKind(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}
	var row struct {
		Type    string `json:"type"`
		Payload struct {
			Type string `json:"type"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &row); err != nil {
		return "", false
	}
	if row.Type != "event_msg" {
		return "", false
	}
	kind := strings.TrimSpace(row.Payload.Type)
	switch kind {
	case boundaryTaskStarted, boundaryTaskComplete, boundaryTurnAborted:
		return kind, true
	default:
		return "", false
	}
}
