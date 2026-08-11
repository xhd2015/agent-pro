# agentrunapi — OnTTYStarted lifecycle hook (true first live TTY)

Classic TDD doctests for **P1 library surface**: neutral
`OnTTYStarted` fires only when a session **truly first gets a live TTY**.
Live follow-up / send into an existing TTY must not fire again. Nil hook is a
no-op (no panic).

Nested root so the parent `tests/agentrunapi` Classify/AutoSendOrResume suite
stays **GREEN** while this tree is **RED** until implementer lands the hook.

**Out of scope:** HTTP / event-bus publish (see `cmd/agent-run/tests/event-bus`
library-wire + double-fire); ai-critic; ForceNew iTerm.

## Version

0.0.2

# DSN (Domain Specific Notion)

Library notifies callers once when a session first obtains a live TTY; live
send does not re-notify. Hook is optional and best-effort from the library POV.

**Participants**

- **Caller** — sets optional `Opts.OnTTYStarted` before `AutoSendOrResume`.
- **`TTYStartedInfo`** — `SessionID`, `Runner`, `Workspace` (empty strings OK
  when unknown; SessionID must match the run session when known).
- **`Opts.OnTTYStarted`** — `func(TTYStartedInfo)`; nil → no-op.
- **`AutoSendOrResume`** — after a dispatch that **newly establishes** a live
  TTY for the session (typically successful **ModeRun** / first open), calls
  `OnTTYStarted` **once**. Does **not** call it on **ModeSend** (live
  follow-up into an existing TTY).
- **Dispatch hooks** (`RunSession` / `SendLive`) — L2 stubs so leaves need no
  real agent binary / iTerm; library still applies OnTTYStarted policy after
  successful dispatch.

**Behaviors**

```
# newly established TTY (ModeRun)
OnTTYStarted set + ModeRun success
  -> hook once; info.SessionID matches opts.SessionID
  -> Runner/Workspace from opts/meta when known
OnTTYStarted nil + ModeRun success
  -> no panic; run completes (hooks run)

# already live (ModeSend)
after first establish (hook count=1), ModeSend follow-up
  -> OnTTYStarted not called again (count stays 1)
```

### Planned public API (Classic TDD — locked for implementer)

```go
// Package: github.com/xhd2015/agent-pro/pkgs/agentrunapi

// TTYStartedInfo is passed to OnTTYStarted when a session first gets a live TTY.
type TTYStartedInfo struct {
	SessionID string
	Runner    string
	Workspace string
}

// On agentrunapi.Opts:
// OnTTYStarted, when non-nil, is invoked once when AutoSendOrResume (or the
// production run path it owns) newly establishes a live TTY for the session.
// Nil = no-op. Not invoked on live send / follow-up into an existing TTY.
// Best-effort: a panic in the hook must not fail the run path (recover +
// optional warn) if product already uses warn-only for bus side effects.
OnTTYStarted func(info TTYStartedInfo)
```

## Decision Tree

```
tests/agentrunapi/on-tty-started/
├── DOCTEST.md
├── SETUP.md
├── newly-established/                 # first live TTY (ModeRun)
│   ├── SETUP.md
│   ├── fires-once/                    # L1: hook set → once + SessionID
│   └── nil-hook-ok/                   # L2: nil hook → no panic; ModeRun ok
└── follow-up-live/                    # second call while live
    ├── SETUP.md
    └── no-second-fire/                # L3: after first fire, ModeSend → no +1
```

Parameter ranking (most → least significant):

1. **Lifecycle** — newly established TTY vs already-live follow-up
2. **Hook presence** — set vs nil (only under newly-established)

## Test Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| 1 | `newly-established/fires-once` | ModeRun + OnTTYStarted set → hook **once**; SessionID (+runner/workspace) | RED |
| 2 | `newly-established/nil-hook-ok` | ModeRun + OnTTYStarted nil → no panic; RunSession once | RED |
| 3 | `follow-up-live/no-second-fire` | ModeRun then ModeSend → total hook calls **1** | RED |

## How to Run

```sh
# From agent-pro module root:
doctest vet ./tests/agentrunapi/on-tty-started
doctest test ./tests/agentrunapi/on-tty-started

doctest test -v ./tests/agentrunapi/on-tty-started/newly-established
doctest test -v ./tests/agentrunapi/on-tty-started/follow-up-live
```

Expect **RED** (compile or assert) until `TTYStartedInfo` + `Opts.OnTTYStarted`
and AutoSendOrResume policy land.

```go
import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// Op selects the scenario shape under test.
//   newly-established — one AutoSendOrResume ModeRun
//   follow-up-live    — ModeRun then ModeSend; same OnTTYStarted recorder
const (
	opNewlyEstablished = "newly-established"
	opFollowUpLive     = "follow-up-live"
)

// Request drives one OnTTYStarted policy scenario.
type Request struct {
	Op string

	// When true, install a recording OnTTYStarted. When false, leave nil.
	InstallHook bool

	SessionID   string
	Prompt      string
	Home        string
	Runner      string
	Workspace   string
	// SeedMeta / probe fields for ModeSend (follow-up leaf).
	SeedMeta     bool
	TerminalID   string
	RunnerSessID string
}

// HookCall is one recorded OnTTYStarted invocation.
type HookCall struct {
	SessionID string
	Runner    string
	Workspace string
}

// Response holds harness observations.
type Response struct {
	ErrString  string
	HookCalls  []HookCall
	HookCount  int
	RunCalls   int
	SendCalls  int
	// Second AutoSendOrResume error (follow-up leaf only).
	ErrString2 string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	resp := &Response{}

	store, err := openStore(t, req)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	record := func(info agentrunapi.TTYStartedInfo) {
		mu.Lock()
		defer mu.Unlock()
		resp.HookCalls = append(resp.HookCalls, HookCall{
			SessionID: info.SessionID,
			Runner:    info.Runner,
			Workspace: info.Workspace,
		})
		resp.HookCount = len(resp.HookCalls)
	}

	switch req.Op {
	case opNewlyEstablished:
		opts := baseOpts(req, store)
		opts.Probe = agentrunapi.EmptyProbe
		if req.InstallHook {
			opts.OnTTYStarted = record
		}
		opts.RunSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			_ = ctx
			_ = o
			_ = meta
			_ = found
			resp.RunCalls++
			return nil
		}
		opts.SendLive = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			_ = ctx
			_ = o
			_ = meta
			resp.SendCalls++
			return nil
		}
		if err := agentrunapi.AutoSendOrResume(context.Background(), opts); err != nil {
			resp.ErrString = err.Error()
		}
		mu.Lock()
		resp.HookCount = len(resp.HookCalls)
		mu.Unlock()
		return resp, nil

	case opFollowUpLive:
		// First: ModeRun establish (missing session → run).
		opts1 := baseOpts(req, store)
		opts1.Probe = agentrunapi.EmptyProbe
		opts1.OnTTYStarted = record
		opts1.RunSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			_ = ctx
			_ = o
			_ = meta
			_ = found
			resp.RunCalls++
			return nil
		}
		opts1.SendLive = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			_ = ctx
			_ = o
			_ = meta
			resp.SendCalls++
			return nil
		}
		if err := agentrunapi.AutoSendOrResume(context.Background(), opts1); err != nil {
			resp.ErrString = err.Error()
		}

		// Seed live meta so second call classifies ModeSend.
		if err := seedLiveSession(t, store, req); err != nil {
			return nil, err
		}
		live := false
		opts2 := baseOpts(req, store)
		opts2.Prompt = "follow-up after live TTY"
		opts2.Probe = func(s agentstorage.Store, meta agentstorage.SessionMeta) (agentrunapi.ProbeReport, error) {
			_ = s
			_ = meta
			return agentrunapi.ProbeReport{RunnerExited: &live}, nil
		}
		opts2.OnTTYStarted = record
		opts2.RunSession = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta, found bool) error {
			_ = ctx
			_ = o
			_ = meta
			_ = found
			resp.RunCalls++
			return nil
		}
		opts2.SendLive = func(ctx context.Context, o agentrunapi.Opts, meta agentstorage.SessionMeta) error {
			_ = ctx
			_ = o
			_ = meta
			resp.SendCalls++
			return nil
		}
		if err := agentrunapi.AutoSendOrResume(context.Background(), opts2); err != nil {
			resp.ErrString2 = err.Error()
		}
		mu.Lock()
		resp.HookCount = len(resp.HookCalls)
		mu.Unlock()
		return resp, nil

	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}

func baseOpts(req *Request, store agentstorage.Store) agentrunapi.Opts {
	return agentrunapi.Opts{
		SessionID:    req.SessionID,
		Prompt:       req.Prompt,
		WorkspaceDir: req.Workspace,
		AgentRunner:  req.Runner,
		Store:        store,
		NewTerminal:  false,
	}
}

func openStore(t *testing.T, req *Request) (agentstorage.Store, error) {
	t.Helper()
	home := req.Home
	if home == "" {
		home = filepath.Join(t.TempDir(), ".agent-run")
	}
	return agentstorage.NewFileStore(home)
}

func seedLiveSession(t *testing.T, store agentstorage.Store, req *Request) error {
	t.Helper()
	runner := strings.TrimSpace(req.Runner)
	if runner == "" {
		runner = "grok-tty"
	}
	termID := strings.TrimSpace(req.TerminalID)
	if termID == "" {
		termID = "term-" + req.SessionID
	}
	runnerSess := strings.TrimSpace(req.RunnerSessID)
	if runnerSess == "" {
		runnerSess = "runner-" + req.SessionID
	}
	meta := agentstorage.SessionMeta{
		SessionID:         req.SessionID,
		Runner:            runner,
		RunnerSessionID:   runnerSess,
		TerminalSessionID: termID,
		Workspace:         req.Workspace,
		Status:            "running",
	}
	return store.CreateSession(req.SessionID, meta)
}
```
