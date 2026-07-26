# sessions.RelocateCWD Tests

Doc-style tests for `sessions.RelocateCWD` in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Relocates a Grok session's
workspace cwd by updating explicit JSON fields and moving the session directory
under the URL-encoded new cwd key. **Classic TDD** — API not implemented yet;
leaves are RED until the implementer lands.

# DSN (Domain Specific Notion)

```
sessions.RelocateCWD(sessionID, targetDir, opts) (*RelocateCWDResult, error)

type RelocateCWDOptions struct {
    GrokHome string // empty → $GROK_HOME or ~/.grok
}

type RelocateCWDResult struct {
    OldCWD, NewCWD, OldSessionDir, NewSessionDir string
    FilesTouched []string // optional
}
```

**Call shape**: sessionID first (required), targetDir second (required), opts
last (nil OK).

**Grok home layout** (fixture-friendly):

```
$GROK_HOME/
  active_sessions.json          # active detection
  sessions/
    <url.PathEscape(abs_cwd)>/  # / → %2F (same as Grok CLI)
      <session-id>/
        summary.json            # info.cwd, git_root_dir?
        prompt_context.json     # working_directory
        updates.jsonl           # moved with dir; not bulk-rewritten
        …
    session_search.sqlite       # DO NOT touch
```

**Active detection** (fixture + implementer contract):

`active_sessions.json` under grok home. Prefer object form matching real Grok /
agenttty:

```json
{"sessions":[{"sessionId":"<id>","cwd":"/abs/path","openedAt":"2026-07-01T12:00:00Z"}]}
```

Also accept a bare JSON array of the same entry objects. A session is **active**
when any entry's `sessionId` or `session_id` equals the requested id (trimmed).
If the file is missing or `{}` / empty `sessions`, the session is not active.

**Migration (v1)**:

1. Validate non-empty sessionID; resolve home; find `sessions/**/<sessionID>/`
2. If active → error, no migration
3. targetDir must exist and be a directory (Abs)
4. Read old cwd from `summary.json` `info.cwd`
5. Update explicit fields only:
   - `summary.json` `info.cwd` → new abs path
   - `summary.json` `git_root_dir` if present and equal to old cwd → new
   - `prompt_context.json` `working_directory` → new (if file exists)
6. Move `sessions/<encode(old)>/<id>/` → `sessions/<encode(new)>/<id>/`
7. Do not rewrite `updates.jsonl` body; do not touch `session_search.sqlite`;
   do not bulk-rewrite `chat_history.jsonl`
8. Collision at destination session path → error

Tests use **filesystem fixtures only** (no real `grok` binary / no LLM).

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/relocate-cwd/
├── DOCTEST.md
├── SETUP.md
├── happy-path/                  inactive session; fields + dir move; sqlite untouched
├── missing-session/             unknown id → not found error
├── target-missing/              target path does not exist → error
├── target-not-dir/              target is a regular file → error
├── active-session-rejected/     active_sessions.json marks id → error; no move
├── empty-session-id/            sessionID "" → error
└── custom-grok-home/            opts.GrokHome selects fixture home only
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `happy-path` | Fixture under encoded old cwd; summary + prompt_context + updates; target exists; inactive → OK; dir moved; `info.cwd` + `working_directory` = target; updates content preserved; sqlite bytes unchanged |
| 2 | `missing-session` | No session dir for id → error mentioning session id |
| 3 | `target-missing` | Valid session; target path missing → error |
| 4 | `target-not-dir` | Target is a file → error |
| 5 | `active-session-rejected` | Session listed in `active_sessions.json` → error; **no** move / no field change |
| 6 | `empty-session-id` | `sessionID == ""` → error |
| 7 | `custom-grok-home` | `opts.GrokHome` = temp home with fixture; relocates under that home only (decoy home untouched) |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/relocate-cwd
doctest test ./agent/grok/sessions/tests/relocate-cwd
doctest test -v ./agent/grok/sessions/tests/relocate-cwd/happy-path
```

Classic TDD: all leaves RED until implementer adds `RelocateCWD` and related types.

```go
import (

	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	// GrokHome is the fixture grok home; passed as opts.GrokHome unless UseNilOpts.
	GrokHome string
	// TempDir is the per-test temp root (workspaces, decoy homes, etc.).
	TempDir string
	// SessionID is the first argument to RelocateCWD.
	SessionID string
	// TargetDir is the second argument (must exist as dir for success leaves).
	TargetDir string
	// OldCWD is the pre-migration workspace absolute path (for path asserts).
	OldCWD string
	// SessionDir is the pre-migration session directory path.
	SessionDir string
	// UpdatesMarker is the exact updates.jsonl body when seeded.
	UpdatesMarker string
	// SQLitePath / SQLiteMarker for "do not touch sqlite" asserts.
	SQLitePath   string
	SQLiteMarker string
	// UseNilOpts when true passes opts == nil (empty-session-id can still set home via env if needed).
	UseNilOpts bool
	// OptsGrokHome overrides opts.GrokHome when set (custom-grok-home).
	// When empty and !UseNilOpts, opts.GrokHome = GrokHome.
	OptsGrokHome string
	// DecoyGrokHome is an alternate home that must remain untouched (custom-grok-home).
	DecoyGrokHome string
	// Active when true, leaf expects active rejection (seeded in leaf Setup).
	Active bool
}

type Response struct {
	Result *sessions.RelocateCWDResult
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	var opts *sessions.RelocateCWDOptions
	if !req.UseNilOpts {
		home := req.OptsGrokHome
		if home == "" {
			home = req.GrokHome
		}
		opts = &sessions.RelocateCWDOptions{GrokHome: home}
	}
	result, err := sessions.RelocateCWD(req.SessionID, req.TargetDir, opts)
	return &Response{Result: result, Err: err}, nil
}

// compile-time reference so harness fragments keep imports used
var (
	_ = fmt.Sprintf
	_ = filepath.Join
)
```
