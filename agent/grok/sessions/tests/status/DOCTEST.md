# sessions.Status / LivePIDs / IsFileActive Tests

Doc-style tests for dual-signal Grok session status in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Covers package APIs
`Status`, `LivePIDsForSession`, `IsFileActive`, and format helpers for CLI
`agent-pro grok session status|info`. **Classic TDD** — surfaces not
implemented yet; leaves are RED until the implementer lands.

# DSN (Domain Specific Notion)

Dual-signal status: **file-active** (from `active_sessions.json`) plus
**live PIDs** (open-file hard hits on grok runners). Rollup state is
`running` | `marked-active` | `inactive`.

**Participants**

- **Caller** — CLI (`session status|info`) or in-process client that needs
  session liveness without scanning real processes in tests.
- **`Status`** — `Status(grokHome, sessionID string, checkPID bool, live *LiveOptions) (*SessionStatus, error)`.
  Finds the session first (unknown id → error). Reads file-active; when
  `checkPID`, scans live PIDs via injectables.
- **`IsFileActive`** — export of former unexported `isSessionActive(grokHome, sessionID) (bool, error)`.
  Object form `{"sessions":[{sessionId|session_id:...}]}` or bare JSON array;
  missing/`{}`/empty → inactive.
- **`LivePIDsForSession`** — `LivePIDsForSession(sessionID string, opts *LiveOptions) ([]LivePID, error)`.
  Scans `ListProcs` for **grok runners** only (same classify as procresolve:
  basename `grok`, exclude pure `grok update`). For each runner, `Lsof(pid)`
  open paths; hard-hit when path parses as
  `…/.grok/sessions/…/<uuid>/…` and uuid equals `sessionID` (case-insensitive
  uuid match, same rules as `pkgs/procresolve` `parseSessionFromPath`).
  Return **all** matches sorted by PID ascending. `Name = filepath.Base(argv0)`;
  `Cmd` = full command line.
- **`LiveOptions`** — injectable `ListProcs func() []procresolve.Proc` and
  `Lsof func(int) []string`. Tests always inject; production may wire
  `ListLiveProcs` / `LiveLsof` (or `dot-pkgs/go-pkgs/proc`).
- **`SessionStatus`** — `SessionID`, `Path` (absolute `summary.json` from Find),
  `FileActive`, `PIDs []LivePID`, `PIDChecked` (false when `checkPID==false` /
  CLI `--no-pid`), `State` string: `running` | `marked-active` | `inactive`.
- **Rollup** — if `PIDChecked && len(PIDs)>0` → `running`; else if
  `FileActive` → `marked-active`; else → `inactive`.
- **Formatters** (CLI text/json + info Active block):
  - `FormatStatusText(st *SessionStatus) string` — State, File yes/no,
    Path (`pathfmt.TildeHome`), PID lines (`pid` + `name`) or none / skipped.
  - `FormatStatusJSON(st *SessionStatus) (string, error)` — no ANSI; fields
    `session_id`, `state`, `file_active`, `pid_checked`, `path` (absolute),
    `pids` array of `{pid,name,cmd}`.
  - `FormatActiveBlock(st *SessionStatus) string` — dual-signal Active section
    for `session info` (file + pid lines); CLI appends after existing
    `FormatInfoText` (signature of FormatInfoText unchanged).

**Behaviors**

```
# status pipeline
sessionID + grokHome + checkPID + LiveOptions
  -> Find(session) ; missing -> error "grok session not found: <id>"
  -> IsFileActive(grokHome, sessionID)
  -> if checkPID: LivePIDsForSession(sessionID, live)
  -> rollup State; PIDChecked = checkPID

# live pid scan (no live processes in tests)
ListProcs() -> filter grok runners (skip grok update, skip non-grok)
  -> Lsof(pid) -> parseSessionFromPath hard hit on sessionID
  -> all matches sorted by PID

# --no-pid / checkPID=false
  -> PIDs empty; PIDChecked false; State from FileActive only
  -> injectable ListProcs ignored

# format
Status result -> FormatStatusText | FormatStatusJSON | FormatActiveBlock
```

**Locked types**

```text
LivePID
  PID int
  Name string  // filepath.Base(argv0)
  Cmd  string  // full command line

LiveOptions
  ListProcs func() []procresolve.Proc
  Lsof      func(int) []string

SessionStatus
  SessionID  string
  Path       string  // absolute summary.json
  FileActive bool
  PIDs       []LivePID
  PIDChecked bool
  State      string  // "running" | "marked-active" | "inactive"

IsFileActive(grokHome, sessionID string) (bool, error)
LivePIDsForSession(sessionID string, opts *LiveOptions) ([]LivePID, error)
Status(grokHome, sessionID string, checkPID bool, live *LiveOptions) (*SessionStatus, error)
FormatStatusText(st *SessionStatus) string
FormatStatusJSON(st *SessionStatus) (string, error)
FormatActiveBlock(st *SessionStatus) string
```

Open-file hard hit path shape (fixture):

```text
…/.grok/sessions/<encoded-or-any>/<uuid>/…
e.g. /tmp/fake/.grok/sessions/%2Fproj/019f…/events.jsonl
```

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/status/
├── DOCTEST.md
├── SETUP.md
├── running/                 # file active + live pid → state running
├── marked-active/           # file active, no matching pid → marked-active
├── inactive/                # not file-active, no pid → inactive
├── no-pid/                  # checkPID=false; injectable pids ignored; state from file
├── unknown-session/         # missing session → error
├── multi-pid/               # two matching pids → sorted by PID; names from basename
├── skip-non-runners/        # bash + grok update with open paths → not counted
├── format-text/             # FormatStatusText: State / File / PID lines
├── format-json/             # FormatStatusJSON fields; no ANSI
└── format-active-block/     # FormatActiveBlock for info dual-signal section
```

Parameter ranking (most → least significant):

1. **Outcome class** — running / marked-active / inactive / no-pid / error
2. **Live scan edges** — multi-pid sort / non-runner exclusion
3. **Format** — structured Status vs text / json / active-block

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `running` | Session on disk; `active_sessions.json` lists id; injected grok pid with open-file hard hit → `State=running`, `FileActive=true`, `PIDChecked=true`, one PID |
| 2 | `marked-active` | File-active; ListProcs empty or no matching open files → `State=marked-active`, `PIDs` empty |
| 3 | `inactive` | Session exists; not in active list; no live hits → `State=inactive` |
| 4 | `no-pid` | `checkPID=false` with injectable pids that *would* match → `PIDChecked=false`, `PIDs` empty, state from file only (`marked-active`) |
| 5 | `unknown-session` | No session dir / Find fails → error containing `grok session not found` and id |
| 6 | `multi-pid` | Two grok runners both hard-hit same session → `PIDs` length 2, sorted by PID asc, `Name` = basename |
| 7 | `skip-non-runners` | Bash + `grok update` hold open session paths; only a real grok runner (if any) counts — here none → empty PIDs, not running from those alone |
| 8 | `format-text` | Running fixture → `FormatStatusText` has state, file yes, Path (summary.json), pid+name lines |
| 9 | `format-json` | Running fixture → JSON fields including absolute `path`; no ANSI |
| 10 | `format-active-block` | Running fixture → `FormatActiveBlock` mentions file active and live pid |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/status
doctest test ./agent/grok/sessions/tests/status
doctest test -v ./agent/grok/sessions/tests/status/running
```

Classic TDD: all leaves RED until implementer adds Status / LivePIDs / IsFileActive / format helpers.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

// FixtureProc maps 1:1 to procresolve.Proc for injectable ListProcs.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

type Request struct {
	GrokHome  string
	TempDir   string
	SessionID string

	// NoPID maps to Status checkPID=false (CLI --no-pid). Default: check PIDs.
	NoPID bool

	// Injectable live snapshot (always set for PID-checking leaves).
	Procs     []FixtureProc
	OpenFiles map[int][]string

	// Format selects formatter after Status succeeds.
	// "" = structured Status only; "text" | "json" | "active-block".
	Format string

	// Op selects entrypoint: "status" (default), "live-pids", "is-file-active".
	Op string
}

type Response struct {
	Status       *sessions.SessionStatus
	LivePIDs     []sessions.LivePID
	IsFileActive bool
	Output       string
	Err          error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	checkPID := !req.NoPID

	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}
	snap := make([]procresolve.Proc, 0, len(req.Procs))
	for _, p := range req.Procs {
		snap = append(snap, procresolve.Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd})
	}
	live := &sessions.LiveOptions{
		ListProcs: func() []procresolve.Proc { return snap },
		Lsof: func(pid int) []string {
			return files[pid]
		},
	}

	op := req.Op
	if op == "" {
		op = "status"
	}

	resp := &Response{}
	switch op {
	case "status":
		st, err := sessions.Status(req.GrokHome, req.SessionID, checkPID, live)
		resp.Status = st
		resp.Err = err
		if err != nil || st == nil {
			return resp, nil
		}
		switch req.Format {
		case "text":
			resp.Output = sessions.FormatStatusText(st)
		case "json":
			out, jerr := sessions.FormatStatusJSON(st)
			if jerr != nil {
				resp.Err = jerr
				return resp, nil
			}
			resp.Output = out
		case "active-block":
			resp.Output = sessions.FormatActiveBlock(st)
		}
	case "live-pids":
		pids, err := sessions.LivePIDsForSession(req.SessionID, live)
		resp.LivePIDs = pids
		resp.Err = err
	case "is-file-active":
		ok, err := sessions.IsFileActive(req.GrokHome, req.SessionID)
		resp.IsFileActive = ok
		resp.Err = err
	default:
		t.Fatalf("unknown Op %q", op)
	}
	return resp, nil
}

// keep imports used in generated harness fragments
var _ = filepath.Join
```