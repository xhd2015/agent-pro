# Grok Sync Worker Tests (pkgs/agentsync)

Doc-style tests for the persistent **grok sync worker** in `pkgs/agentsync` — a single
per-session goroutine that tails `updates.jsonl`, converts ACP lines via
`grok_session.Converter`, appends canonical events, checkpoints byte offset +
`turn_index` in `grok-sync.json`, and **discovers** grok sessions when
`runner_session_id` / `updates_path` are not yet known (PRIMARY empty-chat fix).

# DSN (Domain Specific Notion)

**Participants**

- **updates.jsonl** — grok on-disk ACP stream under `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/`.
- **GrokSyncWorker** — long-lived tail loop: `TailUpdatesFromOffset` from checkpoint
  offset; on each complete line processed + events emitted → persist checkpoint.
- **DiscoverBootstrap** — when `GrokSessionID` / `UpdatesPath` missing, polls
  `agenttty.DiscoverSession` using `initial_prompt` + `sessionCreatedAt` until match;
  persists `runner_session_id` via sink.
- **grok_session.Converter** — coalesces wire chunks; `turn_index` restored from
  checkpoint on worker resume.
- **GrokSyncSink** — storage-agnostic adapter: `AppendEvent`, `LoadCheckpoint`,
  `SaveCheckpoint`, `OnTurnCompleted`, `UpdateRunnerSessionID`.
- **grok-sync.json** — checkpoint at `sessions/grok-tty/<id>/grok-sync.json` with
  `updates_offset` and `turn_index`; written **after** each `events.jsonl` append.
- **EnsureGrokSync** — idempotent entry point; mutex registry guarantees one active
  worker per `(runner, sessionID)`; holds `grok-sync.lock` while active.
- **Test harness** — synthetic ACP lines, temp session dir as fake agent-run home,
  `NewFileGrokSyncSink`, delayed grok session seed for discovery leaves.

**Behaviors**

- **A1** — two turns appended while one worker runs → exactly one user `message` per
  distinct prompt text (duplicate-bug guard, preserved after package move).
- **A2** — concurrent `EnsureGrokSync` calls → single active worker; no duplicate
  events for the same updates line.
- **A3** — process turn 1, stop worker, restart from checkpoint, append turn 2 →
  turn 1 events not re-emitted.
- **A4** — after processing lines, `grok-sync.json` exists with monotonic
  `updates_offset`.
- **A5** — resume from checkpoint with `turn_index=1` → turn 2 user event has
  `extensions.grok_session.turn_index: 1`.
- **A6** — **PRIMARY**: worker starts without grok id/path; delayed `updates.jsonl`
  appears; discovery matches prompt; events emitted; `runner_session_id` persisted.

## Version

0.0.2

## Decision Tree

```
pkgs/agentsync/tests/grok-sync/
├── DOCTEST.md
├── SETUP.md
├── rapid-follow-ups/
│   └── single-user-each/              # A1: no duplicate user per prompt
├── ensure-idempotent/
│   └── single-worker/                 # A2: concurrent Ensure → one worker
├── checkpoint/
│   ├── resume-no-replay/              # A3: stop/restart; turn 1 not replayed
│   ├── file-written/                  # A4: grok-sync.json offset monotonic
│   └── turn-index-restored/           # A5: turn_index restored into converter
└── discover/
    └── bootstrap-delayed-session/     # A6: discovery bootstrap (PRIMARY)
```

Parameter ranking (most → least significant):

1. **Bootstrap mode** — known grok session vs discovery bootstrap (empty id/path)
2. **Worker cardinality** — single persistent worker vs overlapping tails
3. **Lifecycle** — continuous run vs stop/restart from checkpoint
4. **Append timing** — rapid multi-turn vs single batch vs delayed seed
5. **Checkpoint state** — fresh start vs pre-seeded `turn_index` / offset

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| A1 | `rapid-follow-ups/single-user-each` | Two turns while one worker runs; exactly one user message per prompt |
| A2 | `ensure-idempotent/single-worker` | Concurrent `EnsureGrokSync`; worker count 1; no duplicate events |
| A3 | `checkpoint/resume-no-replay` | Stop after turn 1; restart; turn 2 only; turn 1 not re-emitted |
| A4 | `checkpoint/file-written` | `grok-sync.json` exists; `updates_offset` advances monotonically |
| A5 | `checkpoint/turn-index-restored` | Resume `turn_index=1`; turn 2 user event stamped with index 1 |
| A6 | `discover/bootstrap-delayed-session` | Delayed grok session; discovery bootstrap; events + `runner_session_id` |

## How to Run

```sh
doctest vet ./pkgs/agentsync/tests/grok-sync
doctest test ./pkgs/agentsync/tests/grok-sync          # RED before implement

doctest test -v ./pkgs/agentsync/tests/grok-sync/rapid-follow-ups/single-user-each
doctest test -v ./pkgs/agentsync/tests/grok-sync/discover/bootstrap-delayed-session
```

Regression (must stay GREEN after implement):

```sh
doctest test ./pkgs/agenttty/tests/grok-updates-tail
doctest test ./cmd/agent-run/tests/grok-tty/sync-worker
```

```go
import (

	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentsync"
	"github.com/xhd2015/doctest/session"
)

type AppendSchedule struct {
	Delay time.Duration
	Lines []string
}

type DelayedGrokSeed struct {
	Delay         time.Duration
	GrokHome      string
	Workspace     string
	GrokSessionID string
	Prompt        string
	Lines         []string
}

type Request struct {
	TempDir      string
	SessionDir   string
	UpdatesPath  string
	GrokHome     string
	Runner       string
	SessionID    string
	GrokSessionID string
	Workspace    string
	InitialPrompt string
	SessionCreatedAt time.Time

	InitialLines               []string
	AppendSchedules            []AppendSchedule
	PostRestartAppendSchedules []AppendSchedule
	DelayedGrokSeed            *DelayedGrokSeed
	PreCheckpoint              *agentsync.GrokSyncState

	DiscoveryBootstrap bool
	ConcurrentEnsure   bool
	StopAfterTurn      int
	RestartAfterStop   bool

	WorkerStartDelay  time.Duration
	HoldAfterSchedule time.Duration
	DiscoveryTimeout  time.Duration
}

type Response struct {
	Events       []types.AgentEvent
	Checkpoint   agentsync.GrokSyncState
	CheckpointOK bool
	WorkerCount  int
	WorkerActive bool
	RunnerSessionID string
	StopErr      error
	EnsureErr    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.SessionDir == "" {
		req.SessionDir = filepath.Join(req.TempDir, "session")
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.SessionID == "" {
		req.SessionID = "sync-" + strings.ReplaceAll(t.Name(), "/", "_")
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

	discovery := req.DiscoveryBootstrap || (stringsTrim(req.GrokSessionID) == "" && stringsTrim(req.InitialPrompt) != "")
	if discovery {
		req.GrokSessionID = ""
		req.UpdatesPath = ""
	} else {
		if req.GrokSessionID == "" {
			req.GrokSessionID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		}
		if req.UpdatesPath == "" {
			req.UpdatesPath = filepath.Join(req.TempDir, "updates.jsonl")
		}
		if len(req.InitialLines) > 0 {
			if err := writeUpdatesJSONL(req.UpdatesPath, req.InitialLines...); err != nil {
				return nil, err
			}
		} else if err := os.WriteFile(req.UpdatesPath, nil, 0644); err != nil {
			return nil, fmt.Errorf("create updates.jsonl: %w", err)
		}
	}

	sink := agentsync.NewFileGrokSyncSink(req.SessionDir, req.GrokSessionID, req.UpdatesPath)
	if req.InitialPrompt != "" {
		if err := writeSessionMeta(req.SessionDir, req.Runner, req.SessionID, req.InitialPrompt, ""); err != nil {
			return nil, err
		}
	}
	if req.PreCheckpoint != nil {
		if err := sink.SaveCheckpoint(*req.PreCheckpoint); err != nil {
			return nil, err
		}
	}

	if req.DelayedGrokSeed != nil {
		seed := *req.DelayedGrokSeed
		if seed.GrokHome == "" {
			seed.GrokHome = req.GrokHome
		}
		if seed.Workspace == "" {
			seed.Workspace = req.Workspace
		}
		go func() {
			time.Sleep(seed.Delay)
			writeFakeGrokSessionDir(t, seed.GrokHome, seed.Workspace, seed.GrokSessionID, seed.Prompt, seed.Lines...)
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	createdAt := req.SessionCreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().Add(-2 * time.Second)
	}

	opts := agentsync.GrokSyncOptions{
		Runner:           req.Runner,
		SessionID:        req.SessionID,
		GrokSessionID:    req.GrokSessionID,
		UpdatesPath:      req.UpdatesPath,
		Workspace:        req.Workspace,
		GrokHome:         req.GrokHome,
		InitialPrompt:    req.InitialPrompt,
		SessionCreatedAt: createdAt,
		Sink:             sink,
	}

	var ensureErr error
	if req.ConcurrentEnsure {
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs <- agentsync.EnsureGrokSync(ctx, opts)
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				ensureErr = err
			}
		}
	} else {
		ensureErr = agentsync.EnsureGrokSync(ctx, opts)
	}

	startDelay := req.WorkerStartDelay
	if startDelay <= 0 {
		startDelay = 150 * time.Millisecond
	}
	time.Sleep(startDelay)

	resp := &Response{
		Events:    readEventsFromSessionDir(t, req.SessionDir),
		EnsureErr: ensureErr,
	}

	for _, sched := range req.AppendSchedules {
		if req.UpdatesPath == "" {
			if id := stringsTrim(req.GrokSessionID); id != "" {
				req.UpdatesPath = grokUpdatesPath(req.GrokHome, req.Workspace, id)
			}
		}
		time.Sleep(sched.Delay)
		if err := appendUpdatesJSONL(req.UpdatesPath, sched.Lines...); err != nil {
			return resp, err
		}
	}

	hold := req.HoldAfterSchedule
	if hold <= 0 {
		hold = 800 * time.Millisecond
	}
	if discovery {
		discTimeout := req.DiscoveryTimeout
		if discTimeout <= 0 {
			discTimeout = 8 * time.Second
		}
		resp.Events = waitForDiscoveryEvents(t, req, discTimeout)
	} else {
		time.Sleep(hold)
	}

	if req.StopAfterTurn > 0 || req.RestartAfterStop {
		resp.Events = waitForActionDoneCount(t, req.SessionDir, req.StopAfterTurn, 5*time.Second)
	}

	if req.StopAfterTurn > 0 {
		resp.StopErr = agentsync.StopGrokSync(req.Runner, req.SessionID)
		if cp, err := sink.LoadCheckpoint(); err == nil && cp.UpdatesPath != "" {
			resp.Checkpoint = cp
			resp.CheckpointOK = true
		}
		resp.WorkerCount = agentsync.GrokSyncWorkerCount()
		resp.WorkerActive = agentsync.GrokSyncWorkerActive(req.Runner, req.SessionID)
	}

	if req.RestartAfterStop {
		time.Sleep(200 * time.Millisecond)
		if cp, err := sink.LoadCheckpoint(); err == nil {
			opts.GrokSessionID = cp.GrokSessionID
			opts.UpdatesPath = cp.UpdatesPath
		}
		if err := agentsync.EnsureGrokSync(ctx, opts); err != nil && ensureErr == nil {
			ensureErr = err
			resp.EnsureErr = err
		}
		time.Sleep(startDelay)
		for _, sched := range req.PostRestartAppendSchedules {
			time.Sleep(sched.Delay)
			if err := appendUpdatesJSONL(req.UpdatesPath, sched.Lines...); err != nil {
				return resp, err
			}
		}
		time.Sleep(hold)
	}

	if !discovery {
		resp.Events = readEventsFromSessionDir(t, req.SessionDir)
	}
	if cp, err := sink.LoadCheckpoint(); err == nil && cp.UpdatesPath != "" {
		resp.Checkpoint = cp
		resp.CheckpointOK = true
	}
	resp.WorkerCount = agentsync.GrokSyncWorkerCount()
	resp.WorkerActive = agentsync.GrokSyncWorkerActive(req.Runner, req.SessionID)
	resp.RunnerSessionID = readRunnerSessionID(t, req.SessionDir)
	return resp, nil
}
```