# sessions bookmark (pin + multi-runner catalog) Tests

Doc-style tests for session bookmark catalog APIs in
`github.com/xhd2015/agent-pro/agent/grok/sessions` (store may live in
`pkgs/sessionbookmark` re-exported/called from sessions — same call surface).
Covers Grok pin, list/show/remove, enrich modes (light default / stale / heavy),
tag merge, description pointer updates, and format helpers for CLI:

- `agent-pro grok session bookmark <id> …`
- `agent-pro bookmark [list|show|remove] …`

**Classic TDD** — EnrichMode surfaces not implemented yet (product still always
heavy `Find`); new enrich leaves are RED until the implementer lands. No product
implementation in this tree.

# DSN (Domain Specific Notion)

Session **bookmark catalog**: pin a Grok session into a durable multi-runner
store under agent-pro home; list/show/remove with mode-controlled live enrich
for grok rows; format table / show / JSON for CLI.

**Participants**

- **Caller** — CLI (`grok session bookmark` / root `bookmark list|show|remove`)
  or in-process client needing a pin catalog across agent runners.
- **`BookmarkGrok`** —
  `BookmarkGrok(agentProHome, grokHome, sessionID string, opts *PinOptions) (*Bookmark, created bool, err error)`.
  Resolves session via `Find(grokHome, id)` once (pin path **unchanged** by
  enrich modes); unknown → error containing `not found` (and id); no store
  write. On success upserts by key `(agent_runner="grok", session_id)`; always
  refreshes denormalized `session_dir`, `title`, `num_chat_messages` from live
  summary; applies tag / description rules from opts; sets `created_at` only on
  create; always sets `updated_at`. `created=true` on new entry, `false` on
  update.
- **`EnrichMode`** — list/show enrich control (`ListFilter.Enrich` / last arg of
  `GetBookmark`). Zero value = **EnrichLight** (fast default).
  - **EnrichLight** (default): for grok rows use stored `session_dir` only —
    read that dir's `summary.json` (no `filepath.WalkDir` / no `Find`). Readable
    summary → refresh Title, NumChatMessages, SessionDir; `Orphaned=false`.
    Missing/unreadable dir or summary → keep snapshot, `Orphaned=true`, warning
    like `session <id> is bookmarked but not found under GROK_HOME`. Empty
    `session_dir` → stored fields, `Orphaned=false`, **no Find**. Non-grok
    rows: stored only (`Orphaned=false` v1).
  - **EnrichOff** (`--stale`): catalog snapshot only; no FS checks;
    `Orphaned=false` (not computed); no orphan warnings.
  - **EnrichHeavy** (`--enrich`): try light first; if not orphaned and live
    data obtained, done; else `Find(grokHome, id)`; success → refresh +
    `Orphaned=false`; both fail → `Orphaned=true` + warning.
- **`ListBookmarks`** —
  `ListBookmarks(agentProHome, grokHome string, filter ListFilter) (views []BookmarkView, warnings []string, err error)`.
  Missing store → empty list (no error). Corrupt JSON → error; do not wipe.
  Enrich per `filter.Enrich` (zero = light). Filter: `Runner` (""=all), `Tags`
  AND match, `Limit` (0=unlimited).
- **`GetBookmark`** —
  `GetBookmark(agentProHome, runner, sessionID, grokHome string, enrich EnrichMode) (*BookmarkView, error)`.
  `runner==""` → unique match by session_id; 0 → not found; 2+ → error asking
  for runner. Enrich same modes as list (locked as explicit last arg).
- **`RemoveBookmark`** —
  `RemoveBookmark(agentProHome, runner, sessionID string) error`. Unique-match
  rules same as Get when runner empty; not found → error.
- **Formatters** — `FormatBookmarksTable(views)`, `FormatBookmarkShow(view)`,
  `FormatBookmarkJSON(v any)` (single or list); no ANSI in JSON. Table style
  matches sibling formatters (TrimRight OK); asserts prefer contains.
- **Store file** — `{agentProHome}/session_bookmarks.json` version 1 object
  with `bookmarks` array. Fields: agent_runner, session_id, session_dir
  (absolute), title, num_chat_messages, tags (sorted unique), description,
  created_at, updated_at (RFC3339).

**Behaviors**

```
# pin (Grok) — still Find once; enrich modes do not apply
agentProHome + grokHome + sessionID + PinOptions
  -> Find(grokHome, sessionID); missing -> error "not found"; no write
  -> load store (missing -> empty); corrupt -> error; no clobber
  -> upsert key (grok, sessionID): refresh denorm from live
  -> tags: nil keep; non-nil union merge; ClearTags wipe then merge
  -> description: nil keep; non-nil set (incl empty)
  -> tags normalize: trim, drop empty, dedupe case-sensitive, sort
  -> write session_bookmarks.json; return (*Bookmark, created, nil)

# list (enrich mode)
agentProHome + grokHome + ListFilter{..., Enrich}
  -> load store; missing -> empty views; corrupt -> error
  -> EnrichOff: stored snapshot; Orphaned=false; no warnings; no FS
  -> EnrichLight: session_dir + summary.json only; NEVER Find
  -> EnrichHeavy: light first; if orphaned/empty dir then Find; orphan if both fail
  -> filter runner / AND tags / limit
  -> return views, warnings, nil

# show
  -> unique key match (runner optional when unique)
  -> enrich with same EnrichMode as list

# remove
  -> unique key match; delete entry and rewrite store
```

**Locked types**

```text
type EnrichMode int

const (
  EnrichLight EnrichMode = iota // 0 = default: session_dir + summary.json; never Find
  EnrichOff                     // --stale: catalog only; Orphaned=false; no FS
  EnrichHeavy                   // --enrich: light then Find if needed
)

type Bookmark struct {
  AgentRunner     string
  SessionID       string
  SessionDir      string
  Title           string
  NumChatMessages int
  Tags            []string
  Description     string
  CreatedAt       time.Time
  UpdatedAt       time.Time
}

type BookmarkView struct {
  Bookmark
  Orphaned bool
}

type PinOptions struct {
  Tags        []string // nil = keep on update; non-nil = merge (union)
  Description *string  // nil = keep; non-nil = set (including "")
  ClearTags   bool     // wipe tags before merge
}

type ListFilter struct {
  Runner string     // "" = all
  Tags   []string   // AND
  Limit  int        // 0 = unlimited
  Enrich EnrichMode // zero = EnrichLight
}

BookmarkGrok(agentProHome, grokHome, sessionID string, opts *PinOptions) (*Bookmark, bool, error)
ListBookmarks(agentProHome, grokHome string, filter ListFilter) ([]BookmarkView, []string, error)
GetBookmark(agentProHome, runner, sessionID, grokHome string, enrich EnrichMode) (*BookmarkView, error)
RemoveBookmark(agentProHome, runner, sessionID string) error
FormatBookmarksTable(views []BookmarkView) string
FormatBookmarkShow(view *BookmarkView) string
FormatBookmarkJSON(v any) (string, error)
```

**CLI mapping (implementer; L2 tests call package APIs only)**

```text
agent-pro grok session bookmark <session-id> [-t|--tag]... [-d|--description] [--clear-tags] [--json]
agent-pro bookmark              → list (default EnrichLight)
agent-pro bookmark list [--runner] [--tag]... [--limit] [--stale|--enrich] [--json]
agent-pro bookmark show <id> [--runner] [--stale|--enrich] [--json]
agent-pro bookmark remove|rm|unbookmark <id> [--runner] [--json]
agent-pro grok session bookmarks → list --runner grok
agent-pro grok session unbookmark <id> → remove --runner grok
```

`--stale` and `--enrich` are mutually exclusive (CLI error). Default neither → light.
Help: default is cheap session_dir refresh; `--enrich` is slow on large GROK_HOME.
Pin path unchanged (still Find once). CLI flag conflict is implementer/CLI-only
(package takes a single EnrichMode).

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/bookmark/
├── DOCTEST.md
├── SETUP.md
├── success/                         # happy pin / list / show / remove / format
│   ├── pin-new/                     # first pin creates store; created=true
│   ├── pin-with-tags/               # tags sorted/deduped/trimmed
│   ├── pin-update-merge-tags/       # re-pin merges tags; created=false; created_at stable
│   ├── pin-update-description/      # Description pointer sets; tags kept when nil
│   ├── pin-clear-tags/              # ClearTags wipes then optional merge
│   ├── pin-bare/                    # no tags/desc still succeeds
│   ├── list-after-pin/              # pin then list roundtrip (default light)
│   ├── list-table/                  # FormatBookmarksTable headers + fields
│   ├── list-filter-runner/          # filter Runner=grok only
│   ├── list-filter-tag/             # AND tags filter
│   ├── show-detail/                 # Get + FormatBookmarkShow fields
│   ├── show-unique-no-runner/       # runner "" unique match
│   ├── remove/                      # gone; second remove errors
│   └── json-output/                 # FormatBookmarkJSON; no ANSI; key fields
├── errors/                          # fail closed; no silent clobber
│   ├── pin-unknown-session/         # Find fails; store absent/unchanged
│   ├── pin-empty-id/                # empty session id errors
│   ├── remove-missing/              # not found
│   ├── show-missing/                # not found
│   └── corrupt-store/               # list/pin error; file not wiped
├── edge/                            # hybrid / empty / multi-runner
│   ├── orphan-list/                 # pin then delete session; light orphan + warning
│   ├── missing-store-list/          # no file → empty list / empty table
│   ├── preseed-other-runner/        # codex + grok rows both list
│   └── show-ambiguous-needs-runner/ # same session_id two runners → error
└── enrich/                          # EnrichMode: light default / stale / heavy
    ├── list-default-light-refreshes/  # mutate summary; default light shows live
    ├── list-stale-ignores-summary/    # EnrichOff keeps stored title/msgs
    ├── list-light-orphan-no-find/     # wrong session_dir; Find decoy exists; light orphans
    ├── list-heavy-recovers-via-find/  # wrong/empty session_dir; heavy Find recovers
    ├── list-heavy-still-orphan/       # no session anywhere; heavy still orphan
    └── show-default-light/            # GetBookmark light refreshes from session_dir
```

Parameter ranking (most → least significant):

1. **Outcome class** — success / errors / edge / enrich
2. **Operation** — pin / list / show / remove / format
3. **Enrich mode** — light (default) / off (stale) / heavy
4. **Pin mutation** — create vs update tags/description/clear
5. **Filter / identity** — runner, tags, unique vs ambiguous
6. **Store state** — missing / corrupt / preseed multi-runner / orphan / decoy path

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `success/pin-new` | Live session exists; first pin creates `session_bookmarks.json`; `agent_runner=grok`; denorm fields set; `created=true` |
| 2 | `success/pin-with-tags` | Pin with messy tags → stored tags sorted, trimmed, deduped |
| 3 | `success/pin-update-merge-tags` | Preseed bookmark + re-pin with new tags → union; `created=false`; `created_at` stable; denorm refreshed |
| 4 | `success/pin-update-description` | Preseed description; pin with non-nil Description sets new; Tags nil keeps tags |
| 5 | `success/pin-clear-tags` | ClearTags true clears prior tags; optional new tags after wipe |
| 6 | `success/pin-bare` | opts nil / empty still creates bookmark with empty tags and description |
| 7 | `success/list-after-pin` | Pin then ListBookmarks returns one grok view with matching id/title |
| 8 | `success/list-table` | Preseed views → FormatBookmarksTable contains RUNNER / SESSION / MSGS / TITLE (and tags) |
| 9 | `success/list-filter-runner` | Store has grok+codex; filter Runner=grok → only grok |
| 10 | `success/list-filter-tag` | AND tags filter returns only matching bookmarks |
| 11 | `success/show-detail` | GetBookmark + FormatBookmarkShow includes runner, id, title, msgs, dir, tags, description |
| 12 | `success/show-unique-no-runner` | Single match; runner "" succeeds |
| 13 | `success/remove` | Remove deletes entry; second remove errors with not found |
| 14 | `success/json-output` | FormatBookmarkJSON of view/list has key fields; no ANSI |
| 15 | `errors/pin-unknown-session` | Missing session → error not found; store file absent |
| 16 | `errors/pin-empty-id` | Empty session id → error |
| 17 | `errors/remove-missing` | Remove unknown → error |
| 18 | `errors/show-missing` | Show unknown → error |
| 19 | `errors/corrupt-store` | Corrupt JSON → list and pin error; file content not replaced with empty store |
| 20 | `edge/orphan-list` | Pin, delete session dir, list default light → Orphaned=true, warning, stored title kept |
| 21 | `edge/missing-store-list` | No store file → empty views; table empty / "No bookmarks" style |
| 22 | `edge/preseed-other-runner` | Manual JSON with codex+grok; list shows both |
| 23 | `edge/show-ambiguous-needs-runner` | Same session_id for two runners; Get with runner "" errors asking for runner |
| 24 | `enrich/list-default-light-refreshes` | Preseed + mutate summary title/msgs under session_dir; list default (light/`""`) shows live values; Orphaned=false; no warnings |
| 25 | `enrich/list-stale-ignores-summary` | Same mutate; EnrichOff/stale keeps **stored** title/msgs; Orphaned=false; warnings empty |
| 26 | `enrich/list-light-orphan-no-find` | Store wrong session_dir; live session still Find-able under grokHome; light → Orphaned=true + warning (never recovers via Find) |
| 27 | `enrich/list-heavy-recovers-via-find` | Wrong/empty session_dir; live session under grokHome; EnrichHeavy recovers title/dir; Orphaned=false |
| 28 | `enrich/list-heavy-still-orphan` | Catalog row with no session anywhere; EnrichHeavy → Orphaned=true + warning |
| 29 | `enrich/show-default-light` | GetBookmark default light refreshes Title/NumChatMessages from session_dir summary |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/bookmark
doctest test ./agent/grok/sessions/tests/bookmark
doctest test -v ./agent/grok/sessions/tests/bookmark/success/pin-new
doctest test -v ./agent/grok/sessions/tests/bookmark/errors
doctest test -v ./agent/grok/sessions/tests/bookmark/edge
doctest test -v ./agent/grok/sessions/tests/bookmark/enrich
```

Classic TDD: enrich leaves RED until implementer adds `EnrichMode`,
`ListFilter.Enrich`, light/off/heavy enrich paths, and
`GetBookmark(..., enrich EnrichMode)`. Existing leaves stay valid under
**default = light** (orphan-list still orphans when session_dir is gone). Full
suite may fail to compile until product signature matches DSN.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	TempDir      string
	AgentProHome string
	GrokHome     string

	// Op: pin | list | show | remove | pin-list | format-list | format-show | format-json
	// pin-list: BookmarkGrok then ListBookmarks (roundtrip leaves).
	Op string

	SessionID string
	Runner    string // show/remove/filter; "" = unique / all

	// PinOptions mapping
	Tags        []string
	TagsSet     bool   // true → pass Tags slice (even empty) as non-nil merge input
	Description *string
	ClearTags   bool
	// NilOpts when true passes opts=nil to BookmarkGrok.
	NilOpts bool

	// ListFilter mapping
	FilterTags []string
	Limit      int

	// Enrich: "" | "light" | "default" → EnrichLight;
	// "stale" | "off" → EnrichOff; "heavy" | "enrich" → EnrichHeavy.
	// Zero/empty maps to EnrichLight (product default).
	Enrich string

	// Format inputs (format-* ops); when empty, list/show ops may fill Output.
	FormatViews []sessions.BookmarkView
	FormatView  *sessions.BookmarkView
	FormatAny   any

	// Fixture markers written by Setup helpers.
	Title             string
	NumChatMessages   int
	// LiveTitle / LiveNumChatMessages: optional post-preseed summary mutation
	// expectations (enrich leaves). When set, Assert compares against live values.
	LiveTitle           string
	LiveNumChatMessages int
	FixtureCWD          string
	SessionDir          string
	// Stored snapshot fields when store title/msgs intentionally differ from live.
	StoredTitle           string
	StoredNumChatMessages int
	PreseedCreatedAt      time.Time
	PreseedUpdatedAt      time.Time
	CorruptMarker         string // exact corrupt file body when set
	ExpectStoreAbsent     bool
}

type Response struct {
	Bookmark  *sessions.Bookmark
	Created   bool
	Views     []sessions.BookmarkView
	View      *sessions.BookmarkView
	Warnings  []string
	Output    string
	StorePath string
	Err       error
}

// enrichMode maps req.Enrich strings to sessions.EnrichMode.
func enrichMode(s string) sessions.EnrichMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "light", "default":
		return sessions.EnrichLight
	case "stale", "off":
		return sessions.EnrichOff
	case "heavy", "enrich":
		return sessions.EnrichHeavy
	default:
		return sessions.EnrichLight
	}
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	resp := &Response{
		StorePath: filepath.Join(req.AgentProHome, "session_bookmarks.json"),
	}
	op := req.Op
	if op == "" {
		op = "pin"
	}
	mode := enrichMode(req.Enrich)

	switch op {
	case "pin":
		var opts *sessions.PinOptions
		if !req.NilOpts {
			opts = &sessions.PinOptions{
				Description: req.Description,
				ClearTags:   req.ClearTags,
			}
			if req.TagsSet {
				// Non-nil slice (may be empty) triggers merge semantics.
				tags := req.Tags
				if tags == nil {
					tags = []string{}
				}
				opts.Tags = tags
			}
		}
		bm, created, err := sessions.BookmarkGrok(req.AgentProHome, req.GrokHome, req.SessionID, opts)
		resp.Bookmark = bm
		resp.Created = created
		resp.Err = err

	case "list":
		views, warnings, err := sessions.ListBookmarks(req.AgentProHome, req.GrokHome, sessions.ListFilter{
			Runner: req.Runner,
			Tags:   req.FilterTags,
			Limit:  req.Limit,
			Enrich: mode,
		})
		resp.Views = views
		resp.Warnings = warnings
		resp.Err = err

	case "pin-list":
		var opts *sessions.PinOptions
		if !req.NilOpts {
			opts = &sessions.PinOptions{
				Description: req.Description,
				ClearTags:   req.ClearTags,
			}
			if req.TagsSet {
				tags := req.Tags
				if tags == nil {
					tags = []string{}
				}
				opts.Tags = tags
			}
		}
		bm, created, err := sessions.BookmarkGrok(req.AgentProHome, req.GrokHome, req.SessionID, opts)
		resp.Bookmark = bm
		resp.Created = created
		if err != nil {
			resp.Err = err
			return resp, nil
		}
		views, warnings, lerr := sessions.ListBookmarks(req.AgentProHome, req.GrokHome, sessions.ListFilter{
			Runner: req.Runner,
			Tags:   req.FilterTags,
			Limit:  req.Limit,
			Enrich: mode,
		})
		resp.Views = views
		resp.Warnings = warnings
		resp.Err = lerr

	case "show":
		view, err := sessions.GetBookmark(req.AgentProHome, req.Runner, req.SessionID, req.GrokHome, mode)
		resp.View = view
		resp.Err = err

	case "remove":
		resp.Err = sessions.RemoveBookmark(req.AgentProHome, req.Runner, req.SessionID)

	case "format-list":
		views := req.FormatViews
		if views == nil {
			var err error
			views, resp.Warnings, err = sessions.ListBookmarks(req.AgentProHome, req.GrokHome, sessions.ListFilter{
				Runner: req.Runner,
				Tags:   req.FilterTags,
				Limit:  req.Limit,
				Enrich: mode,
			})
			resp.Views = views
			if err != nil {
				resp.Err = err
				return resp, nil
			}
		}
		resp.Output = sessions.FormatBookmarksTable(views)
		resp.Views = views

	case "format-show":
		view := req.FormatView
		if view == nil {
			v, err := sessions.GetBookmark(req.AgentProHome, req.Runner, req.SessionID, req.GrokHome, mode)
			resp.View = v
			if err != nil {
				resp.Err = err
				return resp, nil
			}
			view = v
		}
		resp.Output = sessions.FormatBookmarkShow(view)
		resp.View = view

	case "format-json":
		var payload any = req.FormatAny
		if payload == nil {
			if req.FormatView != nil {
				payload = req.FormatView
			} else if req.FormatViews != nil {
				payload = req.FormatViews
			} else {
				views, warnings, err := sessions.ListBookmarks(req.AgentProHome, req.GrokHome, sessions.ListFilter{
					Runner: req.Runner,
					Tags:   req.FilterTags,
					Limit:  req.Limit,
					Enrich: mode,
				})
				resp.Views = views
				resp.Warnings = warnings
				if err != nil {
					resp.Err = err
					return resp, nil
				}
				payload = views
			}
		}
		out, err := sessions.FormatBookmarkJSON(payload)
		resp.Output = out
		resp.Err = err

	default:
		t.Fatalf("unknown Op %q", op)
	}
	return resp, nil
}

var _ = time.Time{}
```
