# WatchEvents Status Decoupling Tests

Doc-style tests for `pkgs/agentevents.WatchEvents` — tail `events.jsonl` until client
`ctx` is cancelled. Session `meta.status` must not stop the watch early.

# DSN (Domain Specific Notion)

**Participants**

- **File store (`pkgs/agentstorage`)** — `AGENT_RUN_HOME/sessions/<runner>/<id>/events.jsonl`
  and `meta.json` with `status` field (`running`, `finished`, …).
- **WatchEvents** — reads from byte offset, calls `onLine` for each new NDJSON row, uses
  `logs.WatchLine` until `ctx.Done()`. Must **not** return early when status is `finished`.
- **Test harness** — seeds finished session, starts `WatchEvents` in goroutine, appends
  a new event while watch is alive, collects received lines.

**Behaviors**

- **C1 finished append** — session already `finished`; append while watch ctx alive →
  `onLine` receives the new row.

## Version

0.0.2

## Decision Tree

```
pkgs/agentevents/tests/watchevents/
├── DOCTEST.md
├── SETUP.md
└── finished-status/
    ├── SETUP.md
    └── delivers-appended-line-while-watching/   # C1
```

Parameter ranking (most → least significant):

1. **Session status** — `finished` (must not gate) vs `running` (baseline, covered elsewhere).
2. **Append timing** — while watch ctx alive vs before watch starts.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `finished-status/delivers-appended-line-while-watching` | C1: `WatchEvents` on finished session delivers appended line before ctx cancel |

## How to Run

```sh
doctest vet ./pkgs/agentevents/tests/watchevents
doctest test ./pkgs/agentevents/tests/watchevents
doctest test -v ./pkgs/agentevents/tests/watchevents/finished-status/delivers-appended-line-while-watching
```

```go
import (

	"context"
	"fmt"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Home      string
	TempDir   string
	Runner    string
	SessionID string
	AfterOffset int64
	AppendDelay time.Duration
	WatchHold   time.Duration
	AppendText  string
}

type Response struct {
	ReceivedLines []string
	ReceivedTexts []string
	GotAppended   bool
	WatchErr      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runWatchEventsProbe(t, req)
}
```