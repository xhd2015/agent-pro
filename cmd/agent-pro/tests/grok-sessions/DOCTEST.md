# Grok Sessions Tests

Doc-style tests for `agent/grok/sessions`, which lists Grok CLI sessions from
`summary.json` files under a synthetic `GROK_HOME`, formats them as a table
with relative last-active times and message counts, and shows detailed session
info from `summary.json`, `signals.json`, and the session directory layout.

# DSN (Domain Specific Notion)

Grok CLI stores session metadata under `{grokHome}/sessions/{encoded_cwd}/{uuid}/`.
Each session directory contains `summary.json` with session id, cwd, title,
message counts, model/agent metadata, and activity timestamps. Optional files
include `signals.json` (token usage), `updates.jsonl` (conversation log), and
`prompt_context.json`. The encoded cwd directory name is `url.PathEscape(abs_cwd)`.

The sessions package walks `{grokHome}/sessions/*/<uuid>/summary.json`, parses
each file, skips malformed entries, sorts by `last_active_at` descending (tie-break
by session id), truncates to the requested limit, and formats results as a table.
`FormatListTable` shows `num_chat_messages` in an `MSGS` column and accepts a
fixed `now` clock so relative times (`just now`, `5m ago`, `2h ago`) are
deterministic in tests.

For session detail, `Find` locates a session by exact UUID (no prefix matching),
`Info` aggregates summary fields, filesystem paths, and token usage from
`signals.json`, and `FormatInfoText` renders key-value blocks for the CLI
`agent-pro grok session info <id>` command.

The test harness builds a temporary Grok home, writes minimal fixtures under
encoded cwd paths, and calls the package API directly (no real `~/.grok`).

## Version

0.0.2

## Decision Tree

```
operation?
├── list/
│   ├── default-limit-20/     25 fixtures → List returns 20 newest
│   ├── custom-limit/         5 fixtures, limit=3 → exactly 3
│   ├── sorted-newest-first/  3 sessions with known timestamps → descending order
│   ├── table-format/         FormatListTable has SESSION ID, LAST ACTIVE, TITLE, CWD
│   ├── empty/                empty sessions tree → "No sessions found"
│   ├── malformed-skipped/    mix of valid + corrupt summary.json → only valid returned
│   ├── table-with-msgs/      MSGS column; num_chat_messages=42 → "42"
│   └── message-count-zero/   empty title, num_chat_messages=0 → "0"
└── info/
    ├── known-session/        full summary + file paths + tokens from signals.json
    ├── unknown-session/      missing UUID → grok session not found
    ├── no-signals/           no signals.json → info succeeds, no Tokens section
    └── untitled-session/     empty title → (untitled), num_chat_messages=1
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/default-limit-20` | 25 summary.json files, default limit 20 → returns 20 newest sessions |
| 2 | `list/custom-limit` | 5 sessions, limit=3 → exactly 3 returned |
| 3 | `list/sorted-newest-first` | 3 sessions with known timestamps → newest first |
| 4 | `list/table-format` | Table output contains SESSION ID, LAST ACTIVE, TITLE, CWD with relative times |
| 5 | `list/empty` | Empty sessions tree → "No sessions found" |
| 6 | `list/malformed-skipped` | Valid + corrupt summary.json → only valid sessions returned |
| 7 | `list/table-with-msgs` | Table has MSGS column; fixture with num_chat_messages=42 shows `42` |
| 8 | `list/message-count-zero` | Empty-title fixture with num_chat_messages=0 shows `0` |
| 9 | `info/known-session` | Info returns summary fields, file paths, and token usage from signals.json |
| 10 | `info/unknown-session` | Unknown full UUID → `grok session not found` error |
| 11 | `info/no-signals` | Session without signals.json → info succeeds, no Tokens section |
| 12 | `info/untitled-session` | Empty title → shows `(untitled)`, num_chat_messages=1 |

## How to Run

```sh
cd knowledge_research/external/agent-pro
doctest vet ./cmd/agent-pro/tests/grok-sessions
doctest test -v ./cmd/agent-pro/tests/grok-sessions
```

```go
import (
	"fmt"
	"testing"
	"time"

	sessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
)

type Request struct {
	Operation string    // "list" or "info"; default "list"
	GrokHome  string
	SessionID string    // info only; exact UUID
	Limit     int       // 0 → default 20 for list
	Now       time.Time // fixed clock for relative times in formatters
}

type Response struct {
	Sessions []sessions.Session
	Info     *sessions.SessionInfo
	Output   string
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
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

		list, err := sessions.List(req.GrokHome, limit)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Sessions = list
		resp.Output = sessions.FormatListTable(list, req.GrokHome, now)
	case "info":
		info, err := sessions.Info(req.GrokHome, req.SessionID)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Info = info
		resp.Output = sessions.FormatInfoText(info, req.GrokHome, now)
	default:
		resp.Err = fmt.Errorf("unknown operation: %s", op)
	}
	return resp, nil
}
```