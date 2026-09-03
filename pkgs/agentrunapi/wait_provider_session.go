package agentrunapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	defaultProviderWaitTimeout = 60 * time.Second
	defaultProviderWaitPoll    = 200 * time.Millisecond
)

// WaitProviderSessionOpts waits for an agent-run TTY session to expose a
// grok/codex provider session id via process open-files (ResolveFromPID).
//
// Typical use after `agent-run run --new-terminal --session-id <id> …`:
// poll registry/tty for PID, then resolve provider id. Prefer CommandPID
// (PTY agent child) over serve PID.
type WaitProviderSessionOpts struct {
	Home      string // agent-run home (required)
	Runner    string // e.g. grok-tty / codex-tty (required)
	SessionID string // agent-run session id (required)

	Timeout      time.Duration // 0 → 60s
	PollInterval time.Duration // 0 → 200ms

	// ReadRegistryPIDs returns serve PID and command PID for the session.
	// err with both 0 → not ready yet (keep polling). nil → production registry read.
	ReadRegistryPIDs func(home, runner, sessionID string) (servePID, commandPID int, err error)

	// ResolveFromPID resolves a provider session id from a process tree rooted at pid.
	// nil → procresolve.ResolveFromPID with live list/lsof.
	ResolveFromPID func(pid int) (*procresolve.Result, error)

	Sleep func(time.Duration) error
}

// WaitProviderSessionResult is a successful provider-session bind.
type WaitProviderSessionResult struct {
	ProviderSessionID string // grok/codex session id
	Kind              string // grok | codex
	RunnerPID         int    // pid that yielded the open-file hit
	ServePID          int
	CommandPID        int
}

// WaitProviderSessionID polls until the agent-run session's TTY registry has a
// PID and procresolve finds a hard grok/codex session id on that process tree.
func WaitProviderSessionID(opts WaitProviderSessionOpts) (WaitProviderSessionResult, error) {
	var zero WaitProviderSessionResult

	home := strings.TrimSpace(opts.Home)
	if home == "" {
		return zero, fmt.Errorf("agent-run home is required")
	}
	runner := strings.TrimSpace(opts.Runner)
	if runner == "" {
		return zero, fmt.Errorf("runner is required")
	}
	sessionID := strings.TrimSpace(opts.SessionID)
	if sessionID == "" {
		return zero, fmt.Errorf("session id is required")
	}

	wantKind := providerKindForRunner(runner)

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultProviderWaitTimeout
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultProviderWaitPoll
	}

	readPIDs := opts.ReadRegistryPIDs
	if readPIDs == nil {
		readPIDs = defaultReadRegistryPIDs
	}
	resolve := opts.ResolveFromPID
	if resolve == nil {
		resolve = defaultResolveFromPID
	}
	sleepFn := opts.Sleep
	if sleepFn == nil {
		sleepFn = func(d time.Duration) error {
			time.Sleep(d)
			return nil
		}
	}

	deadline := time.Now().Add(timeout)
	var lastResolveErr error
	for {
		servePID, commandPID, readErr := readPIDs(home, runner, sessionID)
		if readErr == nil {
			pids := pickResolvePIDs(servePID, commandPID)
			for _, pid := range pids {
				res, err := resolve(pid)
				if err != nil {
					lastResolveErr = err
					continue
				}
				if res == nil || strings.TrimSpace(res.SessionID) == "" || res.Confidence != "hard" {
					continue
				}
				kind := strings.TrimSpace(res.Kind)
				if wantKind != "" && kind != "" && kind != wantKind {
					continue
				}
				if kind == "" {
					kind = wantKind
				}
				return WaitProviderSessionResult{
					ProviderSessionID: strings.TrimSpace(res.SessionID),
					Kind:              kind,
					RunnerPID:         res.RunnerPID,
					ServePID:          servePID,
					CommandPID:        commandPID,
				}, nil
			}
		}

		if !time.Now().Before(deadline) {
			msg := fmt.Sprintf("timeout waiting for %s provider session id for agent-run session %q", wantKindOrRunner(wantKind, runner), sessionID)
			if lastResolveErr != nil {
				msg += fmt.Sprintf(" (last resolve err: %v)", lastResolveErr)
			}
			if readErr != nil {
				msg += fmt.Sprintf(" (last registry err: %v)", readErr)
			}
			return zero, fmt.Errorf("%s", msg)
		}
		sleep := interval
		if rem := time.Until(deadline); rem < sleep {
			if rem <= 0 {
				continue
			}
			sleep = rem
		}
		if err := sleepFn(sleep); err != nil {
			return zero, err
		}
	}
}

func wantKindOrRunner(kind, runner string) string {
	if kind != "" {
		return kind
	}
	return runner
}

func providerKindForRunner(runner string) string {
	r := strings.ToLower(strings.TrimSpace(runner))
	switch {
	case strings.Contains(r, "codex"):
		return "codex"
	case strings.Contains(r, "grok"):
		return "grok"
	default:
		return ""
	}
}

// pickResolvePIDs prefers command_pid (agent child), then serve pid.
func pickResolvePIDs(servePID, commandPID int) []int {
	var out []int
	seen := map[int]bool{}
	for _, pid := range []int{commandPID, servePID} {
		if pid <= 0 || seen[pid] {
			continue
		}
		seen[pid] = true
		out = append(out, pid)
	}
	return out
}

func defaultReadRegistryPIDs(home, runner, sessionID string) (servePID, commandPID int, err error) {
	cfg := registryConfigForRunner(home, runner)
	entry, err := ttywatch.ReadRegistry(cfg, sessionID)
	if err != nil {
		// Fallback: tty.json may exist briefly before registry settles.
		if snap := readTTYJSONPIDs(home, runner, sessionID); snap.servePID > 0 || snap.commandPID > 0 {
			return snap.servePID, snap.commandPID, nil
		}
		return 0, 0, err
	}
	if entry == nil {
		return 0, 0, fmt.Errorf("registry entry nil")
	}
	return entry.PID, entry.CommandPID, nil
}

type ttyJSONPIDs struct {
	servePID   int
	commandPID int
}

func readTTYJSONPIDs(home, runner, sessionID string) ttyJSONPIDs {
	// Best-effort: sessions/<id>/tty.json stores the serve pid (not command_pid).
	_ = runner
	path := filepath.Join(home, "sessions", sessionID, "tty.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ttyJSONPIDs{}
	}
	var snap struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return ttyJSONPIDs{}
	}
	return ttyJSONPIDs{servePID: snap.PID}
}

func defaultResolveFromPID(pid int) (*procresolve.Result, error) {
	return procresolve.ResolveFromPID(pid, procresolve.Options{
		ListProcs: procresolve.ListLiveProcs,
		Lsof:      procresolve.LiveLsof,
	})
}
