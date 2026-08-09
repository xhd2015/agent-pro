# sessions.ListWithOptions — place / recent / active / role / forked / grep filters + KIND

Doc-style tests for **list filter pipeline** on Grok CLI sessions in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Covers the injectable L2
API `ListOptions` / `ListWithOptions` that powers CLI:

```text
agent-pro grok sessions [--here] [--dir DIR]... [--recent WINDOW]
  [--active] [--main-agent] [--sub-agent] [--forked]
  [--grep PATTERN] [--limit N] [--color]
```

**Classic TDD** — place/recent/active/grep already GREEN (23 leaves). This tree
extends additively with **role**, **forked**, and **KIND** column leaves that
are **RED** until implementer lands:

- `ListOptions.MainAgent` / `SubAgent` / `Forked`
- `Session.Kind` display tokens populated from summary
- `FormatListTable` / `FormatListTableWithHits` KIND column after SESSION ID

**L2 in-process only** (library API); no product-binary e2e. CLI Abs existence
checks for `--dir` stay out of scope here.

# DSN (Domain Specific Notion)

Discover sessions under a synthetic `GROK_HOME`, then apply optional **place**,
**recent**, **active**, **role**, **forked**, and **grep** filters before sort
+ limit. Every listed session carries a **Kind** display token; list tables
always show a **KIND** column.

**Participants**

- **Caller** — CLI `grok sessions` or in-process client needing filtered
  session lists without `os.Chdir` / env mutation in tests.
- **`ListWithOptions`** —
  `ListWithOptions(grokHome string, opts ListOptions) ([]Session, error)`.
  Discovers all valid `summary.json` sessions (same as `List` / `discoverSessions`),
  then applies the pipeline below. Returns plain `[]Session` with `Kind` set.
  Grep is a **presence** filter only — hit lines remain `ListWithGrep` /
  `FormatListTableWithHits`.
- **`ListOptions`** — injectable filter + limit controls (zero value ≈
  `List(grokHome, 0)`: default limit 20, no place/recent/active/role/forked/grep).
- **`PlaceCWDs`** — caller-resolved absolute paths. Non-empty → place filter ON
  (OR match after Abs+Clean). Empty/nil → skip place.
- **`Recent` / `RecentSet`** — when set, keep sessions with
  `last_active_at >= Now - Recent` (inclusive). `Now` zero → `time.Now()`.
  `RecentSet && Recent <= 0` → error.
- **`Active`** — when true, keep only `IsFileActive` sessions.
- **`MainAgent` / `SubAgent`** — mutually exclusive role filters.
  - **Sub-agent class**: `session_kind ∈ {subagent, subagent_resume, subagent_fork}`
    OR (`session_kind` empty/absent AND `parent_session_id` non-empty).
  - **Main-agent class**: everything else (includes plain `fork` and missing
    kind with no parent).
  - `MainAgent && SubAgent` → error.
  - Neither → no role filter (all roles).
- **`Forked`** — when true, keep forked sessions:
  `session_kind ∈ {fork, subagent_fork}` OR `forked_at` is a non-empty
  non-whitespace string. **ANDs** with role when both set.
- **`Grep` / `GrepSet`** — content presence filter; empty Grep when set → error.
- **`Limit`** — max sessions **after** all filters. `<= 0` → 20; `> 100` → 100.
- **`Session.Kind`** — display token always populated on list results:

  | Condition | KIND token |
  |-----------|------------|
  | `subagent_fork` | `sub-f` |
  | `fork` | `fork` |
  | `subagent_resume` | `sub+` |
  | `subagent` | `sub` |
  | else (main, missing, …) | `main` |

  Priority: more specific kinds first (`subagent_fork` → `fork` → resume →
  subagent → default `main`). Classification for filters may also use
  `parent_session_id` / `forked_at`; display Kind follows the table above.
- **`FormatListTable` / `FormatListTableWithHits`** — header order:

  ```text
  SESSION ID  KIND  LAST ACTIVE  TITLE  MSGS  CWD
  ```

  Empty list still: `No sessions found` (no header).

**Behaviors**

```
# filter pipeline (locked order)
discover sessions
  -> place?     cwd ∈ PlaceCWDs (OR; skip if PlaceCWDs empty)
  -> recent?    last_active >= Now - Recent (when RecentSet)
  -> active?    IsFileActive (when Active)
  -> role?      main-agent class XOR sub-agent class (when MainAgent|SubAgent)
  -> forked?    fork kinds or forked_at (when Forked)
  -> grep?      content hits (when GrepSet)
  -> sort last_active desc (tie-break id desc, same as List)
  -> limit

# Kind always set on survivors (and on all discovered sessions returned)
# empty survivors
[]Session{}, nil  → FormatListTable => "No sessions found"
```

**Locked types (intended API)**

```text
type ListOptions struct {
  Limit     int
  PlaceCWDs []string
  Recent    time.Duration
  RecentSet bool
  Active    bool
  Now       time.Time
  Grep      string
  GrepSet   bool
  MainAgent bool // --main-agent
  SubAgent  bool // --sub-agent
  Forked    bool // --forked
  // MainAgent && SubAgent → error from ListWithOptions
}

type Session struct {
  // existing fields...
  Kind string // display: main|sub|sub+|sub-f|fork
}

ListWithOptions(grokHome string, opts ListOptions) ([]Session, error)
```

**CLI mapping (implementer; not tested in this tree)**

```text
--here           → PlaceCWDs append cwd
--dir DIR        → PlaceCWDs append Abs+Clean
--recent WINDOW  → RecentSet + Recent
--active         → Active=true
--main-agent     → MainAgent=true
--sub-agent      → SubAgent=true
--forked         → Forked=true
--grep PATTERN   → GrepSet + Grep
--limit N        → Limit
```

Place flags form **OR**; role/forked/recent/active/grep **AND**. Tests inject
opts only — **never** `os.Chdir`, `t.Chdir`, `os.Setenv`, or `t.Setenv`.

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/list-filter/
├── DOCTEST.md
├── SETUP.md
├── default/                      # no place/recent/active/role/forked/grep
│   ├── all-sessions/
│   └── empty-after-discover/
├── place/                        # PlaceCWDs OR match
│   ├── single-cwd/
│   ├── multi-or/
│   ├── no-match-empty/
│   ├── abs-clean-equality/
│   ├── dedup-overlapping/
│   └── empty-cwd-excluded/
├── recent/                       # last_active window
│   ├── within-window/
│   ├── all-outside-empty/
│   └── boundary-inclusive/
├── active/                       # file-active only
│   ├── only-file-active/
│   └── none-listed-empty/
├── role/                         # MainAgent / SubAgent class filter  [NEW]
│   ├── main/
│   │   └── keeps-main-class/     # main + fork + empty-no-parent; drops sub*
│   ├── sub/
│   │   └── keeps-sub-class/      # subagent family + empty+parent; drops main/fork
│   └── none/
│       └── all-with-kind/        # neither role flag → all; Kind populated
├── forked/                       # Forked filter  [NEW]
│   ├── matches/                  # fork, subagent_fork, forked_at kept
│   ├── excludes-plain/           # plain main/sub without forked_at dropped
│   └── whitespace-forked-at/     # whitespace-only forked_at not forked
├── combo/                        # AND across dimensions
│   ├── place-and-recent/
│   ├── place-and-active/
│   ├── recent-and-active/
│   ├── place-and-grep/
│   ├── recent-and-grep/
│   ├── place-recent-active/
│   ├── main-and-forked/          # [NEW] MainAgent ∩ Forked
│   ├── sub-and-forked/           # [NEW] SubAgent ∩ Forked
│   └── place-and-main/           # [NEW] place AND MainAgent
├── format/                       # KIND column rendering  [NEW]
│   ├── header-kind-column/       # SESSION ID then KIND then LAST ACTIVE
│   ├── kind-tokens/              # main|sub|sub+|sub-f|fork in rows
│   ├── empty-no-header/          # empty → "No sessions found"
│   └── hits-kind-column/         # FormatListTableWithHits also has KIND
├── limit/
│   ├── after-all-filters/
│   └── default-limit-20/
└── errors/
    ├── recent-nonpositive/
    ├── empty-grep-when-set/
    └── main-and-sub-exclusive/   # [NEW] MainAgent && SubAgent → error
```

Parameter ranking (most → least significant):

1. **Filter dimension** — default | place | recent | active | role | forked | combo | format | limit | errors
2. **Role class** — main | sub | neither (both → errors)
3. **Forked membership** — match / exclude / whitespace edge
4. **AND combos** — role×forked, place×role, place×recent×active×grep
5. **Format** — header order, tokens, empty phrase, hits parity
6. **Limit / validation**

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `default/all-sessions` | No filters; three sessions under distinct cwds; high Limit → all three newest-first |
| 2 | `default/empty-after-discover` | Empty sessions tree; ListWithOptions → empty; FormatListTable `"No sessions found"` |
| 3 | `place/single-cwd` | PlaceCWDs=`[A]`; sessions under A and B → only A |
| 4 | `place/multi-or` | PlaceCWDs=`[A,B]`; sessions A,B,C → A and B only (OR) |
| 5 | `place/no-match-empty` | PlaceCWDs points at unused path; sessions exist → empty list + format phrase |
| 6 | `place/abs-clean-equality` | PlaceCWD with trailing slash; session cwd cleaned abs → still matches |
| 7 | `place/dedup-overlapping` | PlaceCWDs lists same abs path twice; one session → single result |
| 8 | `place/empty-cwd-excluded` | Session with empty `info.cwd` + place filter on → not listed |
| 9 | `recent/within-window` | Recent=1h; session at Now−10m kept; Now−3h dropped |
| 10 | `recent/all-outside-empty` | All sessions older than window → empty |
| 11 | `recent/boundary-inclusive` | `last_active == Now - Recent` → kept (inclusive) |
| 12 | `active/only-file-active` | Two sessions; only one in active_sessions.json → that one |
| 13 | `active/none-listed-empty` | Active=true; active list empty → empty |
| 14 | `combo/place-and-recent` | Place A AND recent: in-place old dropped; other-cwd recent dropped; in-place recent kept |
| 15 | `combo/place-and-active` | Place A AND active: in-place inactive dropped; other-cwd active dropped |
| 16 | `combo/recent-and-active` | Recent AND active: recent-inactive and old-active dropped; recent-active kept |
| 17 | `combo/place-and-grep` | Place A + Grep token: out-of-place content match excluded; in-place match kept |
| 18 | `combo/recent-and-grep` | Recent + Grep: old matching content dropped; recent matching kept |
| 19 | `combo/place-recent-active` | Place×Recent×Active three-way AND smoke |
| 20 | `limit/after-all-filters` | Many place survivors; Limit=2 → exactly 2 newest survivors |
| 21 | `limit/default-limit-20` | Limit=0 after filters with 25 survivors → 20 |
| 22 | `errors/recent-nonpositive` | RecentSet with Recent=0 → error; no panic |
| 23 | `errors/empty-grep-when-set` | GrepSet with Grep="" → error |
| 24 | `role/main/keeps-main-class` | MainAgent: keep main + fork + empty-kind-no-parent; drop subagent family + empty+parent |
| 25 | `role/sub/keeps-sub-class` | SubAgent: keep subagent / resume / sub-fork + empty+parent; drop plain main + fork |
| 26 | `role/none/all-with-kind` | Neither role flag: all sessions returned; Kind tokens populated |
| 27 | `forked/matches` | Forked: keep kind=fork, kind=subagent_fork, non-empty forked_at |
| 28 | `forked/excludes-plain` | Forked: drop plain main and plain subagent without forked_at |
| 29 | `forked/whitespace-forked-at` | Forked: whitespace-only forked_at does not count as forked |
| 30 | `combo/main-and-forked` | MainAgent ∩ Forked: keep fork; drop plain main and subagent_fork |
| 31 | `combo/sub-and-forked` | SubAgent ∩ Forked: keep subagent_fork; drop plain subagent and main fork |
| 32 | `combo/place-and-main` | Place A ∩ MainAgent: keep main-in-A; drop sub-in-A and main-in-B |
| 33 | `format/header-kind-column` | FormatListTable header: SESSION ID then KIND then LAST ACTIVE… |
| 34 | `format/kind-tokens` | FormatListTable rows render main, sub, sub+, sub-f, fork |
| 35 | `format/empty-no-header` | Empty list FormatListTable still exactly `No sessions found` |
| 36 | `format/hits-kind-column` | FormatListTableWithHits header also includes KIND after SESSION ID |
| 37 | `errors/main-and-sub-exclusive` | MainAgent && SubAgent → error |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/list-filter
doctest test ./agent/grok/sessions/tests/list-filter
doctest test -v ./agent/grok/sessions/tests/list-filter/role/main/keeps-main-class
```

Existing place/recent/active/grep leaves stay GREEN. New role/forked/format
leaves are RED until Kind + role/fork filters + KIND column land.

```go
import (
	"reflect"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

// Request drives ListWithOptions via injectable opts only (no Chdir/Setenv).
type Request struct {
	TempDir  string
	GrokHome string
	Now      time.Time
	// Home is optional shorten-path base for FormatListTable.
	Home string

	Limit     int
	PlaceCWDs []string
	Recent    time.Duration
	RecentSet bool
	Active    bool
	Grep      string
	GrepSet   bool
	MainAgent bool
	SubAgent  bool
	Forked    bool

	// WantFormat fills Response.Output via FormatListTable when list succeeds.
	WantFormat bool
	// WantFormatHits fills Response.Output via FormatListTableWithHits
	// (sessions wrapped as SessionMatch with empty Hits; colorMode "never").
	WantFormatHits bool
}

type Response struct {
	Sessions []sessions.Session
	Output   string
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	now := req.Now
	if now.IsZero() {
		now = fixedNow
	}

	// Intended ListOptions fields (implementer):
	//   MainAgent, SubAgent, Forked bool
	// Wired via reflect so classic-TDD leaves compile before fields land;
	// once exported fields exist, they are set here and filters apply.
	opts := sessions.ListOptions{
		Limit:     req.Limit,
		PlaceCWDs: req.PlaceCWDs,
		Recent:    req.Recent,
		RecentSet: req.RecentSet,
		Active:    req.Active,
		Now:       now,
		Grep:      req.Grep,
		GrepSet:   req.GrepSet,
	}
	setListOptionsBool(&opts, "MainAgent", req.MainAgent)
	setListOptionsBool(&opts, "SubAgent", req.SubAgent)
	setListOptionsBool(&opts, "Forked", req.Forked)

	list, err := sessions.ListWithOptions(req.GrokHome, opts)
	resp := &Response{Sessions: list, Err: err}
	if err == nil && req.WantFormat {
		resp.Output = sessions.FormatListTable(list, req.Home, now)
	}
	if err == nil && req.WantFormatHits {
		matches := make([]sessions.SessionMatch, len(list))
		for i, s := range list {
			matches[i] = sessions.SessionMatch{Session: s}
		}
		resp.Output = sessions.FormatListTableWithHits(matches, req.Home, now, "never")
	}
	return resp, nil
}

// setListOptionsBool sets opts.Name when the field exists (RED until implementer
// adds MainAgent/SubAgent/Forked). No-op if the field is missing.
func setListOptionsBool(opts *sessions.ListOptions, name string, val bool) {
	v := reflect.ValueOf(opts).Elem().FieldByName(name)
	if v.IsValid() && v.CanSet() && v.Kind() == reflect.Bool {
		v.SetBool(val)
	}
}
```
