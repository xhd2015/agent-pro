# sessions user-prompt history (Prompts / ListPrompts)

Doc-style tests for **user-prompt history** of Grok CLI sessions in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Covers package L2 APIs
`ParseRecentWindow`, `Prompts`, `ListPrompts`, `FormatPromptsText`, and
`FormatPromptsListText` for CLI:

```text
agent-pro grok session prompts <session-id>
agent-pro grok session prompts [--recent <window>] [--limit N]
```

**Classic TDD** — greenfield. Leaves stay RED until the implementer lands the
API (and wire-timestamp plumbing in `agent/event/grok_session` as needed).
This tree is **L2 in-process only** (library API); no product-binary e2e.

# DSN (Domain Specific Notion)

Read-only extraction of **user prompts** from Grok session storage, with
optional multi-session time window and limit selection, plus compact text
formatting for CLI stdout.

**Participants**

- **Caller** — CLI `session prompts` or in-process client that needs user
  prompt lines without full `session view` transcript chrome.
- **`Prompts`** — `Prompts(grokHome, sessionID string) (*SessionPrompts, error)`.
  Find session; convert `updates.jsonl` via `grok_session`; collect user
  messages chronologically. Unknown id → error containing
  `grok session not found` (same style as `sessionNotFoundError`).
- **`ListPrompts`** — `ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error)`.
  Newest-first by `last_active_at`. Selection matrix:
  - `!RecentSet && !LimitSet` → default **10** sessions
  - `!RecentSet && LimitSet` → limit **N**
  - `RecentSet && !LimitSet` → all sessions with ≥1 in-window user prompt (no default cap)
  - `RecentSet && LimitSet` → in-window sessions only, stop at **N**
  When `RecentSet`, a prompt is included only if
  `Timestamp ∈ [Now-Recent, Now]` (inclusive ends). Sessions with zero
  in-window prompts are **skipped** and do not count toward limit.
  When `!RecentSet`, each selected session contributes **all** its user
  prompts (full history for those sessions).
- **`ParseRecentWindow`** — `ParseRecentWindow(s string) (time.Duration, error)`.
  `^([0-9]+)([dhm])$` case-insensitive; `1d` = 24h rolling; reject `0`,
  empty, bare numbers, `2w`, etc. Error text mentions `Nd`, `Nh`, or `Nm`.
- **`UserPrompt`** — `Index` (1-based chrono), `Timestamp` from wire (zero if
  unknown), `Text` (raw coalesced user text before format collapse).
- **`SessionPrompts`** — embeds existing `Session` (ID, LastActiveAt, CWD,
  Title, …) plus `UserPrompts []UserPrompt`.
- **Formatters** — compact lines only:
  - `[YYYY-MM-DD HH:MM:SS] prompt text…` (location from opts; tests use UTC)
  - Missing timestamp → `[—]` still print text
  - Collapse internal whitespace; soft-truncate body ~200 runes + `…`
  - Multi: session header `── <id>  ·  <relative>  ·  <title>  ·  <short cwd>`
  - Empty multi/single format: friendly `No user prompts found` (window-specific
    wording allowed if it still matches that core phrase)
  - Trailing newline on all formatted stdout strings (CLI contract)
  - No `👤`, no multi-line USER cards

**Behaviors**

```
# single
grokHome + sessionID
  -> Find session
  -> missing -> error "grok session not found: <id>"
  -> read updates.jsonl -> convert user messages (wire timestamps)
  -> SessionPrompts{Session, UserPrompts}

# multi
grokHome + ListPromptsOptions{Now, Recent, RecentSet, Limit, LimitSet}
  -> discover sessions newest-first by last_active
  -> filter by window / apply limit matrix
  -> per session: user prompts (windowed when RecentSet)
  -> []SessionPrompts

# window parse
"30m"|"2h"|"1d" -> duration ; invalid -> error

# format
SessionPrompts | []SessionPrompts + FormatPromptsOptions
  -> compact text (+ optional footer); trailing \n
```

**Data source**

```text
{grokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/
  summary.json     # id, cwd, title, last_active_at
  updates.jsonl    # primary: user_message_chunk + wire timestamps
```

Do **not** use `chat_history.jsonl` as primary. Wire timestamps plumbed into
`AgentEvent.Timestamp` (ms); coalesced user message uses **first** chunk time.
Do **not** use convert-time `time.Now()` for prompt times.

**Locked types (intended API)**

```text
UserPrompt
  Index     int
  Timestamp time.Time
  Text      string

SessionPrompts
  Session              // existing sessions.Session
  UserPrompts []UserPrompt

ListPromptsOptions
  Now       time.Time
  Recent    time.Duration
  RecentSet bool
  Limit     int
  LimitSet  bool
  Home      string     // optional path shorten for formatters

FormatPromptsOptions
  Now       time.Time
  Home      string
  Location  *time.Location  // nil → time.Local; tests pass time.UTC
  Window    time.Duration   // footer only
  Limit     int             // footer only
  RecentSet bool
  LimitSet  bool

ParseRecentWindow(s string) (time.Duration, error)
Prompts(grokHome, sessionID string) (*SessionPrompts, error)
ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error)
FormatPromptsText(sp *SessionPrompts, opts FormatPromptsOptions) string
FormatPromptsListText(list []SessionPrompts, opts FormatPromptsOptions) string
```

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/prompts/
├── DOCTEST.md
├── SETUP.md
├── parse-window/                 # ParseRecentWindow
│   ├── valid/                    # Nd|Nh|Nm accepted
│   │   ├── 30m/
│   │   ├── 2h/
│   │   ├── 1d/
│   │   └── case-insensitive-2H/
│   └── invalid/                  # reject with clear error
│       ├── empty/
│       ├── zero-0m/
│       ├── bare-number/
│       └── weeks-2w/
├── single/                       # Prompts(sessionID)
│   ├── known-two-prompts/
│   ├── multi-chunk-coalesce/
│   ├── assistant-only-empty/
│   ├── whitespace-and-truncate/  # structured + format collapse/truncate
│   ├── unknown-session/
│   ├── empty-session-id/
│   └── missing-updates-file/
├── multi/                        # ListPrompts selection matrix
│   ├── no-recent/                # !RecentSet
│   │   ├── default-limit-10/
│   │   ├── limit-3/
│   │   ├── empty-home/
│   │   └── full-history-all-prompts/
│   └── with-recent/              # RecentSet
│       ├── window-filter-no-default-cap/
│       ├── recent-plus-limit-clips/
│       ├── recent-plus-limit-one/
│       ├── last-active-in-prompts-out/
│       └── all-outside-empty/
└── format/                       # Format* text contract
    ├── single-compact-utc/
    ├── missing-timestamp-em-dash/
    ├── multi-session-header/
    ├── empty-friendly-message/
    └── no-emoji-user-headers/
```

Parameter ranking (most → least significant):

1. **API surface** — parse-window / single / multi / format
2. **Multi selection** — RecentSet × LimitSet matrix (`no-recent` vs `with-recent`)
3. **Outcome** — happy path vs empty vs not-found / invalid input
4. **Wire / text edges** — coalesce, truncate, missing timestamp, headers

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `parse-window/valid/30m` | `"30m"` → 30 minutes |
| 2 | `parse-window/valid/2h` | `"2h"` → 2 hours |
| 3 | `parse-window/valid/1d` | `"1d"` → 24 hours |
| 4 | `parse-window/valid/case-insensitive-2H` | `"2H"` accepted like `"2h"` |
| 5 | `parse-window/invalid/empty` | empty string → error mentions Nd/Nh/Nm |
| 6 | `parse-window/invalid/zero-0m` | `"0m"` rejected |
| 7 | `parse-window/invalid/bare-number` | `"30"` bare number rejected |
| 8 | `parse-window/invalid/weeks-2w` | `"2w"` rejected |
| 9 | `single/known-two-prompts` | Two separate user messages with wire ms → 2 prompts, times match wire (UTC) |
| 10 | `single/multi-chunk-coalesce` | Consecutive user chunks → one prompt; first-chunk timestamp |
| 11 | `single/assistant-only-empty` | Only assistant/tool updates → empty `UserPrompts`, no error |
| 12 | `single/whitespace-and-truncate` | Long/whitespace text collapses + truncates in format (~200 runes + `…`) |
| 13 | `single/unknown-session` | Missing id → `grok session not found` + id |
| 14 | `single/empty-session-id` | Empty / whitespace id → not found error |
| 15 | `single/missing-updates-file` | summary only, no `updates.jsonl` → empty prompts, no error |
| 16 | `multi/no-recent/default-limit-10` | 15 sessions, no flags → 10 newest session blocks |
| 17 | `multi/no-recent/limit-3` | `LimitSet` 3 → 3 newest |
| 18 | `multi/no-recent/empty-home` | Empty sessions tree → empty list; format friendly empty |
| 19 | `multi/no-recent/full-history-all-prompts` | No recent → selected session includes prompts outside any implicit window |
| 20 | `multi/with-recent/window-filter-no-default-cap` | Recent 1h; >10 sessions with in-window prompts → all returned (no default 10) |
| 21 | `multi/with-recent/recent-plus-limit-clips` | 7 in-window sessions, limit 2 → 2 blocks newest-first |
| 22 | `multi/with-recent/recent-plus-limit-one` | 1 in-window, limit 5 → 1 block |
| 23 | `multi/with-recent/last-active-in-prompts-out` | last_active in window but all prompt ts outside → session skipped |
| 24 | `multi/with-recent/all-outside-empty` | All prompts outside window → empty list |
| 25 | `format/single-compact-utc` | `[YYYY-MM-DD HH:MM:SS] text` UTC; trailing newline |
| 26 | `format/missing-timestamp-em-dash` | Zero timestamp → `[—]` prefix |
| 27 | `format/multi-session-header` | Multi header has id, title, cwd; prompt lines follow |
| 28 | `format/empty-friendly-message` | Empty list → contains `No user prompts found`; trailing newline |
| 29 | `format/no-emoji-user-headers` | No `👤` / multi-line USER card chrome |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/prompts
doctest test ./agent/grok/sessions/tests/prompts
doctest test -v ./agent/grok/sessions/tests/prompts/single/known-two-prompts
```

Classic TDD: all leaves RED until implementer adds `Prompts` / `ListPrompts` /
`ParseRecentWindow` / `Format*` (and wire timestamps on user events).

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

// Op selects which package entrypoint Run exercises.
//   "parse"         — ParseRecentWindow(RecentRaw)
//   "single"        — Prompts(GrokHome, SessionID)
//   "list"          — ListPrompts(GrokHome, list opts)
//   "format-single" — Prompts then FormatPromptsText
//   "format-list"   — ListPrompts then FormatPromptsListText
//   "format-empty"  — FormatPromptsListText(nil/empty) without FS list
//   "format-synthetic" — FormatPromptsText on req.Synthetic (no FS)

type Request struct {
	TempDir   string
	GrokHome  string
	SessionID string
	Op        string

	// ParseRecentWindow
	RecentRaw string

	// List / format clock and selection (injectable; root Setup sets fixed Now)
	Now       time.Time
	Recent    time.Duration
	RecentSet bool
	Limit     int
	LimitSet  bool
	Home      string
	Location  *time.Location

	// Synthetic structured data for format-only leaves that skip FS.
	Synthetic *sessions.SessionPrompts
	// When true, format-list uses empty slice without calling ListPrompts.
	FormatEmptyList bool
}

type Response struct {
	// Parse
	Window time.Duration

	// Single
	Single *sessions.SessionPrompts

	// Multi
	List []sessions.SessionPrompts

	// Format
	Output string

	Err error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	resp := &Response{}
	loc := req.Location
	if loc == nil {
		loc = time.UTC
	}
	fmtOpts := sessions.FormatPromptsOptions{
		Now:       req.Now,
		Home:      req.Home,
		Location:  loc,
		Window:    req.Recent,
		Limit:     req.Limit,
		RecentSet: req.RecentSet,
		LimitSet:  req.LimitSet,
	}
	listOpts := sessions.ListPromptsOptions{
		Now:       req.Now,
		Recent:    req.Recent,
		RecentSet: req.RecentSet,
		Limit:     req.Limit,
		LimitSet:  req.LimitSet,
		Home:      req.Home,
	}

	switch req.Op {
	case "parse":
		w, err := sessions.ParseRecentWindow(req.RecentRaw)
		resp.Window = w
		resp.Err = err
	case "single":
		sp, err := sessions.Prompts(req.GrokHome, req.SessionID)
		resp.Single = sp
		resp.Err = err
	case "list":
		list, err := sessions.ListPrompts(req.GrokHome, listOpts)
		resp.List = list
		resp.Err = err
	case "format-single":
		sp, err := sessions.Prompts(req.GrokHome, req.SessionID)
		resp.Single = sp
		resp.Err = err
		if err == nil {
			resp.Output = sessions.FormatPromptsText(sp, fmtOpts)
		}
	case "format-list":
		list, err := sessions.ListPrompts(req.GrokHome, listOpts)
		resp.List = list
		resp.Err = err
		if err == nil {
			resp.Output = sessions.FormatPromptsListText(list, fmtOpts)
		}
	case "format-empty":
		resp.Output = sessions.FormatPromptsListText(nil, fmtOpts)
	case "format-synthetic":
		resp.Single = req.Synthetic
		resp.Output = sessions.FormatPromptsText(req.Synthetic, fmtOpts)
	default:
		t.Fatalf("unknown Op %q", req.Op)
	}
	return resp, nil
}
```
