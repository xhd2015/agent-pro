# pkgs/itermsnapshot — agent-aware enrich on P1 snapshot

Classic TDD doctests for plan phase **P2**: package
`github.com/xhd2015/agent-pro/pkgs/itermsnapshot` — **RED** until the
implementer adds the package. Wraps P1
`shell/iterm2/snapshot` Capture + `procresolve` agent attach for **busy**
panes only. **No kool import.** Does not re-implement hierarchy/process enrich.

**Out of scope:** P3 kool rewire, P4 kck list, save/restore, CLI render.

## Version

0.0.2

# DSN (Domain Specific Notion)

Library package that takes a P1 iTerm2 inventory and attaches grok/codex
session agents onto busy panes via procresolve — composition, not model
mutation of `snapshot.SnapshotSession`.

### Participants

- **Caller** — library client that invokes **Capture** with **CaptureOpts**
  (production or L2 inject path).
- **Base inventory** — either inject **CaptureOpts.Snapshot** (pre-built
  `*snapshot.Snapshot`) or **BaseCapture** / default `snapshot.Capture` for
  live hierarchy + process enrich (P1).
- **Resolver** — **CaptureOpts.ResolveFromPID** (or production default wired
  to `procresolve.ResolveFromPID` with live opts). Soft on error / none.
- **Result** — **Result.Snapshot** (base inventory) + **Result.Agents** map
  keyed by iTerm session **ID** → **SessionAgent** (Kind, SessionID, Title,
  Tree).

### Behaviors

- **NoEnrich** — skip all agent attach; Agents empty even if busy panes would
  hit.
- **Busy-only** — attach only when `Idle != nil && !*Idle` (kool parity);
  idle and unknown (`Idle==nil`) never attach.
- **PID pick** — prefer `session.PID`, else `session.ShellPID`; if neither
  positive → no attach.
- **Resolve hard hit** — Kind grok|codex with non-empty SessionID → Agent;
  Title from `Result.GrokTitle`; Tree mapped from `Result.Tree`.
- **Resolve soft miss** — error, nil result, Kind empty/none, or empty
  SessionID → no agent for that session (does not fail Capture).
- **Empty inventory** — zero windows → empty Agents; still success.
- **Multi busy** — each busy session resolved independently; keys are
  session IDs.

## Decision Tree

```text
itermsnapshot/tests/
├── no-enrich/
│   └── busy-would-hit/          # NoEnrich + busy + resolve hit → Agents empty
├── empty-snapshot/
│   └── zero-windows/            # empty Snapshot → empty Agents
└── enrich/                      # NoEnrich=false
    ├── idle-session/            # Idle=true → no agent
    ├── unknown-session/         # Idle=nil → no agent
    ├── busy/
    │   ├── resolve-grok/        # hard grok hit → Agent attached
    │   ├── resolve-codex/       # hard codex hit → Agent attached
    │   ├── resolve-none/        # Kind=none → no agent
    │   ├── resolve-error/       # resolve error → soft no agent
    │   └── shell-pid-fallback/  # PID nil, ShellPID set → uses ShellPID
    └── multi-busy/              # two busy sessions, independent agents
```

Parameter ranking (most → least significant):

1. **NoEnrich** — global skip vs enrich path
2. **Inventory shape** — empty vs sessions present
3. **Session Idle class** — idle / unknown / busy
4. **Resolve outcome** (busy only) — hard hit kind / none / error / PID source
5. **Multiplicity** — single vs multi busy sessions

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `no-enrich/busy-would-hit` | NoEnrich skips attach despite busy + resolve hit |
| 2 | `empty-snapshot/zero-windows` | empty Snapshot → empty Agents, success |
| 3 | `enrich/idle-session` | Idle=true → no agent (resolve unused) |
| 4 | `enrich/unknown-session` | Idle=nil → no agent |
| 5 | `enrich/busy/resolve-grok` | busy + grok hard hit → Agent Kind/SessionID/Title/Tree |
| 6 | `enrich/busy/resolve-codex` | busy + codex hard hit → Agent |
| 7 | `enrich/busy/resolve-none` | busy + Kind=none → no agent |
| 8 | `enrich/busy/resolve-error` | busy + resolve error → soft no agent |
| 9 | `enrich/busy/shell-pid-fallback` | PID nil, ShellPID set → resolve called with ShellPID |
| 10 | `enrich/multi-busy` | two busy sessions → independent Agents by session ID |

## How to Run

```sh
# from agent-pro module root
cd /Users/xhd2015/Projects/xhd2015/kck/external/agent-pro-master-2026-08-06-1

# implementer may need replace to local P1 snapshot:
#   replace github.com/xhd2015/dot-pkgs/go-pkgs => ../dot-pkgs-master-2026-08-06-1/go-pkgs

doctest vet ./pkgs/itermsnapshot/tests
doctest test ./pkgs/itermsnapshot/tests

doctest test -v ./pkgs/itermsnapshot/tests/no-enrich/busy-would-hit
doctest test -v ./pkgs/itermsnapshot/tests/enrich/busy/resolve-grok
doctest test -v ./pkgs/itermsnapshot/tests/enrich/multi-busy
```

Classic TDD: expect **RED** (compile or assert failure) until
`pkgs/itermsnapshot` production package exists and implements the locked API
below.

## Locked public API (implementer contract)

Import path: `github.com/xhd2015/agent-pro/pkgs/itermsnapshot`

Dependencies (no kool):

- `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot` (P1)
- `github.com/xhd2015/agent-pro/pkgs/procresolve`

```go
package itermsnapshot

import (
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// CaptureOpts controls agent-enrich Capture. L2 injects Snapshot + ResolveFromPID.
type CaptureOpts struct {
	// NoEnrich skips procresolve agent attach entirely.
	NoEnrich bool

	// Snapshot, when non-nil, is the base inventory; BaseCapture is not called.
	// L2 tests inject a fixture *snapshot.Snapshot (no AppleScript / live ps).
	Snapshot *snapshot.Snapshot

	// BaseCapture obtains the base inventory when Snapshot is nil.
	// Production default: snapshot.Capture (or equivalent via snapshot.Collector).
	// Hard errors from BaseCapture propagate from Capture.
	BaseCapture func() (*snapshot.Snapshot, []string, error)

	// ResolveFromPID attaches agents for busy panes.
	// Production default: procresolve.ResolveFromPID with live ListProcs/Lsof
	// and EnrichInfo=true (kool parity). Soft on error/nil/none.
	ResolveFromPID func(pid int) (*procresolve.Result, error)
}

// Result is the enriched view: base Snapshot plus agents by session ID.
type Result struct {
	Snapshot *snapshot.Snapshot
	// Agents keyed by SnapshotSession.ID (iTerm session UUID).
	// Nil or empty when no agents attached.
	Agents map[string]*SessionAgent
}

// SessionAgent is the procresolve-derived agent for one busy pane.
type SessionAgent struct {
	Kind      string // grok | codex
	SessionID string
	Title     string // from procresolve.Result.GrokTitle when present
	Tree      []AgentTreeNode
}

// AgentTreeNode is one process in the agent process tree.
type AgentTreeNode struct {
	PID  int
	PPID int
	Role string // input | agent-run | … | grok | codex | other
	Cmd  string
}

// Capture runs base inventory (inject Snapshot or BaseCapture / default), then
// attaches agents for busy panes unless NoEnrich.
// Returns (*Result, warnings, error). Hard error only from base capture.
// Agent resolve failures are soft (no entry in Agents).
func Capture(opts CaptureOpts) (*Result, []string, error)
```

### Attach semantics (kool `attachAgent` parity)

For each session in `Result.Snapshot` windows/tabs/sessions, when not NoEnrich:

1. Skip if `Idle == nil` OR `*Idle == true` (busy-only: Idle non-nil and false).
2. `pid = *PID` if PID non-nil; else `*ShellPID` if ShellPID non-nil; else skip.
3. Skip if `pid <= 0`.
4. Call `ResolveFromPID(pid)`:
   - error or nil result → no agent for session
   - Kind `""` or `"none"` or SessionID `""` → no agent
   - else set `Agents[session.ID] = &SessionAgent{Kind, SessionID, Title: GrokTitle, Tree}`
5. Tree mapping: each `procresolve.ProcNode` → `AgentTreeNode` (PID, PPID, Role, Cmd).

Does **not** mutate `snapshot.SnapshotSession` (P1 model has no Agent field).

### Production defaults (implementer)

When `Snapshot == nil` and `BaseCapture == nil`: use `snapshot.Capture`.
When `ResolveFromPID == nil` and not NoEnrich: default to
`procresolve.ResolveFromPID(pid, Options{ListProcs: ListLiveProcs, Lsof: LiveLsof, EnrichInfo: true})`.

### L2 testability

Leaves inject `CaptureOpts.Snapshot` + `CaptureOpts.ResolveFromPID` only.
No process globals, no env, no AppleScript. Parallel-safe: pure per-call opts.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/itermsnapshot"
	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

// Request is filled root→leaf. Each leaf owns inject data via Setup chain.
type Request struct {
	// NoEnrich maps to CaptureOpts.NoEnrich.
	NoEnrich bool

	// Snapshot is the injected base inventory (always set by leaves for L2).
	Snapshot *snapshot.Snapshot

	// ResolveFromPID maps to CaptureOpts.ResolveFromPID (may be nil when
	// NoEnrich or when assert only cares that resolve is unused).
	ResolveFromPID func(pid int) (*procresolve.Result, error)

	// ResolveCallPIDs records PIDs passed to ResolveFromPID (leaf Setup wraps).
	// Optional; used by shell-pid-fallback / multi-busy when needed.
	ResolveCallPIDs *[]int
}

// Response observes Capture outcomes.
type Response struct {
	Result   *itermsnapshot.Result
	Warnings []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	res, warn, err := itermsnapshot.Capture(itermsnapshot.CaptureOpts{
		NoEnrich:       req.NoEnrich,
		Snapshot:       req.Snapshot,
		ResolveFromPID: req.ResolveFromPID,
	})
	return &Response{Result: res, Warnings: warn}, err
}

// --- assert helpers used by leaves ---

func mustResult(t *testing.T, resp *Response, err error) *itermsnapshot.Result {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Result == nil {
		t.Fatal("expected non-nil Result")
	}
	if resp.Result.Snapshot == nil {
		t.Fatal("expected non-nil Result.Snapshot")
	}
	return resp.Result
}

func agentCount(agents map[string]*itermsnapshot.SessionAgent) int {
	return len(agents)
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
func strPtr(v string) *string { return &v }

// oneBusySession builds a one-window snapshot with a single busy session.
func oneBusySession(sessID, name, tty string, pid int) *snapshot.Snapshot {
	return oneSessionSnap(sessID, name, tty, boolPtr(false), intPtr(pid), nil)
}

// oneSessionSnap builds a minimal one-window / one-tab / one-session Snapshot.
func oneSessionSnap(sessID, name, tty string, idle *bool, pid, shellPID *int) *snapshot.Snapshot {
	return &snapshot.Snapshot{
		CapturedAt: "2026-07-25T12:00:00Z",
		Host:       "testhost",
		Source:     "iterm2",
		Summary:    snapshot.SnapshotSummary{Windows: 1, Tabs: 1, Sessions: 1},
		Windows: []snapshot.SnapshotWindow{
			{
				Index: 1,
				Name:  "W1",
				Tabs: []snapshot.SnapshotTab{
					{
						Index: 1,
						Name:  "T1",
						Sessions: []snapshot.SnapshotSession{
							{
								Index:    1,
								ID:       sessID,
								Name:     name,
								TTY:      tty,
								Profile:  "Default",
								Idle:     idle,
								PID:      pid,
								ShellPID: shellPID,
							},
						},
					},
				},
			},
		},
	}
}

// resolveHit returns a ResolveFromPID that always yields the given hard hit
// for any positive pid (used by single-session leaves).
func resolveHit(kind, sessionID, title string, tree []procresolve.ProcNode) func(int) (*procresolve.Result, error) {
	return func(pid int) (*procresolve.Result, error) {
		if pid <= 0 {
			return nil, fmt.Errorf("unexpected pid %d", pid)
		}
		return &procresolve.Result{
			InputPID:   pid,
			Kind:       kind,
			SessionID:  sessionID,
			Source:     "open-files",
			Confidence: "hard",
			GrokTitle:  title,
			Tree:       tree,
		}, nil
	}
}

// resolveByPID maps input pid → result; missing key → Kind none.
func resolveByPID(m map[int]*procresolve.Result) func(int) (*procresolve.Result, error) {
	return func(pid int) (*procresolve.Result, error) {
		if r, ok := m[pid]; ok {
			return r, nil
		}
		return &procresolve.Result{InputPID: pid, Kind: "none"}, nil
	}
}
```
