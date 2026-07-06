# grok updates.jsonl tail Tests

Doc-style tests for `agenttty.TailUpdatesFromOffset` — event-driven tailing of grok
ACP `updates.jsonl` with bootstrap read from `startOffset` and `logs.WatchLine` for
subsequent appends until `ctx.Done()`.

# DSN (Domain Specific Notion)

**Participants**

- **updates.jsonl file** — JSONL under a grok session dir; each line is an ACP session
  update (`user_message_chunk`, `agent_message_chunk`, `tool_call`, `tool_call_update`,
  `turn_completed`, …).
- **TailUpdatesFromOffset** — reads from `startOffset` through current EOF (bootstrap),
  then watches for new appends until the caller cancels `ctx`.
- **grok_session.Converter** — coalesces wire chunks into canonical `types.AgentEvent`
  values; `turn_completed` emits `ActionDone` and increments `turn_index` but must
  **not** end the tail early.
- **Test harness** — writes synthetic ACP lines, starts tail in a goroutine, schedules
  delayed appends, collects emitted events, cancels context to stop.

**Behaviors**

- Bootstrap at `startOffset=0` replays all lines already present before watch starts.
- Bootstrap at `startOffset=EOF` skips bytes already on disk (stale-session semantics
  when `updatesTailStartOffset` returns file size).
- Mid-run appends after watch starts are converted and emitted while `ctx` is alive.
- After `turn_completed`, tail **continues** watching; a second turn's user/assistant
  chunks must appear before context cancellation (primary bug fix).

## Version

0.0.2

## Decision Tree

```
pkgs/agenttty/tests/grok-updates-tail/
├── DOCTEST.md
├── SETUP.md
├── multi-turn/
│   ├── SETUP.md
│   └── streams-second-turn-after-completed/   # PRIMARY: turn 2 after turn_completed
├── mid-run-append/
│   └── streams-appended-lines-before-cancel/  # Regression: delayed append while watching
└── bootstrap-offset/
    ├── SETUP.md
    ├── reads-preseeded-from-zero/             # startOffset=0 includes pre-seeded lines
    └── eof-skips-prior-content/               # startOffset=EOF skips stale bytes
```

Parameter ranking (most → least significant):

1. **Tail lifecycle** — continues until `ctx.Done()` vs erroneous exit on `turn_completed`
2. **Append timing** — pre-seeded bootstrap vs delayed mid-run vs post-`turn_completed`
3. **Start offset** — byte 0 vs EOF (skip stale)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `multi-turn/streams-second-turn-after-completed` | Turn 1 ends with `turn_completed`; turn 2 marker emitted before ctx cancel |
| 2 | `mid-run-append/streams-appended-lines-before-cancel` | Empty/minimal seed; delayed append streams marker while tail alive |
| 3 | `bootstrap-offset/reads-preseeded-from-zero` | `startOffset=0` emits pre-seeded user + assistant before watch |
| 4 | `bootstrap-offset/eof-skips-prior-content` | `startOffset=EOF` does not replay prior on-disk content |

## How to Run

```sh
doctest vet ./pkgs/agenttty/tests/grok-updates-tail
doctest test ./pkgs/agenttty/tests/grok-updates-tail
doctest test -v ./pkgs/agenttty/tests/grok-updates-tail/multi-turn/streams-second-turn-after-completed
doctest test -v ./pkgs/agenttty/tests/grok-updates-tail/mid-run-append/streams-appended-lines-before-cancel
doctest test -v ./pkgs/agenttty/tests/grok-updates-tail/bootstrap-offset/reads-preseeded-from-zero

# Regression via agent-run integration (sibling tree)
doctest vet ./cmd/agent-run/tests/grok-tty/run/streams-second-turn-after-completed
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-second-turn-after-completed
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-events-before-exit
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
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

type AppendSchedule struct {
	Delay time.Duration
	Lines []string
}

type Request struct {
	TempDir           string
	UpdatesPath       string
	StartOffset       int64
	StartOffsetAtEOF  bool
	InitialLines      []string
	AppendSchedules   []AppendSchedule
	TailStartDelay    time.Duration
	HoldAfterSchedule time.Duration
}

type Response struct {
	Events     []types.AgentEvent
	EventTexts []string
	TailErr    error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	req.TempDir = t.TempDir()
	updatesPath := req.UpdatesPath
	if updatesPath == "" {
		updatesPath = filepath.Join(req.TempDir, "updates.jsonl")
		req.UpdatesPath = updatesPath
	}

	if len(req.InitialLines) > 0 {
		if err := writeUpdatesJSONL(updatesPath, req.InitialLines...); err != nil {
			return nil, err
		}
	} else if err := os.WriteFile(updatesPath, nil, 0644); err != nil {
		return nil, fmt.Errorf("create updates.jsonl: %w", err)
	}

	startOffset := req.StartOffset
	if req.StartOffsetAtEOF {
		info, err := os.Stat(updatesPath)
		if err != nil {
			return nil, err
		}
		startOffset = info.Size()
	}

	ctx, cancel := context.WithCancel(context.Background())

	var mu sync.Mutex
	var events []types.AgentEvent
	emit := func(ev types.AgentEvent) error {
		mu.Lock()
		events = append(events, ev)
		mu.Unlock()
		return nil
	}

	tailDone := make(chan error, 1)
	go func() {
		tailDone <- agenttty.TailUpdatesFromOffset(ctx, updatesPath, startOffset, emit)
	}()

	startDelay := req.TailStartDelay
	if startDelay <= 0 {
		startDelay = 150 * time.Millisecond
	}
	time.Sleep(startDelay)

	for _, sched := range req.AppendSchedules {
		time.Sleep(sched.Delay)
		if err := appendUpdatesJSONL(updatesPath, sched.Lines...); err != nil {
			cancel()
			return nil, err
		}
	}

	hold := req.HoldAfterSchedule
	if hold <= 0 {
		hold = 600 * time.Millisecond
	}
	time.Sleep(hold)

	cancel()
	tailErr := <-tailDone

	resp := &Response{Events: events, TailErr: tailErr}
	for _, ev := range events {
		if text := strings.TrimSpace(ev.Text); text != "" {
			resp.EventTexts = append(resp.EventTexts, text)
		}
	}
	return resp, nil
}
```