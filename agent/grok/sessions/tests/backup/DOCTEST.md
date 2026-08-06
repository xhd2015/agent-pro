# sessions.Backup Tests

Doc-style tests for `sessions.Backup` in
`github.com/xhd2015/agent-pro/agent/grok/sessions`. Builds a self-describing
session backup directory (`manifest.json` + `payload/`) with optional
`.tar.gz` archive. **Classic TDD** for `--dry-run`: live backup leaves stay
GREEN; new `dry-run/` leaves are RED until the implementer lands `DryRun`.

# DSN (Domain Specific Notion)

Portable **backup** of one Grok session (parent tree + related children by
default) into a self-describing directory for future restore. No restore in
v1. Manifest is the single source of truth (no `manifest.md`). Optional
**dry-run** plans the backup without writing directory, archive, or manifest.

**Participants**

- **Caller** — CLI `agent-pro grok session backup` or in-process client.
- **`Backup`** — `Backup(sessionID string, opts *BackupOptions) (*BackupResult, error)`.
  Finds session; enforces busy gate; copies payload; writes `manifest.json`;
  optionally archives. With `DryRun`, plans only (no writes).
- **`BackupOptions`** — injectables and output controls (see locked types).
- **`BackupResult`** — paths of written dir / archive / manifest plus identity
  fields; on dry-run, write paths empty and plan fields populated.
- **Busy gate** — hard error (no payload written) if **either**
  `IsFileActive(grokHome, sessionID)` **or** `LivePIDsForSession` non-empty
  (via `opts.Live`). **Same gate on dry-run** (still errors, still writes nothing).
- **Source layout** — `$GROK_HOME/sessions/<cwd_key>/<id>/`, sibling child
  dirs, per-cwd `prompt_history.jsonl`, `active_sessions.json`,
  `relocations/<id>.lock`, `logs/unified.jsonl`, `sessions/session_search.sqlite`.
- **Artifact layout** — `<dir>/manifest.json` + `<dir>/payload/...` (see below).
  Dry-run produces **no** on-disk artifact.

**Behaviors**

```
sessionID + BackupOptions
  -> resolve GrokHome; Find(session) ; missing -> "grok session not found: <id>"
  -> busy: IsFileActive OR LivePIDs non-empty -> error, no payload (live + dry-run)
  -> if DryRun:
       -> still validate ArchivePath ends with .tar.gz (else error)
       -> skip OutDir exist/non-empty checks; skip ArchivePath already-exists check
       -> walk plan (parent + optional children, prompt/lock/meta sizing)
       -> return BackupResult{DryRun:true, Dir/ArchivePath/ManifestPath empty,
            PlannedFiles, PlannedBytes, RelatedSessions} — write nothing
  -> (live) resolve OutDir (empty -> MkdirTemp kept) / ArchivePath validation
  -> copy parent tree; if IncludeChildren: copy each child_session_id from
       subagents/*/meta.json
  -> filter prompt_history lines for parent+child ids
  -> extract active_sessions entry if present (note: with either-signal busy
       gate a listed parent aborts; extract remains for completeness / race)
  -> copy relocations/<id>.lock if present
  -> scan logs/unified.jsonl for parent+child ids -> manifest.logs meta only
  -> note session_search.sqlite presence (never copy)
  -> write manifest.json with digests + check_results
  -> optional tar.gz of dir; keep dir when archive used
```

**Payload layout** (live only)

```text
<dir>/
  manifest.json
  payload/
    sessions/<cwd-key>/<parent-id>/          # recursive
    sessions/<cwd-key>/<child-id>/           # each included child
    sessions/<cwd-key>/prompt_history.session.jsonl
    bookkeeping/active_sessions.entry.json   # if entry present
    bookkeeping/relocations/<id>.lock        # if present
```

No logs under `payload/`. No `session_search.sqlite` under `payload/`.

**Output rules** (options / CLI mapping) — **live** (`DryRun=false`)

| OutDir | ArchivePath | Dir | Archive |
|--------|-------------|-----|---------|
| empty | empty | create temp, keep | none |
| set | empty | write to OutDir | none |
| empty | set `.tar.gz` | create temp, **keep** | create archive (must not exist) |
| set | set | write to OutDir | also create archive |

- `ArchivePath` must end with `.tar.gz` else error (live **and** dry-run).
- `ArchivePath` must not already exist else error (**live only**; dry-run skips).
- `OutDir` if exists and non-empty → error (**live only**; dry-run skips write checks).

**Dry-run rules** (`DryRun=true` / CLI `--dry-run`)

- Write **nothing**: no `OutDir` creation, no temp dir, no archive, no `manifest.json`.
- Busy gate identical to live.
- `OutDir` / `ArchivePath` are plan inputs only (CLI may print "Would backup…");
  result write paths stay empty.
- Digests not required (no on-disk files list with sha256).
- Plan metrics: `PlannedFiles > 0` when session content exists; `PlannedBytes >= 0`.

**Locked types**

```text
func Backup(sessionID string, opts *BackupOptions) (*BackupResult, error)

type BackupOptions struct {
    GrokHome        string        // empty → $GROK_HOME or ~/.grok
    OutDir          string        // empty → create temp dir (kept); dry-run: not created
    ArchivePath     string        // empty → no archive; must end ".tar.gz"
    IncludeChildren *bool         // nil → true (include children); false → skip
    Live            *LiveOptions  // PID busy gate injectables (tests always set)
    DryRun          bool          // true → plan only; write nothing
}

type BackupResult struct {
    Dir          string // absolute backup directory; empty when DryRun
    ArchivePath  string // absolute archive path, or ""; empty when DryRun
    ManifestPath string // Dir + "/manifest.json"; empty when DryRun
    SessionID    string
    CWD          string
    CWDKey       string // url.PathEscape(abs cwd)

    // Dry-run / plan fields (zero values on live success unless noted):
    DryRun          bool     // true when this result is a dry-run plan
    PlannedFiles    int      // estimated file count in plan (dry-run); 0 on live
    PlannedBytes    int64    // estimated payload bytes (dry-run); 0 on live
    RelatedSessions []string // parent + included children (dry-run always; optional live)
}

// manifest.json (version 1) — field names locked
type BackupManifest struct {
    Version         int                    `json:"version"`          // 1
    Kind            string                 `json:"kind"`             // "agent-pro.grok.session.backup"
    CreatedAt       string                 `json:"created_at"`       // RFC3339
    GrokHome        string                 `json:"grok_home"`
    SessionID       string                 `json:"session_id"`
    CWD             string                 `json:"cwd"`
    CWDKey          string                 `json:"cwd_key"`
    RelatedSessions []string               `json:"related_sessions"` // parent + included children
    Files           []BackupFileEntry      `json:"files"`
    Checks          []string               `json:"checks"`
    CheckResults    map[string]BackupCheck `json:"check_results"`
    Logs            BackupLogsMeta         `json:"logs"`
    SQLite          BackupSQLiteNote       `json:"sqlite"`
    Stats           map[string]any         `json:"stats,omitempty"`
    Warnings        []string               `json:"warnings,omitempty"`
}

type BackupFileEntry struct {
    Path   string `json:"path"`   // relative to backup dir (e.g. payload/sessions/...)
    Source string `json:"source"` // absolute source path when applicable
    Bytes  int64  `json:"bytes"`
    SHA256 string `json:"sha256"`
    Role   string `json:"role"`   // session_file | prompt_history | active_entry | relocation_lock | ...
}

type BackupCheck struct {
    OK     bool   `json:"ok"`
    Detail string `json:"detail,omitempty"`
}

type BackupLogsMeta struct {
    Path       string          `json:"path"`        // source logs/unified.jsonl abs or rel note
    MatchCount int             `json:"match_count"`
    LastLines  []BackupLogLine `json:"last_lines"`  // len ≤ 3, file order (last matches)
}

type BackupLogLine struct {
    Line int    `json:"line"` // 1-based source line number
    Text string `json:"text"`
    Time string `json:"time,omitempty"` // parsed when present (e.g. "ts" field)
}

type BackupSQLiteNote struct {
    Path    string `json:"path"`
    Present bool   `json:"present"`
    Note    string `json:"note,omitempty"`
}
```

**Checks** (names may include): digests after copy, summary id matches,
prompt_history id counts, child set present (when included), sqlite note,
busy not applicable post-copy. Successful backups should have
`check_results[*].ok == true` for required checks (or all recorded checks ok).

Tests use **filesystem fixtures** under `t.TempDir()/.grok` and injectable
`LiveOptions` — never real `~/.grok` / live `ps`/`lsof`.

## Version

0.0.2

## Decision Tree

```
agent/grok/sessions/tests/backup/
├── DOCTEST.md
├── SETUP.md
├── success/                         # live Backup returns result, writes artifact
│   ├── happy-path/                  # temp dir default; parent+1 child; payload+manifest
│   ├── no-children/                 # IncludeChildren=false → child dir not copied
│   ├── out-dir-writes/              # explicit OutDir
│   ├── archive-tar-gz/              # ArchivePath .tar.gz created; dir kept
│   ├── prompt-history-filtered/     # only matching session_id lines
│   ├── logs-meta-only/              # manifest.logs meta; no log file in payload
│   └── sqlite-not-copied/           # sqlite not under payload; note present
├── dry-run/                         # DryRun=true: plan only, write nothing
│   ├── happy-plan/                  # plan ok; OutDir not created; PlannedFiles; children
│   ├── no-children/                 # IncludeChildren=false → child not in RelatedSessions
│   ├── out-dir-not-written/         # OutDir (+ optional archive) never created
│   ├── file-active-errors/          # busy file-active still errors; nothing written
│   ├── pid-live-errors/             # busy live pid still errors; nothing written
│   ├── unknown-session/             # Find fails under dry-run
│   └── archive-suffix/              # ArchivePath without .tar.gz errors even in dry-run
└── errors/                          # live Backup returns error; no usable payload
    ├── unknown-session/             # Find fails
    ├── file-active-errors/          # IsFileActive → error
    ├── pid-live-errors/             # LivePIDs non-empty → error
    ├── archive-must-suffix/         # ArchivePath not ending .tar.gz
    └── archive-exists/              # ArchivePath already exists
```

Parameter ranking (most → least significant):

1. **Mode** — live write vs dry-run plan
2. **Outcome** — success vs error
3. **Error class** — unknown / busy (file|pid) / archive validation
4. **Output mode** — temp / OutDir / archive (plan inputs on dry-run)
5. **Children** — include vs skip
6. **Payload / meta edges** — prompt filter, logs meta, sqlite note (live)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `success/happy-path` | Inactive parent+1 child; prompt lines; relocation lock; temp OutDir; payload complete; manifest kind/version; related_sessions; checks ok |
| 2 | `success/no-children` | `IncludeChildren=false` → child session dir not under payload; parent still copied |
| 3 | `success/out-dir-writes` | Explicit `OutDir` receives `manifest.json` + `payload/` |
| 4 | `success/archive-tar-gz` | `ArchivePath` ends `.tar.gz`; archive created; backup dir kept |
| 5 | `success/prompt-history-filtered` | Extract contains only lines for parent/child session ids |
| 6 | `success/logs-meta-only` | `logs.match_count=N`; `last_lines` len≤3 with time when parseable; no log file under payload |
| 7 | `success/sqlite-not-copied` | Source sqlite present; not under payload; manifest.sqlite.present true |
| 8 | `errors/unknown-session` | Missing session → `grok session not found` + id |
| 9 | `errors/file-active-errors` | Listed in active_sessions → error; OutDir has no payload |
| 10 | `errors/pid-live-errors` | Injected live PID → error; no payload |
| 11 | `errors/archive-must-suffix` | ArchivePath without `.tar.gz` → error |
| 12 | `errors/archive-exists` | ArchivePath already exists → error |
| 13 | `dry-run/happy-plan` | `DryRun=true`; plan succeeds; OutDir not created; `PlannedFiles>0`; children in `RelatedSessions` |
| 14 | `dry-run/no-children` | Dry-run + `IncludeChildren=false` → child id not in plan `RelatedSessions` |
| 15 | `dry-run/out-dir-not-written` | Dry-run with OutDir + ArchivePath set → neither path created on disk |
| 16 | `dry-run/file-active-errors` | Dry-run still errors on file-active; nothing written |
| 17 | `dry-run/pid-live-errors` | Dry-run still errors on live pid; nothing written |
| 18 | `dry-run/unknown-session` | Dry-run unknown id → not found error |
| 19 | `dry-run/archive-suffix` | Dry-run still rejects ArchivePath without `.tar.gz` |

## How to Run

```sh
doctest vet ./agent/grok/sessions/tests/backup
doctest test ./agent/grok/sessions/tests/backup
doctest test -v ./agent/grok/sessions/tests/backup/success/happy-path
doctest test -v ./agent/grok/sessions/tests/backup/dry-run
```

Classic TDD: live `success/` + `errors/` stay GREEN; new `dry-run/` leaves RED
until implementer adds `BackupOptions.DryRun` and dry-run result fields.

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

	// CWD / CWDKey of the parent session fixture (absolute cwd).
	CWD    string
	CWDKey string

	// ChildSessionID when a linked child session is seeded (may be empty).
	ChildSessionID string

	// Output controls (map to BackupOptions).
	OutDir      string
	ArchivePath string
	// NoChildren when true sets IncludeChildren=false; otherwise IncludeChildren=true.
	NoChildren bool
	// DryRun maps to BackupOptions.DryRun (plan only; write nothing).
	DryRun bool

	// Injectable live snapshot for busy gate (always wire ListProcs/Lsof).
	Procs     []FixtureProc
	OpenFiles map[int][]string

	// Markers for asserts.
	ParentMarker   string // content unique to parent summary/file
	ChildMarker    string // content unique to child
	PromptNoiseID  string // session id of noise lines in prompt_history
	SQLiteMarker   string
	SQLitePath     string
	LogMatchCount  int    // expected manifest.logs.match_count when set > 0
	RelocationPath string // source lock path when seeded
	ExpectActive   bool   // leaf expects file-active busy error
}

type Response struct {
	Result *sessions.BackupResult
	Err    error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

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

	include := true
	if req.NoChildren {
		include = false
	}
	opts := &sessions.BackupOptions{
		GrokHome:        req.GrokHome,
		OutDir:          req.OutDir,
		ArchivePath:     req.ArchivePath,
		IncludeChildren: &include,
		Live:            live,
		DryRun:          req.DryRun,
	}

	result, err := sessions.Backup(req.SessionID, opts)
	return &Response{Result: result, Err: err}, nil
}

var _ = filepath.Join
```
