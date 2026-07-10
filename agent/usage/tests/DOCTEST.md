# agent/usage facade tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/usage`, the unified
in-process usage fetch facade over Grok and Codex TTY providers — including
Codex **registry flock lifetime** during long-running `FetchStatus`.

# DSN (Domain Specific Notion)

**Caller** (ai-critic menubar services or other libraries) invokes
`usage.Fetch(ctx, providerID)` and receives a normalized **Snapshot**.

**usage.Fetch** dispatches by `ProviderID`:

- `Grok` → `agent/grok/tty.FetchUsage` (direct PTY; honors `GROK_SHOW_USAGE_COMMAND`)
- `Codex` → `agent/codex/tty.FetchStatus` (in-process ttywatch session; honors
  `CODEX_SHOW_STATUS_COMMAND` and `TTY_WATCH_HOME`)

**Codex FetchStatus lock participants**

- **Registry** — `{TTY_WATCH_HOME}/registry/`; exclusive flock on `.lock`
- **ReserveCustomSessionID** — under flock: prune stale, write `.session.claim`, return `release`
- **FetchStatus** — must call **`release()` before** `StartInProcess` / prompt-status waits so
  other processes can acquire the registry flock while a usage fetch is still “in progress”
- **Ephemeral session** — claim / `session.json` keep the custom session id reserved; same-id
  re-reserve still fails with “already in use” after the flock is free
- **Fake TUI hooks** — replace production agent binaries for deterministic doctests

```
caller -> usage.Fetch(ctx, Grok|Codex)
  -> provider tty helper -> fake TUI (env command hook) -> parse -> Snapshot
doctest <- Snapshot fields (provider, usage %, reset, credits)

# Codex lock lifetime (Mode=lock-during-fetch)
FetchStatus -> ReserveCustomSessionID -> release flock  # must not defer across StartInProcess
            -> StartInProcess / wait (long)
concurrent probe -> LOCK_NB on registry/.lock  OR  ReserveCustomSessionID(other-id)
doctest <- LockAcquiredDuring / SecondReserveOK while fetch still running
```

Tests call `usage.Fetch` / `ttywatch` APIs directly (no CLI subprocess).

## Version

0.0.2

## Decision Tree

```
agent/usage/tests/
├── DOCTEST.md
├── SETUP.md                                    # env helpers, fake TUI scripts, isolated TTY_WATCH_HOME
├── fetch-grok-mock/                            # Mode=fetch: GROK fake → weekly limit + reset
├── fetch-codex-mock/                           # Mode=fetch: CODEX fake → monthly % + credits + reset
└── fetch-codex-lock-lifetime/                  # Mode=lock-during-fetch: registry flock while fetch runs
    ├── SETUP.md                                # Codex + blocking fake TUI + Mode
    ├── free-while-in-progress/                 # concurrent flock / other-id reserve succeed mid-fetch
    └── same-id-still-in-use/                   # same session id still rejected as already in use
```

Parameter ranking (most → least significant):

1. **Observation mode** — snapshot fetch vs lock-lifetime probe during long fetch
2. **Provider** — Grok vs Codex (different tty backends and env hooks)
3. **Lock probe kind** — free lock / other-id reserve vs same-id “already in use”
4. **Runner backend** — fake TUI via env command hook (deterministic fixtures)
5. **Snapshot fields** — provider-specific parsed values in normalized struct

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fetch-grok-mock` | `Fetch(ctx, Grok)` with fake hook returns weekly 1% + July 9 reset |
| 2 | `fetch-codex-mock` | `Fetch(ctx, Codex)` with fake hook returns 58% usage, credits 6519/11250, reset |
| 3 | `fetch-codex-lock-lifetime/free-while-in-progress` | Mid long-running Codex fetch, registry `.lock` is free (LOCK_NB + other-id reserve succeed); cleanup after cancel (RED) |
| 4 | `fetch-codex-lock-lifetime/same-id-still-in-use` | Mid long-running Codex fetch, re-reserve of the same custom id fails with “already in use” (not lock-busy) (RED) |

## How to Run

```sh
doctest vet ./agent/usage/tests
doctest test ./agent/usage/tests/...
doctest test -v ./agent/usage/tests/fetch-grok-mock
doctest test -v ./agent/usage/tests/fetch-codex-mock
doctest test -v ./agent/usage/tests/fetch-codex-lock-lifetime/free-while-in-progress
doctest test -v ./agent/usage/tests/fetch-codex-lock-lifetime/same-id-still-in-use
```

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type Request struct {
	Provider          usage.ProviderID
	TTYWatchHome      string
	ShowUsageCommand  string // GROK_SHOW_USAGE_COMMAND
	ShowStatusCommand string // CODEX_SHOW_STATUS_COMMAND
	SessionID         string // CODEX_SHOW_STATUS_SESSION_ID; default codex-status-usage
	TimeoutSeconds    string // provider timeout env override; empty = default
	// Mode selects Run path: ""|"fetch" = usage.Fetch snapshot; "lock-during-fetch" = concurrent flock probe.
	Mode string
	// ProbeSessionID is the other custom id used for concurrent ReserveCustomSessionID (lock-during-fetch).
	ProbeSessionID string
	// SameIDProbe when true records a concurrent ReserveCustomSessionID of the fetch session id.
	SameIDProbe bool
}

type Response struct {
	Snapshot *usage.Snapshot

	// lock-during-fetch observations
	ClaimOrRegistrySeen    bool
	LockAcquiredDuring     bool
	LockAcquireErr         string
	SecondReserveOK        bool
	SecondReserveErr       string
	SameIDReserveErr       string
	SameIDReserveSucceeded bool
	FetchErr               string
	FetchCompleted         bool
	RegistryGoneAfter      bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "fetch"
	}

	switch mode {
	case "fetch":
		return runFetchSnapshot(t, req)
	case "lock-during-fetch":
		return runLockDuringFetch(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", mode)
	}
}

func runFetchSnapshot(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Provider == "" {
		t.Fatal("req.Provider must be set by leaf Setup")
	}

	timeout := 75 * time.Second
	if req.TimeoutSeconds != "" {
		if sec, err := time.ParseDuration(req.TimeoutSeconds + "s"); err == nil {
			timeout = sec + 20*time.Second
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	applyUsageEnv(t, req)

	snap, err := usage.Fetch(ctx, req.Provider)
	if err != nil {
		return nil, err
	}
	return &Response{Snapshot: snap}, nil
}

func runLockDuringFetch(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Provider != usage.Codex {
		t.Fatal("lock-during-fetch requires Provider=Codex")
	}
	if req.TTYWatchHome == "" {
		t.Fatal("lock-during-fetch requires isolated TTYWatchHome")
	}
	if strings.TrimSpace(req.ShowStatusCommand) == "" {
		t.Fatal("lock-during-fetch requires a blocking ShowStatusCommand")
	}

	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = "codex-status-usage"
	}
	probeID := strings.TrimSpace(req.ProbeSessionID)
	if probeID == "" {
		probeID = "probe-other-session"
	}

	// Short overall budget: we cancel after the mid-fetch probe.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	applyUsageEnv(t, req)

	type fetchResult struct {
		err error
	}
	done := make(chan fetchResult, 1)
	go func() {
		_, err := usage.Fetch(ctx, usage.Codex)
		done <- fetchResult{err: err}
	}()

	claimPath := filepath.Join(req.TTYWatchHome, "registry", "."+sid+".claim")
	registryJSON := filepath.Join(req.TTYWatchHome, "registry", sid+".json")
	lockPath := filepath.Join(req.TTYWatchHome, "registry", ".lock")

	// Wait until reserve has completed (claim written and/or session registered).
	waitDeadline := time.Now().Add(15 * time.Second)
	seen := false
	for time.Now().Before(waitDeadline) {
		if fileExists(claimPath) || fileExists(registryJSON) {
			seen = true
			break
		}
		select {
		case fr := <-done:
			return &Response{
				ClaimOrRegistrySeen: false,
				FetchErr:            errString(fr.err),
				FetchCompleted:      true,
			}, nil
		case <-time.After(25 * time.Millisecond):
		}
	}

	// Brief settle so an early-release implementation unlocks before StartInProcess work.
	// With deferred release spanning the whole fetch, the exclusive flock remains held.
	time.Sleep(100 * time.Millisecond)

	resp := &Response{ClaimOrRegistrySeen: seen}

	// Probe 1: non-blocking exclusive flock on registry/.lock (same home, mid-fetch).
	// tryRegistryLockNB unlocks immediately on success so we do not hold the lock.
	acquired, lockErr := tryRegistryLockNB(lockPath)
	resp.LockAcquiredDuring = acquired
	if lockErr != nil {
		resp.LockAcquireErr = lockErr.Error()
	}

	// Probe 2: ReserveCustomSessionID for a *different* session id on the same home.
	// Must succeed quickly when flock is free; fails with lock-busy when held for whole fetch.
	releaseOther, errOther := ttywatch.ReserveCustomSessionID(
		ttywatch.DefaultRegistryConfig(req.TTYWatchHome), probeID,
	)
	if errOther != nil {
		resp.SecondReserveOK = false
		resp.SecondReserveErr = errOther.Error()
	} else {
		resp.SecondReserveOK = true
		releaseOther()
		// Drop provisional claim for the probe id (tempdir cleanup is not enough for same-home reuse).
		_ = os.Remove(filepath.Join(req.TTYWatchHome, "registry", "."+probeID+".claim"))
	}

	// Optional probe 3: same custom id must still report "already in use" once flock is free.
	if req.SameIDProbe {
		releaseSame, errSame := ttywatch.ReserveCustomSessionID(
			ttywatch.DefaultRegistryConfig(req.TTYWatchHome), sid,
		)
		if errSame != nil {
			resp.SameIDReserveErr = errSame.Error()
			resp.SameIDReserveSucceeded = false
		} else {
			resp.SameIDReserveSucceeded = true
			releaseSame()
		}
	}

	// End the long fetch; require Kill/cleanup on cancel path.
	cancel()
	var fr fetchResult
	select {
	case fr = <-done:
		resp.FetchCompleted = true
		resp.FetchErr = errString(fr.err)
	case <-time.After(20 * time.Second):
		resp.FetchCompleted = false
		resp.FetchErr = "fetch did not return after cancel"
	}

	// Allow session.Kill + registry prune a moment.
	cleanupDeadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(cleanupDeadline) {
		if !fileExists(registryJSON) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	resp.RegistryGoneAfter = !fileExists(registryJSON)

	return resp, nil
}

func applyUsageEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.TTYWatchHome != "" {
		t.Setenv("TTY_WATCH_HOME", req.TTYWatchHome)
	}
	if req.ShowUsageCommand != "" {
		t.Setenv("GROK_SHOW_USAGE_COMMAND", req.ShowUsageCommand)
	}
	if req.ShowStatusCommand != "" {
		t.Setenv("CODEX_SHOW_STATUS_COMMAND", req.ShowStatusCommand)
	}
	if req.TimeoutSeconds != "" {
		switch req.Provider {
		case usage.Grok:
			t.Setenv("GROK_SHOW_USAGE_TIMEOUT", req.TimeoutSeconds)
		case usage.Codex:
			t.Setenv("CODEX_SHOW_STATUS_TIMEOUT", req.TimeoutSeconds)
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = "codex-status-usage"
	}
	t.Setenv("CODEX_SHOW_STATUS_SESSION_ID", sid)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// tryRegistryLockNB attempts LOCK_EX|LOCK_NB on path. On success it unlocks and closes before return.
func tryRegistryLockNB(lockPath string) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return false, err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return false, err
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return true, nil
}
```
