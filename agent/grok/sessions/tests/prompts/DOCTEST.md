# sessions user-prompt history (Prompts / ListPrompts)

Doc-style tests for **user-prompt history** of Grok CLI sessions in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Covers package L2 APIs
`ParseRecentWindow`, `Prompts`, `ListPrompts`, `FilterUserPrompts`,
`FormatPromptsText`, and `FormatPromptsListText` for CLI:

```text
agent-pro grok session prompts <session-id>
agent-pro grok session prompts [--recent <window>] [--limit N]
  [--grep P] [--exclude Q] [--head N | --tail N] [--max-body N]
  [--color|--no-color]
```

**Classic TDD** — base list/format/parse/filter is implemented (most leaves
GREEN). **Body length contract** is the active RED surface: product still
soft-truncates every body at ~200 runes; new contract is **full body by
default** plus optional `FormatPromptsOptions.MaxBody*` / CLI `--max-body N`
(runes). This tree is **L2 in-process only** (library API); no product-binary e2e.

# DSN (Domain Specific Notion)

Read-only extraction of **user prompts** from Grok session storage, with
optional multi-session time window and limit selection, optional text filters
and per-session head/tail slice, plus compact text formatting for CLI stdout.

**Participants**

- **Caller** — CLI `session prompts` or in-process client that needs user
  prompt lines without full `session view` transcript chrome.
- **`Prompts`** — `Prompts(grokHome, sessionID string) (*SessionPrompts, error)`.
  Find session; convert `updates.jsonl` via `grok_session`; collect user
  messages chronologically (full history; **no** text filter). Unknown id →
  error containing `grok session not found`.
- **`ListPrompts`** — `ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error)`.
  Newest-first by `last_active_at`. Selection + filter pipeline:
  1. recent window (when `RecentSet`)
  2. grep keep (when `GrepSet`) — case-insensitive **literal** on `UserPrompt.Text`
  3. exclude drop (when `ExcludeSet`) — same matcher
  4. head **or** tail per-session slice (mutually exclusive)
  Session selection matrix (unchanged):
  - `!RecentSet && !LimitSet` → default **10** sessions
  - `!RecentSet && LimitSet` → limit **N**
  - `RecentSet && !LimitSet` → all sessions with ≥1 surviving prompt (no default cap)
  - `RecentSet && LimitSet` → surviving sessions only, stop at **N**
  Sessions with zero prompts **after** the full pipeline are **skipped** and
  do not count toward `--limit`.
- **`FilterUserPrompts`** — pure helper:
  `FilterUserPrompts(prompts []UserPrompt, opts FilterUserPromptsOptions) (kept []UserPrompt, omittedBefore, omittedAfter int, err error)`.
  Applies grep → exclude → head|tail on an in-memory slice (no FS). Used by
  single-session CLI path after `Prompts`, and by `ListPrompts` per session.
  Invalid opts (head+tail, empty pattern when set, N&lt;1) → clear error.
- **`ParseRecentWindow`** — `ParseRecentWindow(s string) (time.Duration, error)`.
  `^([0-9]+)([dhm])$` case-insensitive; `1d` = 24h rolling.
- **`UserPrompt`** — `Index` (1-based chrono), `Timestamp` from wire, `Text`.
- **`SessionPrompts`** — embeds `Session` plus `UserPrompts` and, after
  head/tail clip, `OmittedBefore` / `OmittedAfter` (counts of dropped prompts;
  zero when no clip). Marker is formatter chrome only — not a `UserPrompt`.
- **Formatters** — compact lines only:
  - `[YYYY-MM-DD HH:MM:SS] prompt text` (location from opts; tests use UTC)
  - Missing timestamp → `[—]` still print text
  - **Always** collapse internal whitespace (tabs/newlines → single spaces;
    one line per prompt body)
  - **Default body** (`!MaxBodySet`): **full** collapsed text — no soft-cap,
    no body-length `…`
  - **`MaxBodySet`**: soft-truncate body to **`MaxBodyRunes` runes** (N ≥ 1)
    + Unicode `…` (U+2026) **outside** the N content runes (same form as
    `softTruncateRunes`); invalid N &lt; 1 when set → clear error from Write*
  - With `GrepSet` and **no** MaxBody: full collapsed body; when
    `ColorMode=always`, bold-red highlight on the **first** include-match
  - With `GrepSet` **and** MaxBody: window body around first match within
    **N** runes (reuse window helpers; N replaces hard-coded 200); ellipsis
    on cut sides; color highlight when always
  - Head clip: print kept lines then `(...M omitted...)` if `OmittedAfter=M>0`
  - Tail clip: print `(...M omitted...)` if `OmittedBefore=M>0` then kept lines
  - Omission marker is dim meta when color on (clips *which* prompts, not body)
  - Multi: session header `── <id>  ·  <relative>  ·  <title>  ·  <short cwd>`
  - Footer `N sessions, M user messages` counts **printed** prompts only
    (excludes virtual omission lines)
  - Empty: `No user prompts found` + trailing newline
  - No `👤`, no multi-line USER cards

**Behaviors**

```
# single (full history)
grokHome + sessionID -> Prompts -> SessionPrompts

# single + filters (CLI path)
Prompts -> FilterUserPrompts(opts) -> FormatPromptsText
  invalid opts -> error

# multi
ListPrompts(opts: recent/limit + grep/exclude/head|tail)
  -> per session: window -> grep -> exclude -> head|tail
  -> skip empty survivors; limit counts survivors only
  -> []SessionPrompts with Omitted*

# format
SessionPrompts | []SessionPrompts + FormatPromptsOptions{Grep*, MaxBody*, ColorMode, …}
  -> compact text; full body default; optional MaxBody soft-cap; omission markers + footer; trailing \n
```

**Data source**

```text
{grokHome}/sessions/{url.PathEscape(abs_cwd)}/{uuid}/
  summary.json     # id, cwd, title, last_active_at
  updates.jsonl    # primary: user_message_chunk + wire timestamps
```

Do **not** use `chat_history.jsonl` as primary.

**Locked types (intended API; additive fields zero-value = old behavior)**

```text
UserPrompt
  Index     int
  Timestamp time.Time
  Text      string

SessionPrompts
  Session
  UserPrompts    []UserPrompt
  OmittedBefore  int   // tail clip; 0 if none
  OmittedAfter   int   // head clip; 0 if none

ListPromptsOptions
  Now, Recent, RecentSet, Limit, LimitSet, Home   // existing
  Grep, GrepSet
  Exclude, ExcludeSet
  Head, HeadSet      // int; N >= 1 when set
  Tail, TailSet

FilterUserPromptsOptions
  Grep, GrepSet
  Exclude, ExcludeSet
  Head, HeadSet
  Tail, TailSet

FormatPromptsOptions
  Now, Home, Location, Window, Limit, RecentSet, LimitSet, ColorMode  // existing
  Grep, GrepSet          // highlight (always → bold red); window only when MaxBodySet
  MaxBodyRunes int       // soft-cap body to N runes + "…" when MaxBodySet
  MaxBodySet   bool      // true if --max-body provided

ParseRecentWindow(s string) (time.Duration, error)
Prompts(grokHome, sessionID string) (*SessionPrompts, error)
ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error)
FilterUserPrompts(prompts []UserPrompt, opts FilterUserPromptsOptions) (kept []UserPrompt, omittedBefore, omittedAfter int, err error)
FormatPromptsText(sp *SessionPrompts, opts FormatPromptsOptions) string
FormatPromptsListText(list []SessionPrompts, opts FormatPromptsOptions) string
WritePromptsText / WritePromptsList — stream form; may return error on invalid MaxBody
```

Matcher: case-insensitive **literal** (same family as package `findLiteralCI` /
sessions list `--grep`). Grep/exclude target **`UserPrompt.Text` only**.

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/prompts/
├── DOCTEST.md
├── SETUP.md
├── parse-window/                 # ParseRecentWindow
├── single/                       # Prompts(sessionID); whitespace+full-body format
├── multi/                        # ListPrompts selection matrix
├── format/                       # Format* text contract + body length policy
│   ├── … existing compact/empty/header leaves …
│   ├── multi-long-bodies-full/   # multi list: long bodies full by default
│   └── max-body/                 # MaxBodySet soft-cap + invalid N
│       ├── cap-200/
│       ├── cap-1/
│       └── invalid-zero/
└── filter/                       # grep / exclude / head|tail / format chrome
    ├── grep/
    ├── exclude/
    ├── head-tail/
    │   ├── … existing short-text head/tail …
    │   └── head-long-body-full/  # head keeps full body (no body soft-cap)
    ├── errors/
    └── format-chrome/
        ├── omission-marker-exact/
        ├── color-grep-highlight/
        ├── grep-long-full-body/  # grep + long body, no MaxBody → full + highlight
        └── grep-max-body-window/ # grep + MaxBody → window ≤ N around match
```

Parameter ranking (most → least significant):

1. **API surface** — parse-window / single / multi / format / filter
2. **Body length policy** (under format / format-chrome) — default full vs MaxBody
3. **Filter kind** — grep | exclude | head-tail | errors | format-chrome
4. **Multi selection interaction** — limit / recent with text filters
5. **Outcome** — keep / skip session / empty / error / marker placement

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
| 12 | `single/whitespace-and-truncate` | Long/whitespace text: collapse still; **full** body by default (no body-cap `…`) |
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
| 30 | `format/multi-long-bodies-full` | Multi list: each long prompt body printed full (no soft-cap) |
| 31 | `format/max-body/cap-200` | MaxBodySet MaxBodyRunes=200 on 220-x body → ~200 content runes + `…` |
| 32 | `format/max-body/cap-1` | MaxBody N=1 on long body → 1 content rune + `…` |
| 33 | `format/max-body/invalid-zero` | MaxBodySet MaxBodyRunes=0 → clear error (max-body / ≥ 1) |
| 34 | `filter/grep/single-keep-matches` | Single session: grep keeps only matching prompts (structured) |
| 35 | `filter/grep/multi-keep-matches` | Multi: each session keeps only matching prompt texts |
| 36 | `filter/grep/case-insensitive` | Grep `error` matches `ERROR` / `Error` |
| 37 | `filter/grep/session-no-match-skipped` | Session with prompts but no grep hit is skipped (does not consume limit) |
| 38 | `filter/grep/no-matches-empty-message` | All filtered out → format contains `No user prompts found` |
| 39 | `filter/grep/with-limit-survivors` | Limit counts only sessions with ≥1 post-grep match |
| 40 | `filter/grep/with-recent-then-text` | Recent window first, then grep on remaining |
| 41 | `filter/exclude/drops-matching` | Exclude drops matching lines; non-matches remain |
| 42 | `filter/exclude/after-grep` | Grep keep then exclude drop (AND-not) |
| 43 | `filter/head-tail/head-clips-marker-after` | 5 prompts, head 2 → first 2 + trailing `(...3 omitted...)` |
| 44 | `filter/head-tail/tail-clips-marker-before` | 5 prompts, tail 2 → leading `(...3 omitted...)` + last 2 |
| 45 | `filter/head-tail/head-covers-all-no-marker` | head N ≥ total → all lines, no omission marker |
| 46 | `filter/head-tail/multi-per-session-head` | Per-session head 1; each block capped + own M |
| 47 | `filter/head-tail/footer-real-prompt-count` | Footer user-message count excludes virtual omission lines |
| 48 | `filter/head-tail/head-long-body-full` | Head keeps long body **full** by default; omission marker still after |
| 49 | `filter/errors/head-and-tail-both` | HeadSet+TailSet → error |
| 50 | `filter/errors/head-zero` | HeadSet Head=0 → error |
| 51 | `filter/errors/tail-zero` | TailSet Tail=0 → error |
| 52 | `filter/errors/empty-grep` | GrepSet with empty Grep → error |
| 53 | `filter/errors/empty-exclude` | ExcludeSet with empty Exclude → error |
| 54 | `filter/format-chrome/omission-marker-exact` | Output contains exact `(...M omitted...)` substring |
| 55 | `filter/format-chrome/color-grep-highlight` | ColorMode always + grep short body → ANSI CSI on match |
| 56 | `filter/format-chrome/grep-long-full-body` | Grep + long body, no MaxBody → full body + bold-red first match |
| 57 | `filter/format-chrome/grep-max-body-window` | Grep + MaxBody N → window ≤ N around first match; side ellipsis |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/prompts
doctest test ./agent/grok/sessions/tests/prompts
doctest test -v ./agent/grok/sessions/tests/prompts/format/max-body/cap-200
```

Classic TDD: short-body / filter / selection leaves stay GREEN. Leaves that
demand **full body by default** or **MaxBody** behavior stay RED until
implementer changes formatters + additive `MaxBodyRunes`/`MaxBodySet` fields
(and Write* validation for invalid N).

```go
import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

// applyFormatMaxBody sets MaxBodyRunes / MaxBodySet on opts when those fields
// exist on FormatPromptsOptions. Classic-TDD: suite compiles before implementer
// lands the additive fields; after they land, opts carry MaxBody for format.
func applyFormatMaxBody(opts *sessions.FormatPromptsOptions, maxRunes int, maxSet bool) {
	if opts == nil {
		return
	}
	v := reflect.ValueOf(opts).Elem()
	if f := v.FieldByName("MaxBodyRunes"); f.IsValid() && f.CanSet() && f.Kind() == reflect.Int {
		f.SetInt(int64(maxRunes))
	}
	if f := v.FieldByName("MaxBodySet"); f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
		f.SetBool(maxSet)
	}
}

// Op selects which package entrypoint Run exercises.
//   "parse"         — ParseRecentWindow(RecentRaw)
//   "single"        — Prompts; when any filter set → FilterUserPrompts on result
//   "list"          — ListPrompts (applies selection + filter pipeline)
//   "filter"        — FilterUserPrompts(req.FilterInput, filter opts) pure
//   "format-single" — Prompts (+ FilterUserPrompts if filters) then FormatPromptsText
//   "format-list"   — ListPrompts then FormatPromptsListText
//   "format-empty"  — FormatPromptsListText(nil/empty) without FS list
//   "format-synthetic" — FormatPromptsText on req.Synthetic (no FS; no re-filter)

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

	// Text filters + per-session slice (zero-value = no filter / no slice)
	Grep       string
	GrepSet    bool
	Exclude    string
	ExcludeSet bool
	Head       int
	HeadSet    bool
	Tail       int
	TailSet    bool

	// Format
	ColorMode string // "auto"|"always"|"never"; empty → never in Format*

	// Body soft-cap (zero-value = full body; MaxBodySet → N runes + "…")
	MaxBodyRunes int
	MaxBodySet   bool

	// Pure FilterUserPrompts input (Op "filter")
	FilterInput []sessions.UserPrompt

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

	// Pure filter
	Filtered      []sessions.UserPrompt
	OmittedBefore int
	OmittedAfter  int

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
		ColorMode: req.ColorMode,
		Grep:      req.Grep,
		GrepSet:   req.GrepSet,
	}
	// Demand MaxBody* on FormatPromptsOptions (implementer adds fields).
	// Reflection keeps the suite compiling until those fields exist; once
	// present they are set so MaxBody leaves exercise the real API.
	applyFormatMaxBody(&fmtOpts, req.MaxBodyRunes, req.MaxBodySet)
	listOpts := sessions.ListPromptsOptions{
		Now:        req.Now,
		Recent:     req.Recent,
		RecentSet:  req.RecentSet,
		Limit:      req.Limit,
		LimitSet:   req.LimitSet,
		Home:       req.Home,
		Grep:       req.Grep,
		GrepSet:    req.GrepSet,
		Exclude:    req.Exclude,
		ExcludeSet: req.ExcludeSet,
		Head:       req.Head,
		HeadSet:    req.HeadSet,
		Tail:       req.Tail,
		TailSet:    req.TailSet,
	}
	filterOpts := sessions.FilterUserPromptsOptions{
		Grep:       req.Grep,
		GrepSet:    req.GrepSet,
		Exclude:    req.Exclude,
		ExcludeSet: req.ExcludeSet,
		Head:       req.Head,
		HeadSet:    req.HeadSet,
		Tail:       req.Tail,
		TailSet:    req.TailSet,
	}
	hasFilter := req.GrepSet || req.ExcludeSet || req.HeadSet || req.TailSet

	applyFilterToSession := func(sp *sessions.SessionPrompts) error {
		if sp == nil || !hasFilter {
			return nil
		}
		kept, ob, oa, err := sessions.FilterUserPrompts(sp.UserPrompts, filterOpts)
		if err != nil {
			return err
		}
		sp.UserPrompts = kept
		sp.OmittedBefore = ob
		sp.OmittedAfter = oa
		return nil
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
		if err == nil {
			if ferr := applyFilterToSession(sp); ferr != nil {
				resp.Err = ferr
			}
		}
	case "list":
		list, err := sessions.ListPrompts(req.GrokHome, listOpts)
		resp.List = list
		resp.Err = err
	case "filter":
		kept, ob, oa, err := sessions.FilterUserPrompts(req.FilterInput, filterOpts)
		resp.Filtered = kept
		resp.OmittedBefore = ob
		resp.OmittedAfter = oa
		resp.Err = err
	case "format-single":
		sp, err := sessions.Prompts(req.GrokHome, req.SessionID)
		resp.Single = sp
		resp.Err = err
		if err == nil {
			if ferr := applyFilterToSession(sp); ferr != nil {
				resp.Err = ferr
			} else {
				// Prefer Write* so invalid MaxBody can surface as resp.Err.
				var b strings.Builder
				if werr := sessions.WritePromptsText(&b, sp, fmtOpts); werr != nil {
					resp.Err = werr
				} else {
					resp.Output = b.String()
				}
			}
		}
	case "format-list":
		list, err := sessions.ListPrompts(req.GrokHome, listOpts)
		resp.List = list
		resp.Err = err
		if err == nil {
			var b strings.Builder
			if werr := sessions.WritePromptsList(&b, list, fmtOpts); werr != nil {
				resp.Err = werr
			} else {
				resp.Output = b.String()
			}
		}
	case "format-empty":
		var b strings.Builder
		if werr := sessions.WritePromptsList(&b, nil, fmtOpts); werr != nil {
			resp.Err = werr
		} else {
			resp.Output = b.String()
		}
	case "format-synthetic":
		resp.Single = req.Synthetic
		var b strings.Builder
		if werr := sessions.WritePromptsText(&b, req.Synthetic, fmtOpts); werr != nil {
			resp.Err = werr
		} else {
			resp.Output = b.String()
		}
	default:
		t.Fatalf("unknown Op %q", req.Op)
	}
	return resp, nil
}
```
