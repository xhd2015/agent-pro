# procresolve — resolve session id from pid via process tree + open files

Classic TDD doctests for plan phase **P1**: package
`github.com/xhd2015/agent-pro/pkgs/procresolve` — **GREEN**.

Given an **input pid**, walk the process tree (injectable snapshot), find leaf
**grok** / **codex** runner binaries, and resolve **session ids** from open
files (`lsof` inject) — **not** from cmdline flags.

**P2 library** (FormatTree, EnrichInfo) lives in the nested root
`./tests/procresolve/p2/` (own `DOCTEST.md` / `Run`). **P2 CLI** lives in
`./tests/proc-resolve-cli/`.

**Out of scope (this root):** iTerm / kool, `agent-pro proc resolve` CLI,
Unicode tree formatter, `groksessions.Info` title enrichment, cmdline
`--session-id` / `--resume` as primary resolve, cwd→newest session heuristics.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — later CLI or in-process client that has a pid (agent-run, serve,
  shell, or the runner itself) and needs the active session id.
- **`ResolveFromPID`** — package entrypoint
  `ResolveFromPID(pid int, opts Options) (*Result, error)`. Builds a descendant
  tree from the input pid, classifies nodes, picks grok/codex candidates
  (prefer deeper leaves), runs `Lsof` on each until a hard session path hit.
- **Process snapshot (`ListProcs`)** — injectable `func() []Proc`. Each `Proc`
  has `PID`, `PPID`, and `Cmd` (path + argv as one string for classification).
  Production will list live processes; tests always inject fixtures.
- **Open-files probe (`Lsof`)** — injectable `func(pid int) []string` returning
  absolute open-file paths for one pid. Production wraps `lsof`; tests never
  shell out.
- **Classifier** — maps each node’s cmd to a role:
  - `input` — the requested input pid (always present when pid is found)
  - `agent-run` — argv0/path is agent-run and not the serve subcommand
  - `agent-run-serve` — agent-run with `serve`
  - `grok` — grok runner; **exclude** pure `grok update` (not a session runner)
  - `codex` — codex runner
  - `other` — everything else (bash, node wrappers, …)
- **Session path parser** — from open-file paths only:
  - **Grok hard hit:** primary artifact under `…/.grok/sessions/…/<uuid>/`
    whose basename is `events.jsonl` or `updates.jsonl` (uuid often `019f…`).
    Bare session-directory opens (startup history scans) are **not** hard hits.
  - **Codex:** path under `…/.codex/sessions/…/rollout-…-<uuid>` or
    `…/.codex/thread-writer-locks/<uuid>.lock`
- **Result** — Kind (`grok`|`codex`|`none`), SessionID, Source
  (`open-files` when input itself is the runner; `open-files+tree` when a
  descendant runner was used), Confidence (`hard` on hit, empty when none),
  RunnerPID / RunnerCmd, full `Tree []ProcNode`, Warnings.

**Behaviors**

```
# resolve pipeline
input pid
  -> ListProcs snapshot
  -> build tree: input + descendants (MaxDepth)
  -> classify roles (exclude "grok update" as runner)
  -> candidates = grok|codex nodes; prefer deeper leaves
  -> for each candidate: Lsof -> parse session uuid from path
  -> grok: skip non-primary opens (require events.jsonl|updates.jsonl)
  -> first hard hit wins
  -> if none: Kind=none, SessionID="", Confidence="", no error
  -> if pid absent from snapshot: error ("pid not found")

# must NOT
  parse cmdline --session-id / --resume as primary session source
  treat grok bare …/sessions/…/<uuid> directory opens as hard hits
```

**Locked types (implementer contract)**

```text
Options
  GrokHome, CodexHome string
  MaxDepth int
  ListProcs func() []Proc
  Lsof     func(pid int) []string

Proc
  PID, PPID int
  Cmd string

ProcNode
  PID, PPID int
  Role string  // input | agent-run | agent-run-serve | grok | codex | other
  Cmd string

Result
  InputPID int
  Kind string       // grok | codex | none
  SessionID string
  Source string     // open-files | open-files+tree
  Confidence string // hard | "" when none
  RunnerPID int
  RunnerCmd string
  Tree []ProcNode
  Warnings []string
```

## Version

0.0.2

## Decision Tree

```
tests/procresolve/
├── DOCTEST.md
├── SETUP.md
├── resolve-hit/                         # hard session id found
│   ├── SETUP.md
│   ├── grok/
│   │   ├── SETUP.md
│   │   ├── bare-input/                  # input pid is grok; Source open-files
│   │   └── agent-run-tree/              # agent-run → serve → grok; open-files+tree
│   └── codex/
│       ├── SETUP.md
│       └── multi-node/                  # agent-run → serve → node → codex leaf
├── resolve-none/                        # Kind=none, no error
│   ├── SETUP.md
│   ├── plain-bash/                      # bash pid, empty Lsof
│   └── exclude-grok-update/             # only "grok update" (not a runner)
├── resolve-error/                       # hard failure
│   ├── SETUP.md
│   └── unknown-pid/                     # pid missing from ListProcs → error
└── p2/                                  # nested DOCTEST root (P2 library; see p2/DOCTEST.md)
    ├── format-tree/
    └── enrich-info/
```

Parameter ranking (most → least significant):

1. **Outcome class** — hard hit vs none vs error (`resolve-hit` / `resolve-none` / `resolve-error`)
2. **Runner kind** (hits only) — grok vs codex
3. **Topology** — bare input runner vs multi-level agent-run tree vs excluded update

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `resolve-hit/grok/bare-input` | Input pid is grok; Lsof has grok session uuid → Kind=grok, Confidence=hard, Source open-files, RunnerPID=input |
| 2 | `resolve-hit/grok/agent-run-tree` | agent-run → serve → grok; session from grandchild; Tree roles; Source open-files+tree |
| 3 | `resolve-hit/codex/multi-node` | agent-run → serve → node (no files) → codex leaf with rollout path → Kind=codex |
| 4 | `resolve-none/plain-bash` | Plain bash, empty Lsof → Kind=none, empty SessionID, no error |
| 5 | `resolve-none/exclude-grok-update` | Only `grok update` in tree → Kind=none (not a session runner) |
| 6 | `resolve-error/unknown-pid` | Pid not in ListProcs → error mentioning `pid not found` |

P2 library leaves: see `./p2/DOCTEST.md`. CLI: `./tests/proc-resolve-cli/`.

## How to Run

```sh
doctest vet ./tests/procresolve
doctest test ./tests/procresolve

doctest test -v ./tests/procresolve/resolve-hit/grok/bare-input
doctest test -v ./tests/procresolve/resolve-hit/grok/agent-run-tree
doctest test -v ./tests/procresolve/resolve-hit/codex/multi-node
doctest test -v ./tests/procresolve/resolve-none/plain-bash
doctest test -v ./tests/procresolve/resolve-none/exclude-grok-update
doctest test -v ./tests/procresolve/resolve-error/unknown-pid

# P2 library (nested root)
doctest vet ./tests/procresolve/p2
doctest test ./tests/procresolve/p2
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

// FixtureProc is one process row in the injectable snapshot.
// Mapped 1:1 to procresolve.Proc in Run.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

// Request is filled root→leaf: PID + Procs + OpenFiles (+ optional homes/depth).
type Request struct {
	PID       int
	Procs     []FixtureProc
	OpenFiles map[int][]string // pid → absolute open paths from fake Lsof
	MaxDepth  int
	GrokHome  string
	CodexHome string
}

// Response observes ResolveFromPID output.
type Response struct {
	Result *procresolve.Result
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	procs := make([]procresolve.Proc, 0, len(req.Procs))
	for _, p := range req.Procs {
		procs = append(procs, procresolve.Proc{
			PID:  p.PID,
			PPID: p.PPID,
			Cmd:  p.Cmd,
		})
	}
	// Snapshot for injectors (stable across calls within one Resolve).
	snap := procs
	files := req.OpenFiles
	if files == nil {
		files = map[int][]string{}
	}

	opts := procresolve.Options{
		GrokHome:  req.GrokHome,
		CodexHome: req.CodexHome,
		MaxDepth:  req.MaxDepth,
		ListProcs: func() []procresolve.Proc {
			return snap
		},
		Lsof: func(pid int) []string {
			return files[pid]
		},
	}

	result, err := procresolve.ResolveFromPID(req.PID, opts)
	return &Response{Result: result}, err
}
```
