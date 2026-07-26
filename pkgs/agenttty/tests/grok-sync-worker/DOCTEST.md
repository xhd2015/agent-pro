# Grok Sync Worker Tests

Doc-style tests for the persistent **grok sync worker** — a single per-session
goroutine that tails `updates.jsonl`, converts ACP lines via `grok_session.Converter`,
appends canonical events, and checkpoints byte offset + `turn_index` in
`grok-sync.json`. Replaces overlapping `startGrokFollowUpEventTail` watchers that
duplicate user/assistant/done events on rapid follow-ups.

# DSN (Domain Specific Notion)

**Participants**

- **updates.jsonl** — grok on-disk ACP stream (`user_message_chunk`, `agent_message_chunk`,
  `turn_completed`, …) under `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/`.
- **GrokSyncWorker** — long-lived tail loop: `TailUpdatesFromOffset` from checkpoint
  offset; on each complete line processed + events emitted → persist checkpoint.
- **grok_session.Converter** — coalesces wire chunks; `turn_index` restored from
  checkpoint on worker resume.
- **GrokSyncSink** — storage-agnostic adapter: `AppendEvent`, `LoadCheckpoint`,
  `SaveCheckpoint`, `OnTurnCompleted`.
- **grok-sync.json** — checkpoint at `sessions/grok-tty/<id>/grok-sync.json` with
  `updates_offset` and `turn_index`; written **after** each `events.jsonl` append.
- **EnsureGrokSync** — idempotent entry point; mutex registry guarantees one active
  worker per `(runner, sessionID)`.
- **Test harness** — synthetic ACP lines (reuse builders from `grok-updates-tail`),
  temp session dir as fake agent-run home, `NewFileGrokSyncSink` implementing `GrokSyncSink`.

**Behaviors**

- **S1** — two turns appended while one worker runs → exactly one user `message` per
  distinct prompt text (PRIMARY duplicate-bug guard).
- **S2** — concurrent `EnsureGrokSync` calls → single active worker; no duplicate
  events for the same updates line.
- **S3** — process turn 1, stop worker, restart from checkpoint, append turn 2 →
  turn 1 events not re-emitted.
- **S4** — after processing lines, `grok-sync.json` exists with monotonic
  `updates_offset`.
- **S5** — resume from checkpoint with `turn_index=1` → turn 2 user event has
  `extensions.grok_session.turn_index: 1`.

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/grok-sync-worker/
├── DOCTEST.md
├── SETUP.md
├── rapid-follow-ups/
│   └── single-user-each/              # S1: PRIMARY — no duplicate user per prompt
├── ensure-idempotent/
│   └── single-worker/                 # S2: concurrent Ensure → one worker
└── checkpoint/
    ├── resume-no-replay/              # S3: stop/restart; turn 1 not replayed
    ├── file-written/                  # S4: grok-sync.json offset monotonic
    └── turn-index-restored/           # S5: turn_index restored into converter
```

Parameter ranking (most → least significant):

1. **Worker cardinality** — single persistent worker vs overlapping per-message tails
2. **Lifecycle** — continuous run vs stop/restart from checkpoint
3. **Append timing** — rapid multi-turn vs single batch
4. **Checkpoint state** — fresh start vs pre-seeded `turn_index` / offset

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| S1 | `rapid-follow-ups/single-user-each` | Two turns while one worker runs; exactly one user message per prompt |
| S2 | `ensure-idempotent/single-worker` | Concurrent `EnsureGrokSync`; worker count 1; no duplicate events |
| S3 | `checkpoint/resume-no-replay` | Stop after turn 1; restart; turn 2 only; turn 1 not re-emitted |
| S4 | `checkpoint/file-written` | `grok-sync.json` exists; `updates_offset` advances monotonically |
| S5 | `checkpoint/turn-index-restored` | Resume `turn_index=1`; turn 2 user event stamped with index 1 |

## How to Run

```sh
doctest vet ./pkgs/agenttty/tests/grok-sync-worker
doctest test ./pkgs/agenttty/tests/grok-sync-worker   # RED before implement

doctest test -v ./pkgs/agenttty/tests/grok-sync-worker/rapid-follow-ups/single-user-each
doctest test -v ./pkgs/agenttty/tests/grok-sync-worker/ensure-idempotent/single-worker
doctest test -v ./pkgs/agenttty/tests/grok-sync-worker/checkpoint/resume-no-replay
```

Regression (must stay GREEN):

```sh
doctest test ./pkgs/agenttty/tests/grok-updates-tail
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-second-turn-after-completed
```

```go
import (

	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/doctest/session"
)

type AppendSchedule struct {
	Delay time.Duration
	Lines []string
}

type Request struct {
	TempDir      string
	SessionDir   string
	UpdatesPath  string
	Runner       string
	SessionID    string
	GrokSessionID string
	Workspace    string

	InitialLines               []string
	AppendSchedules            []AppendSchedule
	PostRestartAppendSchedules []AppendSchedule
	PreCheckpoint              *agenttty.GrokSyncState

	ConcurrentEnsure bool
	StopAfterTurn    int // 0 = no stop; 1 = stop after first turn_completed
	RestartAfterStop bool

	WorkerStartDelay  time.Duration
	HoldAfterSchedule time.Duration
}

type Response struct {
	Events       []types.AgentEvent
	Checkpoint   agenttty.GrokSyncState
	CheckpointOK bool
	WorkerCount  int
	WorkerActive bool
	StopErr      error
	EnsureErr    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.UpdatesPath == "" {
		req.UpdatesPath = filepath.Join(req.TempDir, "updates.jsonl")
	}
	if req.SessionDir == "" {
		req.SessionDir = filepath.Join(req.TempDir, "session")
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.SessionID == "" {
		req.SessionID = "sync-worker-test"
	}
	if req.GrokSessionID == "" {
		req.GrokSessionID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	}
	if req.Workspace == "" {
		req.Workspace = req.TempDir
	}

	if len(req.InitialLines) > 0 {
		if err := writeUpdatesJSONL(req.UpdatesPath, req.InitialLines...); err != nil {
			return nil, err
		}
	} else if err := os.WriteFile(req.UpdatesPath, nil, 0644); err != nil {
		return nil, fmt.Errorf("create updates.jsonl: %w", err)
	}

	sink := agenttty.NewFileGrokSyncSink(req.SessionDir, req.GrokSessionID, req.UpdatesPath)
	if req.PreCheckpoint != nil {
		if err := sink.SaveCheckpoint(*req.PreCheckpoint); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := agenttty.GrokSyncOptions{
		Runner:        req.Runner,
		SessionID:     req.SessionID,
		GrokSessionID: req.GrokSessionID,
		UpdatesPath:   req.UpdatesPath,
		Workspace:     req.Workspace,
		Sink:          sink,
	}

	var ensureErr error
	if req.ConcurrentEnsure {
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs <- agenttty.EnsureGrokSync(ctx, opts)
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
		ensureErr = agenttty.EnsureGrokSync(ctx, opts)
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
		time.Sleep(sched.Delay)
		if err := appendUpdatesJSONL(req.UpdatesPath, sched.Lines...); err != nil {
			return resp, err
		}
	}

	hold := req.HoldAfterSchedule
	if hold <= 0 {
		hold = 800 * time.Millisecond
	}
	time.Sleep(hold)

	if req.StopAfterTurn > 0 || req.RestartAfterStop {
		resp.Events = waitForActionDoneCount(t, req.SessionDir, req.StopAfterTurn, 5*time.Second)
	}

	if req.StopAfterTurn > 0 {
		resp.StopErr = agenttty.StopGrokSync(req.Runner, req.SessionID)
		if cp, err := sink.LoadCheckpoint(); err == nil && cp.UpdatesPath != "" {
			resp.Checkpoint = cp
			resp.CheckpointOK = true
		}
		resp.WorkerCount = agenttty.GrokSyncWorkerCount()
		resp.WorkerActive = agenttty.GrokSyncWorkerActive(req.Runner, req.SessionID)
	}

	if req.RestartAfterStop {
		time.Sleep(200 * time.Millisecond)
		if err := agenttty.EnsureGrokSync(ctx, opts); err != nil && ensureErr == nil {
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

	resp.Events = readEventsFromSessionDir(t, req.SessionDir)
	if cp, err := sink.LoadCheckpoint(); err == nil && cp.UpdatesPath != "" {
		resp.Checkpoint = cp
		resp.CheckpointOK = true
	}
	resp.WorkerCount = agenttty.GrokSyncWorkerCount()
	resp.WorkerActive = agenttty.GrokSyncWorkerActive(req.Runner, req.SessionID)
	return resp, nil
}
```