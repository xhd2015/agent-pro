# Grok Sessions Tests

Doc-style tests for `agent/grok/sessions`, which lists Grok CLI sessions from
`summary.json` files under a synthetic `GROK_HOME`, formats them as a table
with relative last-active times and message counts, optionally filters by
literal case-insensitive grep over session JSON content (with indented hit
lines and optional ANSI color), and shows detailed session info from
`summary.json`, `signals.json`, and the session directory layout.

# DSN (Domain Specific Notion)

Grok CLI stores session metadata under `{grokHome}/sessions/{encoded_cwd}/{uuid}/`.
Each session directory contains `summary.json` with session id, cwd, title
(`generated_title` / `session_summary`), message counts, model/agent metadata,
and activity timestamps. Optional files include `chat_history.jsonl` (message
log by `type`: system/user/assistant/reasoning/tool_result), `signals.json`
(token usage), `updates.jsonl` (wire log — not searched in v1), and
`prompt_context.json`. The encoded cwd directory name is `url.PathEscape(abs_cwd)`.

The sessions package walks `{grokHome}/sessions/*/<uuid>/summary.json`, parses
each file, skips malformed entries, sorts by `last_active_at` descending
(tie-break by session id), truncates to the requested limit, and formats
results as a table. `FormatListTable` shows `num_chat_messages` in an `MSGS`
column and accepts a fixed `now` clock so relative times (`just now`,
`5m ago`, `2h ago`) are deterministic in tests.

When a grep pattern is set, only sessions with at least one case-insensitive
literal hit in `summary.json` (title-ish fields, cwd, model, agent) or
`chat_history.jsonl` (message text by type/part) are listed. Limit applies
**after** filtering. Under each matching session row, the formatter prints up
to five indented hit lines `  <file>:<line>:<part>: <snippet>`, then
`  ... and N more matches` when remaining hits exist. Each hit **snippet** is
built from the matched field after whitespace collapse (runs of spaces/tabs/
newlines → a single ASCII space, with leading/trailing trim): the window is at
most **1024 runes** (Unicode code points), centered ~50/50 around the first
literal match, with ASCII `...` (3 runes) on each truncated side; when the
match itself is ≥1024 runes the snippet is the first 1024 runes of the match
only. `MatchStart`/`MatchLen` are byte offsets into the final snippet for
coloring the exact match. Color mode `never`/`always`/`auto` controls ANSI
styling of hit lines (tests force `never` or `always`).

For session detail, `Find` locates a session by exact UUID (no prefix matching),
`Info` aggregates summary fields, filesystem paths, and token usage from
`signals.json`, and `FormatInfoText` renders key-value blocks for the CLI
`agent-pro grok session info <id>` command.

The test harness builds a temporary Grok home, writes minimal fixtures under
encoded cwd paths (including optional `chat_history.jsonl`), and calls the
package API directly (no real `~/.grok`).

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
│   ├── message-count-zero/   empty title, num_chat_messages=0 → "0"
│   └── grep/                 --grep pattern filter + indented hits + color
│       ├── title-match/         pattern only in generated_title → summary.json:1:title
│       ├── chat-user-match/     pattern only in chat user line; non-match omitted
│       ├── multi-hit-lines/     summary + chat hits → ordered indented lines
│       ├── hit-cap-five/        >5 hits → 5 lines + "... and N more matches"
│       ├── no-match/            pattern matches nothing → "No sessions found"
│       ├── no-grep-no-hits/     without grep → no indented hit lines (regression)
│       ├── limit-after-filter/  limit applies after grep; newest matching N
│       ├── case-insensitive/    pattern case differs from content → still matches
│       ├── color-always/        Color=always → ANSI around hit styling
│       ├── color-never/         Color=never → no ANSI escapes
│       ├── snippet-window/      long field + mid match → ≤1024 runes, leading+trailing ...
│       ├── snippet-short/       short field → full text, no snippet ellipsis
│       └── snippet-match-only/  match ≥1024 runes → first 1024 of match only
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
| 9 | `list/grep/title-match` | Grep hits `generated_title` only → session listed with `summary.json:1:title` hit |
| 10 | `list/grep/chat-user-match` | Grep hits chat user line only; non-matching sessions omitted |
| 11 | `list/grep/multi-hit-lines` | Several hits (summary + chat) → multiple indented lines in order |
| 12 | `list/grep/hit-cap-five` | More than 5 hits → exactly 5 hit lines + `... and N more matches` |
| 13 | `list/grep/no-match` | Pattern matches nothing → `No sessions found` |
| 14 | `list/grep/no-grep-no-hits` | Without grep, chat_history present → no indented hit lines |
| 15 | `list/grep/limit-after-filter` | Multiple matches, limit=2 → exactly 2 newest matching sessions |
| 16 | `list/grep/case-insensitive` | Pattern case differs from content → still matches |
| 17 | `list/grep/color-always` | Color=always → ANSI escape sequences around match styling |
| 18 | `list/grep/color-never` | Color=never → output has no ANSI escapes |
| 19 | `list/grep/snippet-window` | Long tool_result mid-match → snippet ≤1024 runes with leading and trailing `...` |
| 20 | `list/grep/snippet-short` | Short matching field → full snippet text, no `...` ellipsis |
| 21 | `list/grep/snippet-match-only` | Match length ≥1024 runes → snippet is first 1024 runes of the match |
| 22 | `info/known-session` | Info returns summary fields, file paths, and token usage from signals.json |
| 23 | `info/unknown-session` | Unknown full UUID → `grok session not found` error |
| 24 | `info/no-signals` | Session without signals.json → info succeeds, no Tokens section |
| 25 | `info/untitled-session` | Empty title → shows `(untitled)`, num_chat_messages=1 |

## How to Run

```sh
doctest vet ./cmd/agent-pro/tests/grok-sessions
doctest test -v ./cmd/agent-pro/tests/grok-sessions
# grep branch only:
doctest test -v ./cmd/agent-pro/tests/grok-sessions/list/grep
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
	Grep      string    // empty → no content filter (classic list)
	Color     string    // "never" | "always" | "auto"; empty → "never" in harness
}

type Response struct {
	Sessions []sessions.Session
	Matches  []sessions.SessionMatch // filled when Grep != ""
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

		if req.Grep != "" {
			color := req.Color
			if color == "" {
				color = "never"
			}
			matches, err := sessions.ListWithGrep(req.GrokHome, limit, req.Grep)
			if err != nil {
				resp.Err = err
				return resp, nil
			}
			resp.Matches = matches
			for _, m := range matches {
				resp.Sessions = append(resp.Sessions, m.Session)
			}
			resp.Output = sessions.FormatListTableWithHits(matches, req.GrokHome, now, color)
		} else {
			list, err := sessions.List(req.GrokHome, limit)
			if err != nil {
				resp.Err = err
				return resp, nil
			}
			resp.Sessions = list
			resp.Output = sessions.FormatListTable(list, req.GrokHome, now)
		}
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
