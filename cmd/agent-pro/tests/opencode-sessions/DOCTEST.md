# OpenCode Sessions Tests

Doc-style tests for `agent/opencode/sessions`, which lists OpenCode CLI sessions
from JSON files under `storage/session/{projectID}/`, counts messages from
`storage/message/{sessionID}/`, and shows detailed session info with token and
cost totals from message files under a synthetic OpenCode data directory.

# DSN (Domain Specific Notion)

OpenCode stores session metadata as JSON under
`{dataDir}/storage/session/{projectID}/ses_*.json`. Each file contains session
`id`, `title`, `directory` (cwd), and `time.updated` (epoch milliseconds) for
last activity. Message transcripts live as individual JSON files under
`{dataDir}/storage/message/{sessionID}/msg_*.json`; file count drives the list
table `MSGS` column, and summed `tokens` / `cost` fields feed session info.

The sessions package walks session JSON files, skips malformed entries, sorts by
`time.updated` descending (tie-break by session id), truncates to the requested
limit, and formats results as a unified table matching grok/codex shape.
`FormatListTable` shows message file counts in `MSGS`, relative last-active
times (`just now`, `5m ago`, `2h ago`), and `~`-shortened cwd paths with a
fixed `now` clock for deterministic tests.

For session detail, `Find` locates a session by exact id (no prefix matching),
`Info` aggregates session JSON, message directory paths, per-message token/cost
totals, and `FormatInfoText` renders key-value blocks for
`agent-pro opencode session info <id>`.

The test harness builds a temporary OpenCode data dir, writes minimal fixtures,
and calls the package API directly (no real `~/.local/share/opencode`).

## Version

0.0.2

## Decision Tree

```
operation?
├── list/
│   ├── default-limit-20/     25 fixtures → List returns 20 newest
│   ├── sorted-newest-first/  3 sessions with known time.updated → descending order
│   ├── table-format/         FormatListTable has SESSION ID, LAST ACTIVE, TITLE, MSGS, CWD
│   └── empty/                empty storage tree → "No sessions found"
└── info/
    ├── known-session/        session JSON + 3 messages → paths, MSGS, tokens, cost
    ├── unknown-session/      missing session id → opencode session not found
    └── no-messages/          session with no message dir → MSGS=0, no Tokens section
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/default-limit-20` | 25 session JSON files, default limit 20 → returns 20 newest sessions |
| 2 | `list/sorted-newest-first` | 3 sessions with known `time.updated` → newest first |
| 3 | `list/table-format` | Table output contains SESSION ID, LAST ACTIVE, TITLE, MSGS, CWD with relative times |
| 4 | `list/empty` | Empty storage tree → "No sessions found" |
| 5 | `info/known-session` | Info returns session fields, message paths, summed tokens and cost |
| 6 | `info/unknown-session` | Unknown session id → `opencode session not found` error |
| 7 | `info/no-messages` | Session without message files → info succeeds, MSGS=0, no Tokens section |

## How to Run

```sh
cd knowledge_research/external/agent-pro
doctest vet ./cmd/agent-pro/tests/opencode-sessions
doctest test -v ./cmd/agent-pro/tests/opencode-sessions
```

```go
import (

	"fmt"
	"testing"
	"time"

	sessions "github.com/xhd2015/agent-pro/agent/opencode/sessions"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Operation string    // "list" or "info"; default "list"
	DataDir   string    // synthetic ~/.local/share/opencode root
	SessionID string    // info only; exact session id
	Limit     int       // 0 → default 20 for list
	Now       time.Time // fixed clock for relative times in formatters
}

type Response struct {
	Sessions []sessions.Session
	Info     *sessions.SessionInfo
	Output   string
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	op := req.Operation
	if op == "" {
		op = "list"
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	switch op {
	case "list":
		limit := req.Limit
		if limit == 0 {
			limit = 20
		}

		list, err := sessions.List(req.DataDir, limit)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Sessions = list
		resp.Output = sessions.FormatListTable(list, req.DataDir, now)
	case "info":
		info, err := sessions.Info(req.DataDir, req.SessionID)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Info = info
		resp.Output = sessions.FormatInfoText(info, req.DataDir, now)
	default:
		resp.Err = fmt.Errorf("unknown operation: %s", op)
	}
	return resp, nil
}
```