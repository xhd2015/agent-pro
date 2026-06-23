# Codex Sessions Tests

Doc-style tests for `agent/codex/sessions`, which lists Codex CLI rollout
sessions, shows brief summaries, and prints human-readable logs from JSONL
transcripts under a synthetic `CodexHome`.

# DSN (Domain Specific Notion)

Codex CLI writes session transcripts as JSONL rollout files under
`{codexHome}/sessions/YYYY/MM/DD/`. Each file begins with `session_meta` and
continues with `event_msg` and `response_item` events.

The sessions package walks those files to discover sessions, parses metadata,
and sorts by start time. For brief view it counts lines, infers status from
task lifecycle events, and collects the last few displayable messages. For
full log view it normalizes rollout events into the compact trace print
pipeline (`print.FormatTraceLine`) so output matches other agent trace tools.

The test harness builds a temporary Codex home, writes minimal rollout
fixtures, and calls the package API directly (no real `~/.codex`).

## Version

0.0.2

## Decision Tree

```
operation?
├── list
│   ├── default-limit-20/     25 fixtures → List returns 20 newest
│   ├── custom-limit/         limit=3 → 3 sessions
│   ├── sorted-newest-first/  verify descending started_at order
│   ├── table-format/         FormatListTable has SESSION ID, STARTED, CWD
│   ├── json-format/          FormatListJSON emits sessions array
│   └── empty/                no rollout files → "No sessions found"
├── brief
│   ├── last-three-messages/  5 displayable events → last 3 in brief
│   ├── json-format/          FormatBriefJSON includes recent_messages
│   └── unknown-session/      missing UUID → codex session not found
└── log
    ├── exec-and-message/     exec_command + agent_message → RUN + ASSISTANT
    ├── apply-patch/          apply_patch custom_tool_call → EDIT
    ├── encrypted-reasoning/  encrypted reasoning only → REASONING [Redacted]
    ├── skips-noise/          session_meta, token_count → no output
    ├── tail-last-n/          5 events, tail=2 → only last 2 in log output
    └── unknown-session/      missing UUID → error
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `list/default-limit-20` | 25 rollout files, default limit 20 → returns 20 newest sessions |
| 2 | `list/custom-limit` | 5 sessions, limit=3 → exactly 3 returned |
| 3 | `list/sorted-newest-first` | 3 sessions with known timestamps → newest first |
| 4 | `list/table-format` | Table output contains SESSION ID, STARTED, CWD headers |
| 5 | `list/json-format` | JSON list includes id, started_at, cwd, path per session |
| 6 | `list/empty` | Empty sessions tree → "No sessions found" |
| 7 | `brief/last-three-messages` | 5 displayable events → brief shows last 3 chronologically |
| 8 | `brief/json-format` | Brief JSON includes recent_messages with kind, text, formatted |
| 9 | `brief/unknown-session` | Unknown full UUID → error |
| 10 | `log/exec-and-message` | function_call + agent_message → RUN and ASSISTANT in log |
| 11 | `log/apply-patch` | custom_tool_call apply_patch → EDIT in log |
| 12 | `log/encrypted-reasoning` | reasoning with encrypted_content only → REASONING [Redacted] |
| 13 | `log/skips-noise` | Non-displayable events produce empty log output |
| 14 | `log/tail-last-n` | 5 displayable events, tail=2 → output has last 2 only |
| 15 | `log/unknown-session` | Unknown session ID → error from Find/PrintLog |

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

	sessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
)

type Request struct {
	Operation string // "list", "brief", "log"
	CodexHome string
	SessionID string
	Limit     int    // 0 → default 20 for list
	LastN     int    // 0 → default 3 for brief
	Tail      int    // 0 → full log; >0 → last N displayable events
	Format    string // "table", "json", "" (log writes raw text to Output)
}

type Response struct {
	Sessions []sessions.Session
	Brief    *sessions.SessionBrief
	Output   string
	JSON     []byte
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

	switch req.Operation {
	case "list":
		list, err := sessions.List(req.CodexHome, limit)
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
			resp.Output = sessions.FormatListTable(list, req.CodexHome)
		}
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