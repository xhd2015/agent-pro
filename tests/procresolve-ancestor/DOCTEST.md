# procresolve-ancestor — nearest grok on the PPID chain

Classic TDD doctests for ancestor walk in
`github.com/xhd2015/agent-pro/pkgs/procresolve`. **RED** until the implementer
lands `FindAncestorGrok` and `ResolveFromAncestors`.

Existing `ResolveFromPID` walks **descendants only**. This tree must stay RED
against that API: `Run` calls the new functions; hit leaves start below the
grok; none leaves plant a **descendant decoy grok** so a descendant-only walk
would incorrectly hit.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — later `grok-fork` / `fork` library: has a pid (self or `--pid`)
  and needs the nearest ancestor grok plus its session id.
- **`FindAncestorGrok`** — `FindAncestorGrok(startPID int, opts Options) (Proc, bool)`.
  Walks **startPID then PPID** using `ListProcs`. First `IsGrokRunner(cmd)` wins
  (nearest). Skips `grok update`. Missing start PID → `ok=false`, zero `Proc`
  (no error; signature has no error).
- **`ResolveFromAncestors`** — `ResolveFromAncestors(startPID int, opts Options) (*Result, error)`.
  Finds the nearest grok ancestor, then `ResolveFromPID(thatGrokPID, opts)` so
  the session id still comes from open files (`Lsof` + `ParseSessionFromPath`),
  never cmdline `--resume` / `--session-id`.
- **Process snapshot / Lsof** — same injectable `Options.ListProcs` / `Options.Lsof`
  as `ResolveFromPID`. Tests never list live processes.
- **Classifier** — reuse `IsGrokRunner` (basename `grok`, exclude `grok update`).
  Codex is not a grok hit.

**Behaviors**

- start pid present, grok on PPID chain (including start itself) →
  `FindAncestorGrok` ok, proc is that grok; `ResolveFromAncestors` hard-hit
  from that grok’s open files (`Kind=grok`, `Confidence=hard`,
  `SessionID` from path). `Result.InputPID` / `RunnerPID` are the ancestor
  grok pid (because resolve delegates to `ResolveFromPID(grokPID)`).
- start pid present, no grok ancestor → `FindAncestorGrok` ok=false;
  `ResolveFromAncestors` returns `Kind=none`, `SessionID=""`, no error
  (mirror `ResolveFromPID` miss). Do **not** fall back to descendant groks.
- start pid absent from snapshot → `FindAncestorGrok` ok=false;
  `ResolveFromAncestors` error contains `pid not found`.
- Nested groks: nearest (first walking start then PPID), not topmost.
- `grok update` on the chain is skipped; a real grok further up still wins.
- Codex-only ancestor is not a grok hit.

## Locked contract

```text
FindAncestorGrok(startPID int, opts Options) (proc Proc, ok bool)
  missing start PID → ok=false, zero Proc
  walk start, then PPID, stop at missing parent / pid 0
  first IsGrokRunner(cmd) wins
  skip grok update

ResolveFromAncestors(startPID int, opts Options) (*Result, error)
  start PID absent → error "pid not found"
  no grok ancestor → Kind=none, SessionID="", Confidence="", no error
  else ResolveFromPID(ancestorGrok.PID, opts)
  session id from Lsof paths only
```

## Version

0.0.2

## Decision Tree

```
tests/procresolve-ancestor/
├── DOCTEST.md
├── SETUP.md
├── resolve-hit/                         # grok ancestor + hard session from Lsof
│   ├── from-descendant/                 # grok → bash → start; --resume on cmd ignored
│   ├── from-grok-self/                  # start pid is the grok
│   ├── nested-nearest/                  # main → subagent grok → bash → start
│   ├── skip-update-on-chain/            # real grok above grok update
│   └── start-at-update/                 # start is grok update; parent is real grok
├── resolve-none/                        # no grok ancestor (decoy descendant grok ignored)
│   ├── no-grok-on-chain/
│   ├── codex-ancestor-only/
│   └── update-only/                     # only grok update above start
└── resolve-error/
    └── unknown-pid/                     # start pid missing → pid not found
```

Parameter ranking (most → least significant):

1. **Outcome** — hard hit vs none vs error
2. **Where the walk starts / what sits on the PPID chain** — descendant, self,
   nested, update skip, non-grok ancestors

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `resolve-hit/from-descendant` | grok→bash→start; session from grok Lsof; cmdline `--resume` ignored |
| 2 | `resolve-hit/from-grok-self` | start at grok pid; grok is the ancestor; session from its open files |
| 3 | `resolve-hit/nested-nearest` | main grok + nearer subagent grok; nearest (subagent) wins |
| 4 | `resolve-hit/skip-update-on-chain` | `grok update` between start and real grok; skip update |
| 5 | `resolve-hit/start-at-update` | start pid is `grok update`; parent real grok wins |
| 6 | `resolve-none/no-grok-on-chain` | bash chain + descendant decoy grok → none |
| 7 | `resolve-none/codex-ancestor-only` | codex ancestor + descendant decoy grok → none |
| 8 | `resolve-none/update-only` | only `grok update` above start + decoy descendant → none |
| 9 | `resolve-error/unknown-pid` | pid absent → error contains `pid not found` |

## How to Run

```sh
doctest vet ./tests/procresolve-ancestor
doctest test ./tests/procresolve-ancestor

doctest test -v ./tests/procresolve-ancestor/resolve-hit/from-descendant
doctest test -v ./tests/procresolve-ancestor/resolve-hit/from-grok-self
doctest test -v ./tests/procresolve-ancestor/resolve-hit/nested-nearest
doctest test -v ./tests/procresolve-ancestor/resolve-hit/skip-update-on-chain
doctest test -v ./tests/procresolve-ancestor/resolve-hit/start-at-update
doctest test -v ./tests/procresolve-ancestor/resolve-none/no-grok-on-chain
doctest test -v ./tests/procresolve-ancestor/resolve-none/codex-ancestor-only
doctest test -v ./tests/procresolve-ancestor/resolve-none/update-only
doctest test -v ./tests/procresolve-ancestor/resolve-error/unknown-pid
```

Classic TDD: compile-fail or assert-fail until `FindAncestorGrok` and
`ResolveFromAncestors` exist.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

// FixtureProc is one process row in the injectable snapshot.
type FixtureProc struct {
	PID  int
	PPID int
	Cmd  string
}

// Request is filled root→leaf: start PID + Procs + OpenFiles.
type Request struct {
	PID       int
	Procs     []FixtureProc
	OpenFiles map[int][]string
	MaxDepth  int
	GrokHome  string
	CodexHome string
}

// Response observes both new APIs on the same fixture.
type Response struct {
	Ancestor   procresolve.Proc
	AncestorOK bool
	Result     *procresolve.Result
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

	anc, ok := procresolve.FindAncestorGrok(req.PID, opts)
	result, err := procresolve.ResolveFromAncestors(req.PID, opts)
	return &Response{
		Ancestor:   anc,
		AncestorOK: ok,
		Result:     result,
	}, err
}
```
