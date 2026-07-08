# Grok Reconcile Tests (pkgs/agentsync)

Doc-style tests for **per-session flock** (`grok-sync.lock`) and **background
reconcile** (`ReconcileOnce`, `StartReconciler`) that heal finished sessions with
empty `events.jsonl` when grok updates exist on disk.

# DSN (Domain Specific Notion)

**Participants**

- **grok-sync.lock** — per-session flock file; `LOCK_EX|LOCK_NB` acquire; held while
  worker or reconcile pass is active.
- **SessionLock** — `TryAcquireSessionLock(sessionDir)` returns release func or skips
  when lock held by another acquirer.
- **ReconcileOnce** — scans one candidate session: skip if in-process worker active;
  try flock; call `EnsureGrokSync` with discovery using `meta.initial_prompt`.
- **meta.json** — session metadata with `initial_prompt`, `runner_session_id`, `status`.
- **GrokSyncWorker** — in-process registry; reconcile must not double-sync when active.
- **Test harness** — temp `AGENT_RUN_HOME` layout, synthetic grok updates, file store sink.

**Behaviors**

- **R1** — first acquirer holds `grok-sync.lock`; second `TryAcquireSessionLock`
  returns `acquired=false` without blocking.
- **R2** — `meta.status=finished`, no `events.jsonl`, grok updates pre-seeded with
  matching `initial_prompt`; `ReconcileOnce` populates `events.jsonl`.
- **R3** — in-process worker already tailing; `ReconcileOnce` skips (no duplicate lines).

## Version

0.0.2

## Decision Tree

```
pkgs/agentsync/tests/grok-reconcile/
├── DOCTEST.md
├── SETUP.md
├── lock/
│   └── nonblocking-second-acquirer-skips/   # R1: flock NB second acquirer
└── reconcile/
    ├── heals-finished-empty-events/         # R2: reconcile heals empty chat
    └── skips-when-worker-active/            # R3: no double-sync
```

Parameter ranking (most → least significant):

1. **Operation** — flock acquire vs reconcile pass
2. **Session state** — empty events + finished vs worker already active
3. **Grok data** — pre-seeded updates vs none

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| R1 | `lock/nonblocking-second-acquirer-skips` | Holder keeps lock; second TryLock fails (non-blocking) |
| R2 | `reconcile/heals-finished-empty-events` | Finished session, no events; ReconcileOnce populates events.jsonl |
| R3 | `reconcile/skips-when-worker-active` | Active worker; reconcile does not duplicate-sync same lines |

## How to Run

```sh
doctest vet ./pkgs/agentsync/tests/grok-reconcile
doctest test ./pkgs/agentsync/tests/grok-reconcile    # RED before implement

doctest test -v ./pkgs/agentsync/tests/grok-reconcile/lock/nonblocking-second-acquirer-skips
doctest test -v ./pkgs/agentsync/tests/grok-reconcile/reconcile/heals-finished-empty-events
```

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentsync"
)

type Request struct {
	TempDir    string
	AgentHome  string
	SessionDir string
	GrokHome   string
	Runner     string
	SessionID  string
	Workspace  string

	Mode string // flock-nb | reconcile-heal | reconcile-skip-worker

	InitialPrompt   string
	GrokSessionID   string
	GrokUpdatesPath string
	SessionStatus   string

	ReconcileTimeout time.Duration
	WorkerHold       time.Duration
}

type Response struct {
	FirstLockAcquired  bool
	SecondLockAcquired bool
	Events             []types.AgentEvent
	ReconcileErr       error
	ReconcileSkipped     bool
	EventLineCountBefore int
	EventLineCountAfter  int
	WorkerActive         bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.AgentHome == "" {
		req.AgentHome = filepath.Join(req.TempDir, ".agent-run")
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.SessionID == "" {
		req.SessionID = "reconcile-test"
	}
	if req.Workspace == "" {
		req.Workspace = req.TempDir
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return nil, err
	}
	req.SessionDir = filepath.Join(req.AgentHome, "sessions", req.Runner, req.SessionID)

	switch req.Mode {
	case "flock-nb":
		return runFlockNonBlocking(t, req)
	case "reconcile-heal":
		return runReconcileHeal(t, req)
	case "reconcile-skip-worker":
		return runReconcileSkipWorker(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runFlockNonBlocking(t *testing.T, req *Request) (*Response, error) {
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return nil, err
	}
	release1, acquired1, err := agentsync.TryAcquireSessionLock(req.SessionDir)
	if err != nil {
		return nil, err
	}
	if !acquired1 {
		return &Response{FirstLockAcquired: false}, nil
	}
	defer release1()

	_, acquired2, err := agentsync.TryAcquireSessionLock(req.SessionDir)
	if err != nil {
		return nil, err
	}
	return &Response{
		FirstLockAcquired:  true,
		SecondLockAcquired: acquired2,
	}, nil
}

func runReconcileHeal(t *testing.T, req *Request) (*Response, error) {
	status := req.SessionStatus
	if status == "" {
		status = "finished"
	}
	if err := seedFinishedEmptySession(t, req, status); err != nil {
		return nil, err
	}
	before := len(readEventsFromSessionDir(t, req.SessionDir))

	ctx, cancel := context.WithTimeout(context.Background(), req.ReconcileTimeout)
	defer cancel()
	reconcileErr := agentsync.ReconcileOnce(ctx, agentsync.ReconcileOptions{
		Home:     req.AgentHome,
		GrokHome: req.GrokHome,
		Runner:   req.Runner,
		SessionID: req.SessionID,
	})

	deadline := time.Now().Add(8 * time.Second)
	var events []types.AgentEvent
	for time.Now().Before(deadline) {
		events = readEventsFromSessionDir(t, req.SessionDir)
		if len(events) > before {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return &Response{
		Events:               events,
		ReconcileErr:         reconcileErr,
		EventLineCountBefore: before,
		EventLineCountAfter:  len(events),
	}, nil
}

func runReconcileSkipWorker(t *testing.T, req *Request) (*Response, error) {
	if err := seedRunningSessionWithUpdates(t, req); err != nil {
		return nil, err
	}
	sink := agentsync.NewFileGrokSyncSink(req.SessionDir, req.GrokSessionID, req.GrokUpdatesPath)
	ctx := context.Background()
	opts := agentsync.GrokSyncOptions{
		Runner:        req.Runner,
		SessionID:     req.SessionID,
		GrokSessionID: req.GrokSessionID,
		UpdatesPath:   req.GrokUpdatesPath,
		Workspace:     req.Workspace,
		GrokHome:      req.GrokHome,
		Sink:          sink,
	}
	if err := agentsync.EnsureGrokSync(ctx, opts); err != nil {
		return nil, err
	}
	hold := req.WorkerHold
	if hold <= 0 {
		hold = 400 * time.Millisecond
	}
	time.Sleep(hold)

	before := len(readEventsFromSessionDir(t, req.SessionDir))
	reconcileCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	reconcileErr := agentsync.ReconcileOnce(reconcileCtx, agentsync.ReconcileOptions{
		Home:     req.AgentHome,
		GrokHome: req.GrokHome,
		Runner:   req.Runner,
		SessionID: req.SessionID,
	})
	time.Sleep(300 * time.Millisecond)
	after := len(readEventsFromSessionDir(t, req.SessionDir))

	return &Response{
		Events:               readEventsFromSessionDir(t, req.SessionDir),
		ReconcileErr:         reconcileErr,
		ReconcileSkipped:     true,
		EventLineCountBefore: before,
		EventLineCountAfter:  after,
		WorkerActive:         agentsync.GrokSyncWorkerActive(req.Runner, req.SessionID),
	}, nil
}
```