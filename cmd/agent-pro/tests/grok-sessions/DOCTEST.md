# Grok Sessions Tests

Doc-style tests for `agent/grok/sessions`, which lists Grok CLI sessions from
`summary.json` files under a synthetic `GROK_HOME`, formats them as a table
with relative last-active times and message counts, optionally filters by
literal case-insensitive grep over session JSON content (with indented hit
lines and optional ANSI color), shows detailed session info from
`summary.json`, `signals.json`, and the session directory layout, and analyses
per-session stats (counts, latency, tool handler time, background tasks,
subagents) from `signals.json`, `events.jsonl`, and `updates.jsonl`.

# DSN (Domain Specific Notion)

Grok CLI stores session metadata under `{grokHome}/sessions/{encoded_cwd}/{uuid}/`.
Each session directory contains `summary.json` with session id, cwd, title
(`generated_title` / `session_summary`), message counts, model/agent metadata,
and activity timestamps. Optional files include `chat_history.jsonl` (message
log by `type`: system/user/assistant/reasoning/tool_result), `signals.json`
(token usage plus session-level turn/tool/latency counters), `events.jsonl`
(tool lifecycle lines such as `tool_started` / `tool_completed` with
`duration_ms` and `outcome`), `updates.jsonl` (wire log of session updates:
thought chunks, task_backgrounded / task_completed, subagent_spawned /
subagent_finished), and `prompt_context.json`.
The encoded cwd directory name is `url.PathEscape(abs_cwd)`.

The sessions package walks `{grokHome}/sessions/*/<uuid>/summary.json`, parses
each file, skips malformed entries, sorts by `last_active_at` descending
(tie-break by session id), truncates to the requested limit, and formats
results as a table. `FormatListTable` shows `num_chat_messages` in an `MSGS`
column and accepts a fixed `now` clock so relative times (`just now`,
`5m ago`, `2h ago`) are deterministic in tests.

When one or more grep patterns are set, only sessions with at least one
case-insensitive literal hit in `summary.json` (title-ish fields, cwd, model,
agent) or `chat_history.jsonl` (message text by type/part) are listed. Multiple
patterns are **AND on the same field/line** (every pattern must appear in that
unit). Limit applies **after** filtering. Under each matching session row, the
formatter prints up to five indented hit lines
`  <file>:<line>:<part>: <snippet>`, then `  ... and N more matches` when
remaining hits exist. Each hit **snippet** is built from the matched field
after whitespace collapse (runs of spaces/tabs/newlines → a single ASCII space,
with leading/trailing trim): the window is at most **1024 runes** (Unicode code
points), centered ~50/50 around the **first pattern's** first literal match,
with ASCII `...` (3 runes) on each truncated side; when the match itself is
≥1024 runes the snippet is the first 1024 runes of the match only.
`MatchStart`/`MatchLen` are byte offsets into the final snippet for the first
pattern; color mode highlights **all** patterns in the snippet.
Color mode `never`/`always`/`auto` controls ANSI styling of hit lines (tests
force `never` or `always`).

For session detail, `Find` locates a session by exact UUID (no prefix matching),
`Info` aggregates summary fields, filesystem paths, and token usage from
`signals.json`, and `FormatInfoText` renders key-value blocks for the CLI
`agent-pro grok session info <id>` command.

For session stats, `Stats` locates a session by exact UUID, then aggregates:

1. **Identity** from `summary.json` (via `Find`): id, title, cwd, model, agent,
   created/last-active timestamps.
2. **Counts and session latency** preferentially from `signals.json`
   (`turnCount`, `userMessageCount`, `assistantMessageCount`, `toolCallCount`,
   `toolFailureCount`, `errorCount`, `cancellationCount`, `compactionCount`,
   `sessionDurationSeconds`, `avgResponseTimeMs`, `avgTimeToFirstTokenMs`).
3. **Per-tool handler time** from `events.jsonl` lines with
   `type=tool_completed`: group by `tool_name`, use `duration_ms` for
   min/max/avg/median, and `outcome` for success vs error counts.
   `ToolCompleted` is the number of `tool_completed` lines. When signals are
   missing tool counts, `ToolCalls` falls back toward `tool_started` counts.
4. **Thinking blocks** from `updates.jsonl`: count *coalesced runs* of
   consecutive `agent_thought_chunk` updates (flat `sessionUpdate` or nested
   under `params.update`). A non-thought update ends the current run.
5. **Background tasks** from `updates.jsonl`:
   - optional `task_backgrounded` maps `task_id` → description
   - `task_completed` wall-clock ms from `task_snapshot.start_time` /
     `end_time` (`secs_since_epoch` + nanos); full `command` stored (no
     store truncate); optional `exit_code`; join description by task_id
   - items: `[]BackgroundTaskItem` (DurationMs, Command, Description,
     ExitCode *int, Kind, CWD)
6. **Subagents** from `updates.jsonl`:
   - `subagent_spawned` maps `subagent_id` → description, type, model
   - `subagent_finished` uses `duration_ms`, status, tool_calls, turns,
     tokens_used; join spawn meta by `subagent_id`
   - items: `[]SubagentItem` (DurationMs, ID, Description, Type, Status,
     ToolCalls, Turns, TokensUsed, Model)

Missing optional files do **not** fail `Stats`: `Sources` flags mark which
files were used, and human-readable strings are appended to
`Sources.Warnings`. Primary counts do **not** come from `chat_history.jsonl`.

`FormatStatsText` / `FormatStatsTextOpts` render stable section headers
(Counts, Latency, Tool handler time, Background tasks, Subagents, Sources)
and omit empty optional sections. Human text upgrades:

1. **Pretty durations** on all duration fields: `<1s` → integer `ms`
   (e.g. `400ms`); `≥1s` non-whole → one decimal `s` (e.g. `1.5s`); whole
   seconds and larger use compact units (`3s`, `2m`, `2h39m12s`).
2. **Tool table** when tools exist: header columns `NAME N SUCCESS ERROR AVG
   MED MIN MAX`; rows sorted by **Count desc**, then Name asc; duration
   columns use pretty forms.
3. **Top-N sections** (when `TopN > 0` and data exists):
   - Top tools by total handler time (`Count*AvgMs`)
   - **Top background tasks** table
     `#  DURATION  EXIT  COMMAND` — EXIT is integer or `-` if nil; COMMAND
     display-truncated at **200 runes** then Unicode `…` (store keeps full
     command)
   - **Top subagents** table
     `#  DURATION  STATUS  TYPE  TOOLS  TURNS  DESC` — DESC from spawn
     description (truncate **64 runes**), fallback to short id when empty;
     not UUID-only when description is available
   - `TopN == 0` hides all Top section headers. Default TopN is **5**.
4. **Color** via `ColorMode` `never` | `always` | `auto` (same policy as
   grep `shouldColor` / `NO_COLOR`): dim labels; red when tool ERROR /
   failures / errors > 0; green for source `yes`. Tests force `never` or
   `always` for determinism.

CLI shape (out of package-test scope):
`agent-pro grok session stats <session-id> [--json] [--by-tool] [--top N] [--color|--no-color]`.
JSON remains raw ms, no ANSI, no pretty strings.

The test harness builds a temporary Grok home, writes minimal fixtures under
encoded cwd paths (including optional `chat_history.jsonl`, richer
`signals.json`, `events.jsonl`, `updates.jsonl`), and calls the package API
directly (no real `~/.grok`).

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
│       ├── multi-and/           multiple --grep → AND on same field/line
│       │   ├── same-unit-keep/  both tokens on one line → keep; split other omitted
│       │   └── split-units-drop/ tokens on different lines only → no sessions
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
├── info/
│   ├── known-session/        full summary + file paths + tokens from signals.json
│   ├── unknown-session/      missing UUID → grok session not found
│   ├── no-signals/           no signals.json → info succeeds, no Tokens section
│   └── untitled-session/     empty title → (untitled), num_chat_messages=1
└── stats/
    ├── known-session/           summary + signals + events + updates → full SessionStats
    ├── unknown-session/         missing UUID → grok session not found
    ├── signals-only/            counts/latency from signals; warnings for missing events/updates
    ├── events-tool-avg/         tool_completed duration_ms → avg/med/min/max + success/error
    ├── thinking-blocks/         consecutive agent_thought_chunk runs coalesce; gap starts new
    ├── background-task/         task_completed wall clock → Count / AvgMs / MaxMs
    ├── subagent-duration/       subagent_finished duration_ms → Count / AvgMs / MaxMs
    ├── format-text/                  stable section headers + key counts (pretty durations allowed)
    ├── format-pretty-duration/       Latency: 120s→2m, 1500ms→1.5s, 400ms→400ms
    ├── format-tool-table-sort/       tool rows by N desc; header NAME/N present
    ├── format-top-n/                 TopN=2 caps top-tool lines; rich bg/sub headers; TopN=0 hides Top
    ├── format-top-bg-long-command/   COMMAND 120 runes full; 220 runes → … at 200 display cap
    ├── format-top-bg-exit/           EXIT column shows 1 for exit_code 1
    ├── format-top-subagent-rich/     spawn desc + type + tools/turns in Top subagents
    ├── format-top-subagent-join/     finish without spawn → id fallback, no crash
    ├── format-color-never/           ColorMode never → no ANSI escapes
    └── format-color-always/          ColorMode always + failures/sources → ANSI present
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
| 26 | `stats/known-session` | Full summary + signals + events + updates → identity, counts, tools, tasks, sources |
| 27 | `stats/unknown-session` | Unknown full UUID → `grok session not found` error |
| 28 | `stats/signals-only` | Counts/latency from signals; Sources.Warnings mention missing events/updates |
| 29 | `stats/events-tool-avg` | Known tool_completed durations → correct avg/med/min/max and success/error per tool |
| 30 | `stats/thinking-blocks` | Consecutive thought chunks → 1 block; non-thought gap → second block |
| 31 | `stats/background-task` | task_completed start/end wall clock → BackgroundTasks Count/AvgMs/MaxMs |
| 32 | `stats/subagent-duration` | subagent_finished duration_ms → Subagents Count/AvgMs/MaxMs |
| 33 | `stats/format-text` | FormatStatsText contains Counts/Latency/Tool handler time/Sources headers |
| 34 | `stats/format-pretty-duration` | Latency lines use pretty forms (`2m`, `1.5s`, `400ms`) from default signals fixture |
| 35 | `stats/format-tool-table-sort` | Tool table has NAME/N header; higher-N tools appear before lower-N |
| 36 | `stats/format-top-n` | TopN=2 → ≤2 top-tool rows; rich EXIT/COMMAND + subagent headers; TopN=0 hides Top |
| 37 | `stats/format-top-bg-long-command` | 120-rune COMMAND full (no 60-cap …); 220-rune COMMAND truncated at 200+… |
| 38 | `stats/format-top-bg-exit` | EXIT column shows `1` for task_snapshot.exit_code 1 |
| 39 | `stats/format-top-subagent-rich` | spawn→finish join: DESC/type/tools/turns in Top subagents (not UUID-only) |
| 40 | `stats/format-top-subagent-join` | finish without spawn → id fallback in DESC; Stats succeeds |
| 41 | `stats/format-color-never` | ColorMode never → output has no ANSI CSI escapes |
| 42 | `stats/format-color-always` | ColorMode always with tool errors/sources → ANSI present (dim/green/red) |

## How to Run

```sh
doctest vet ./cmd/agent-pro/tests/grok-sessions
doctest test -v ./cmd/agent-pro/tests/grok-sessions
# stats branch only:
doctest test -v ./cmd/agent-pro/tests/grok-sessions/stats
```

```go
import (

	"fmt"
	"testing"
	"time"

	sessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Operation string    // "list", "info", or "stats"; default "list"
	GrokHome  string
	SessionID string    // info/stats only; exact UUID
	Limit     int       // 0 → default 20 for list
	Now       time.Time // fixed clock for relative times in formatters
	Grep      []string  // empty → no content filter (classic list); AND on same unit
	Color     string    // list grep: "never" | "always" | "auto"; empty → "never" in harness
	// Stats formatter options (FormatStatsTextOpts):
	ColorMode string // "never" | "always" | "auto"; empty → "never" (deterministic tests)
	TopN      int    // top-N section size; meaningful when TopNSet
	TopNSet   bool   // false → Run defaults TopN to 5; true → use TopN (0 hides Top sections)
}

type Response struct {
	Sessions []sessions.Session
	Matches  []sessions.SessionMatch // filled when len(Grep) > 0
	Info     *sessions.SessionInfo
	Stats    *sessions.SessionStats
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

		if len(req.Grep) > 0 {
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
	case "stats":
		stats, err := sessions.Stats(req.GrokHome, req.SessionID)
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		resp.Stats = stats
		colorMode := req.ColorMode
		if colorMode == "" {
			colorMode = "never"
		}
		topN := 5
		if req.TopNSet {
			topN = req.TopN
		}
		resp.Output = sessions.FormatStatsTextOpts(stats, sessions.FormatStatsOptions{
			Home:      req.GrokHome,
			Now:       now,
			ColorMode: colorMode,
			TopN:      topN,
		})
	default:
		resp.Err = fmt.Errorf("unknown operation: %s", op)
	}
	return resp, nil
}
```
