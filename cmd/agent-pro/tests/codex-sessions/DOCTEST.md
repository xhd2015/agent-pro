# Codex Sessions Tests

Doc-style tests for `agent/codex/sessions`, which lists Codex CLI rollout
sessions in a unified grok-shaped table, shows session info (evolved from brief),
and prints human-readable logs from JSONL transcripts under a synthetic
`CodexHome`.

# DSN (Domain Specific Notion)

Codex CLI writes session transcripts as JSONL rollout files under
`{codexHome}/sessions/YYYY/MM/DD/`. Each file begins with `session_meta` and
continues with `event_msg` and `response_item` events.

The sessions package walks those files to discover sessions, parses metadata,
and sorts by last activity. List table output matches the unified shape:
`SESSION ID | LAST ACTIVE | TITLE | MSGS | CWD`, where `MSGS` is the count of
displayable events in the rollout and `LAST ACTIVE` uses relative times with a
fixed `now` clock.

For session info (`agent-pro codex session info <id>`), `Info` aggregates
metadata, status, line count, recent displayable messages, rollout file path,
and token totals from `token_count` events; `FormatInfoText` renders
grok-style key-value blocks. The legacy `Brief` API remains tested under
`brief/` for package-level regression.

For session log (`agent-pro codex session log <id> [--tail N]`), `PrintLog`
normalizes rollout events into the compact trace print pipeline
(`print.FormatTraceLine`) so output matches other agent trace tools.

The test harness builds a temporary Codex home, writes minimal rollout
fixtures, and calls the package API directly (no real `~/.codex`).

## Version

0.0.2

## Decision Tree

```
operation?
├── list/
│   ├── default-limit-20/     25 fixtures → List returns 20 newest
│   ├── custom-limit/         limit=3 → 3 sessions
│   ├── sorted-newest-first/  verify descending last-active order
│   ├── table-format/         FormatListTable grok shape: LAST ACTIVE, TITLE, MSGS, CWD
│   ├── table-with-msgs/      MSGS column from displayable event count
│   ├── json-format/          FormatListJSON emits sessions array
│   ├── list-within-time-budget/  100 large rollouts, limit 20 → List <= 500ms
│   └── empty/                no rollout files → "No sessions found"
├── info/
│   ├── known-session/        FormatInfoText: status, messages, rollout path, token totals
│   ├── last-three-messages/  Info shows last 3 displayable events
│   ├── no-tokens/            no token_count events → no Tokens section
│   └── unknown-session/      missing UUID → codex session not found
├── brief/                    (package API — legacy Brief/FormatBrief*)
│   ├── last-three-messages/
│   ├── json-format/
│   └── unknown-session/
└── log/                      (session log — PrintLog + --tail)
    ├── exec-and-message/
    ├── apply-patch/
    ├── encrypted-reasoning/
    ├── skips-noise/
    ├── tail-last-n/
    └── unknown-session/
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/default-limit-20` | 25 rollout files, default limit 20 → returns 20 newest sessions |
| 2 | `list/custom-limit` | 5 sessions, limit=3 → exactly 3 returned |
| 3 | `list/sorted-newest-first` | 3 sessions with known timestamps → newest first |
| 4 | `list/table-format` | Table has SESSION ID, LAST ACTIVE, TITLE, MSGS, CWD with relative times |
| 5 | `list/table-with-msgs` | Five displayable events → MSGS shows `5` |
| 6 | `list/json-format` | JSON list includes id, started_at, cwd, path per session |
| 7 | `list/list-within-time-budget` | 200 rollouts × 400 events, limit 20 → List within 500ms |
| 8 | `list/empty` | Empty sessions tree → "No sessions found" |
| 8 | `info/known-session` | Info returns status, line count, rollout path, token totals |
| 9 | `info/last-three-messages` | Five displayable events → info shows last 3 |
| 10 | `info/no-tokens` | Session without token_count → info succeeds, no Tokens section |
| 11 | `info/unknown-session` | Unknown full UUID → error |
| 12 | `brief/last-three-messages` | Brief API: 5 events → last 3 in brief |
| 13 | `brief/json-format` | Brief JSON includes recent_messages |
| 14 | `brief/unknown-session` | Brief API: unknown UUID → error |
| 15 | `log/exec-and-message` | function_call + agent_message → RUN and ASSISTANT in log |
| 16 | `log/apply-patch` | custom_tool_call apply_patch → EDIT in log |
| 17 | `log/encrypted-reasoning` | reasoning with encrypted_content only → REASONING [Redacted] |
| 18 | `log/skips-noise` | Non-displayable events produce empty log output |
| 19 | `log/tail-last-n` | 5 displayable events, tail=2 → output has last 2 only |
| 20 | `log/unknown-session` | Unknown session ID → error from Find/PrintLog |

## How to Run

```sh
cd knowledge_research/external/agent-pro
doctest vet ./cmd/agent-pro/tests/codex-sessions
doctest test -v ./cmd/agent-pro/tests/codex-sessions
```

```go
import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
)

type Request struct {
	Operation string    // "list", "info", "brief", "log"
	CodexHome string
	SessionID string
	Limit     int       // 0 → default 20 for list
	LastN     int       // 0 → default 3 for brief/info recent messages
	Tail      int       // 0 → full log; >0 → last N displayable events
	Format    string    // "table", "json", "" (log writes raw text to Output)
	Now       time.Time // fixed clock for relative times in formatters
}

type Response struct {
	Sessions []sessions.Session
	Info     *sessions.SessionInfo
	Brief    *sessions.SessionBrief
	Output   string
	JSON     []byte
	Elapsed  time.Duration // list operation wall time
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	lastN := req.LastN
	if lastN == 0 {
		lastN = 3
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	switch req.Operation {
	case "list":
		start := time.Now()
		list, err := sessions.List(req.CodexHome, limit)
		resp.Elapsed = time.Since(start)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Sessions = list
		switch req.Format {
		case "json":
			data, err := sessions.FormatListJSON(list)
			if err != nil {
				resp.Err = err
				return resp, nil
			}
			resp.JSON = data
		default:
			resp.Output = sessions.FormatListTable(list, req.CodexHome, now)
		}
	case "info":
		info, err := sessions.Info(req.CodexHome, req.SessionID, lastN)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Info = info
		resp.Output = sessions.FormatInfoText(info, req.CodexHome, now)
	case "brief":
		brief, err := sessions.Brief(req.CodexHome, req.SessionID, lastN)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Brief = brief
		switch req.Format {
		case "json":
			data, err := sessions.FormatBriefJSON(brief)
			if err != nil {
				resp.Err = err
				return resp, nil
			}
			resp.JSON = data
		default:
			resp.Output = sessions.FormatBriefText(brief, filepath.Dir(req.CodexHome))
		}
	case "log":
		path, err := sessions.Find(req.CodexHome, req.SessionID)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		var buf bytes.Buffer
		if err := sessions.PrintLog(path, &buf, req.Tail); err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Output = buf.String()
	default:
		resp.Err = fmt.Errorf("unknown operation: %s", req.Operation)
	}
	return resp, nil
}
```