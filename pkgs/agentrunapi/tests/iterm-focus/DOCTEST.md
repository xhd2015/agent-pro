# agentrunapi — iTerm focus from session_id (P2)

Nested library + CLI doctests for **agent-run focus**: resolve a stored session to
iTerm candidate(s) via meta → registry PID → process-tree TTYs → `iterm2.FindByTTY`,
then apply 0/1/multi policy and optional focus.

Classic TDD (**RED** until implementer lands APIs + `agent-run focus` CLI). No live
iTerm or live process tree: inject `ListProcs`, `ListITerm`, `FocusITerm`.

| Phase | Status |
|-------|--------|
| **P2** process-tree TTYs + `FocusSession` / `FindITermForSession` + `agent-run focus` | Classic TDD — **RED** until implementer lands APIs |

**Out of scope:** local-bot CLI; name/title matching; auto-open window; P3 spl wire;
P4 release; L3 e2e.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — supplies `session_id`, optional `--index`, `--dry-run`; library
  injects Store + process/iTerm/focus hooks for tests.
- **Session store** — `agentstorage.Store` flat `sessions/<id>/meta.json`
  (`runner`, `terminal_session_id`, …).
- **TTY registry** — `{home}/<runner>-registry/<terminal_session_id>.json` with
  serve **PID** (TTY often `??` / blank — not used as the sole TTY source).
- **Process tree** — snapshot rows (`pid`, `ppid`, `tty`, `cmd`); walk
  **ancestors + descendants** of the registry PID; collect **real** TTYs only
  (skip `??`, blank, whitespace).
- **iTerm finder** — P1 `iterm2.FindByTTY` / injectable `ListITerm` →
  `[]iterm2.SessionRef`.
- **Focus policy** — 0 matches → error; 1 match → choose it; multi → require
  `Index` (0-based); OOB index → error.
- **Focus runner** — injectable `FocusITerm(ref)` (production: `iterm2.Focus`);
  **never** called when `DryRun`.
- **CLI** — `agent-run focus` via `agentruncli` (`Handle` switch + writer-capable
  `RunFocus` for L2).

**Behaviors**

```
session_id
  -> Store.GetSession -> meta (runner, terminal_session_id)
  -> registry <runner>-registry/<terminal_session_id>.json -> pid
  -> ListProcs tree: ancestors + descendants -> real TTYs only
  -> normalize TTYs; ListITerm + FindByTTY -> candidates
  -> 0 / 1 / multi(+index) policy
  -> FocusITerm(chosen) unless DryRun
```

- **CollectTTYsFromTree** — pure; given proc snapshot + root pid → real TTYs only.
- **FindITermForSession** — resolve + match; no focus side effect.
- **FocusSession** — find + policy + focus (unless dry-run); returns chosen
  candidate.
- **CLI** — `focus <session-id> [--index N] [--dry-run] [--session-id ID]`;
  help documents flags; dry-run never focuses; errors on stderr with `Error:`
  prefix for none/multi; multi lists `--index N` lines; trailing `\n` on help
  and primary stdout.

### Public API (Classic TDD — locked for implementer)

```go
// Package: github.com/xhd2015/agent-pro/pkgs/agentrunapi

// ProcRow is one process snapshot row for tree TTY collection.
type ProcRow struct {
    PID  int
    PPID int
    TTY  string // e.g. /dev/ttys148, "??", or ""
    Cmd  string
}

// FocusCandidate is one iTerm match for a session after tree TTY resolution.
type FocusCandidate struct {
    Index   int // 0-based list index (CLI --index uses the same)
    Ref     iterm2.SessionRef
    PID     int
    TTY     string
    Source  string // e.g. "tree"
    CmdHint string
}

// FocusOpts drives FindITermForSession / FocusSession.
type FocusOpts struct {
    Store     agentstorage.Store
    SessionID string
    Index     *int // nil = require unique match
    DryRun    bool

    // Injectables (nil => production implementations):
    ListProcs  func() []ProcRow
    ListITerm  func() ([]iterm2.SessionRef, error)
    FocusITerm func(iterm2.SessionRef) error
    // optional: ReadRegistry override may be added by implementer if needed
}

// CollectTTYsFromTree returns real TTYs from ancestors + descendants of rootPID.
// Skips "??", blank, and whitespace-only TTY fields. Order is stable / documented
// by implementer (tests accept any order that includes all real TTYs once).
func CollectTTYsFromTree(procs []ProcRow, rootPID int) []string

// FindITermForSession resolves session -> TTYs -> iTerm refs; does not focus.
func FindITermForSession(opts FocusOpts) ([]FocusCandidate, error)

// FocusSession finds candidates, applies 0/1/multi policy, focuses unless DryRun.
func FocusSession(opts FocusOpts) (FocusCandidate, error)
```

```go
// Package: github.com/xhd2015/agent-pro/pkgs/agentruncli

// Handle dispatches "focus" (unknown command must not occur for "focus").
// RunFocus is the L2 entry with injectable writers (Handle may wire os.Stdout/Stderr).
func RunFocus(args []string, stdout, stderr io.Writer) error

// CLI surface:
//   agent-run focus <session-id> [--index N] [--dry-run] [--session-id ID]
//   agent-run focus -h | --help
```

Depends on P1 (`iterm2.SessionRef`, `FindByTTY`, `Focus`, `NormalizeTTY`). Production
implementer adds go.mod `replace` for local go-pkgs; **designer does not**.

## Version

0.0.2

## Decision Tree

```
pkgs/agentrunapi/tests/iterm-focus/
├── DOCTEST.md
├── SETUP.md
├── collect-ttys/                         [Phase=collect-ttys]
│   └── real-ttys-only/                   ancestors+descendants; skip ?? / blank
├── library/                              FocusSession / FindITermForSession
│   ├── resolve-errors/                   [Phase=focus] store/session_id gates
│   │   ├── missing-session-id/
│   │   └── session-not-in-store/
│   ├── none-found/                       [Phase=focus] zero iTerm candidates
│   │   └── no-match/                     no registry / no real tty / empty list
│   ├── single-match/                     [Phase=focus] one candidate
│   │   ├── focuses/                      DryRun=false → FocusITerm once
│   │   └── dry-run/                      DryRun=true → FocusITerm never
│   └── multi-match/                      [Phase=focus] two+ candidates
│       ├── no-index-errors/              nil Index → error; no focus
│       ├── with-index/                   Index=1 → focus that ref
│       └── index-oob/                    Index out of range → error
└── cli/                                  [Phase=cli] agentruncli.RunFocus
    ├── help/                             -h documents focus, --index, --dry-run
    └── dry-run-dispatches/               focus --dry-run never focuses (inject)
```

Parameter ranking (most → least significant):

1. **Surface** — pure collect-ttys vs library policy vs CLI dispatch
2. **Resolve / match outcome** — error | none | single | multi
3. **Dry-run / index** — focus side effect and multi selection

## Test Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| 1 | `collect-ttys/real-ttys-only` | Tree snapshot: real TTYs only from ancestors+descendants | RED |
| 2 | `library/resolve-errors/missing-session-id` | Empty SessionID → error; no focus | RED |
| 3 | `library/resolve-errors/session-not-in-store` | Unknown session → error; no focus | RED |
| 4 | `library/none-found/no-match` | Session ok but no real TTY / no iTerm match → error | RED |
| 5 | `library/single-match/focuses` | Serve `??` + ancestor ttys148 → one cand; FocusITerm called | RED |
| 6 | `library/single-match/dry-run` | Same single match; DryRun → FocusITerm **not** called | RED |
| 7 | `library/multi-match/no-index-errors` | Two matches, nil Index → error; no focus | RED |
| 8 | `library/multi-match/with-index` | Two matches, Index=1 → focus second | RED |
| 9 | `library/multi-match/index-oob` | Index out of range → error; no focus | RED |
| 10 | `cli/help` | `focus -h` documents focus, `--index`, `--dry-run`; exit 0; trailing `\n` | RED |
| 11 | `cli/dry-run-dispatches` | CLI `--dry-run` succeeds without FocusITerm (library inject via store fixtures) | RED |

## How to Run

```sh
cd .../external/agent-pro-master-2026-07-31
doctest vet ./pkgs/agentrunapi/tests/iterm-focus
doctest test ./pkgs/agentrunapi/tests/iterm-focus
doctest test -v ./pkgs/agentrunapi/tests/iterm-focus/library/single-match/dry-run
```

Expect: **RED** (undefined `FocusSession` / `CollectTTYsFromTree` / `RunFocus` or
build fail until implementer lands P2 + go.mod replace for P1 iterm2 APIs).

```go
import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// Request selects Phase and fixtures for pure collect / library focus / CLI.
type Request struct {
	// Phase: collect-ttys | focus | cli
	Phase string

	// --- shared store / session ---
	Home      string
	SessionID string
	Runner    string // default grok-tty
	TermID    string // terminal_session_id
	// RegistryPID: when >0, leaf/helpers may seed registry JSON under Home.
	// FocusSession production reads registry; tests prefer ListProcs inject and
	// seed registry so resolve finds the root PID.
	RegistryPID int
	SeedSession bool
	SeedRegistry bool

	// --- collect-ttys / inject procs ---
	RootPID int
	Procs   []agentrunapi.ProcRow

	// --- focus policy ---
	DryRun     bool
	Index      *int
	ITermRefs  []iterm2.SessionRef
	// ListITermErr, when non-empty, makes ListITerm return that error string.
	ListITermErr string

	// --- cli ---
	CLIArgs []string
}

// Response holds pure / library / CLI results.
type Response struct {
	TTYs       []string
	Candidates []agentrunapi.FocusCandidate
	Chosen     agentrunapi.FocusCandidate
	FocusCalls []iterm2.SessionRef
	Stdout     string
	Stderr     string
}

// Run dispatches on req.Phase and calls product APIs (Classic TDD — missing
// symbols fail RED until implementer lands them).
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	resp := &Response{}

	switch req.Phase {
	case "collect-ttys":
		resp.TTYs = agentrunapi.CollectTTYsFromTree(req.Procs, req.RootPID)
		return resp, nil

	case "focus":
		opts, focusLog, err := buildFocusOpts(t, req)
		if err != nil {
			return resp, err
		}
		chosen, ferr := agentrunapi.FocusSession(opts)
		resp.Chosen = chosen
		resp.FocusCalls = append([]iterm2.SessionRef(nil), focusLog.calls...)
		// Also expose Find path for multi-list asserts when useful.
		if cands, findErr := agentrunapi.FindITermForSession(opts); findErr == nil {
			resp.Candidates = cands
		}
		return resp, ferr

	case "cli":
		var stdout, stderr bytes.Buffer
		err := agentruncli.RunFocus(req.CLIArgs, &stdout, &stderr)
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		return resp, err

	default:
		return nil, fmt.Errorf("unknown Phase %q", req.Phase)
	}
}

// focusCallLog records FocusITerm invocations for asserts.
type focusCallLog struct {
	calls []iterm2.SessionRef
}

func buildFocusOpts(t *testing.T, req *Request) (agentrunapi.FocusOpts, *focusCallLog, error) {
	t.Helper()
	log := &focusCallLog{}

	var store agentstorage.Store
	if req.Home != "" {
		s, err := agentstorage.NewFileStore(req.Home)
		if err != nil {
			return agentrunapi.FocusOpts{}, log, err
		}
		store = s
		if req.SeedSession {
			if err := seedSession(t, store, req); err != nil {
				return agentrunapi.FocusOpts{}, log, err
			}
		}
		if req.SeedRegistry {
			if err := seedRegistry(t, req); err != nil {
				return agentrunapi.FocusOpts{}, log, err
			}
		}
	}

	procs := req.Procs
	opts := agentrunapi.FocusOpts{
		Store:     store,
		SessionID: req.SessionID,
		Index:     req.Index,
		DryRun:    req.DryRun,
		ListProcs: func() []agentrunapi.ProcRow {
			return procs
		},
		ListITerm: func() ([]iterm2.SessionRef, error) {
			if req.ListITermErr != "" {
				return nil, fmt.Errorf("%s", req.ListITermErr)
			}
			return append([]iterm2.SessionRef(nil), req.ITermRefs...), nil
		},
		FocusITerm: func(ref iterm2.SessionRef) error {
			log.calls = append(log.calls, ref)
			return nil
		},
	}
	return opts, log, nil
}

func seedSession(t *testing.T, store agentstorage.Store, req *Request) error {
	t.Helper()
	runner := req.Runner
	if runner == "" {
		runner = "grok-tty"
	}
	meta := agentstorage.SessionMeta{
		Runner:            runner,
		SessionID:         req.SessionID,
		TerminalSessionID: req.TermID,
		Status:            "running",
	}
	return store.CreateSession(req.SessionID, meta)
}

func seedRegistry(t *testing.T, req *Request) error {
	t.Helper()
	runner := req.Runner
	if runner == "" {
		runner = "grok-tty"
	}
	termID := req.TermID
	if termID == "" {
		return fmt.Errorf("seedRegistry: TermID required")
	}
	pid := req.RegistryPID
	if pid == 0 {
		pid = req.RootPID
	}
	dir := filepath.Join(req.Home, runner+"-registry")
	return writeRegistryJSON(dir, termID, pid)
}
```
