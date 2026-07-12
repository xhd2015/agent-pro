# agent-run pty Subcommand Tests

Doc-style tests for `agent-run pty` — PTY resource stats, kill of orphan
`__serve*` processes (with PPID / `--kind` / `--all` filters), and harness
verification that `run --open` leaves no serve after reclaim cleanup.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — dispatches `pty stats`, `pty kill-orphans`, and top-level
  help that mentions `pty`. Also re-execs itself as a detached `__serve_<slug>__`
  child when TTY runners keep a session alive (`run --open` / `--keep-tty`).
- **PTY / process table** — runtime-only view of `kern.tty.ptmx_max` (when
  available), unique `/dev/ptmx` masters, and live processes whose argv matches
  the serve token (`IsServeSubcommand`) or an `agent-run __serve…` command line.
- **Serve process** — detached re-exec child holding a PTY + optional registry
  entry under `AGENT_RUN_HOME/<runner>-registry/<session-id>.json` (pid,
  listen_addr). Kill-orphans **selection** is filtered (see Behaviors); tests
  isolate with `--exe PATH` so host brainstorm/seatalk sessions are not destroyed.
- **Harness reclaim** — `t.Cleanup` / `ReclaimSessionID` (or equivalent kill of
  serve PID + registry remove) used by TestGenerated open paths so keep-alive
  serves do not leak after the case ends. Product KeepAlive semantics for real
  users are unchanged.

**Behaviors**

- `agent-run pty --help` lists `stats` and `kill-orphans` (and documents
  kill-orphans flags on subcommand help).
- `agent-run --help` mentions `pty` alongside existing commands.
- Unknown `pty` subcommand → exit 1, stderr error.
- `agent-run pty stats` prints human-readable PTY limit / masters / free
  estimate (best-effort) and agent-run `__serve` counts/categories; exit 0 on
  success (partial probes OK). Successful stdout ends with trailing `\n`.
- `agent-run pty kill-orphans [--dry-run] [--exe PATH] [--all]
  [--kind=test-generated] [--kind=workdir-at-tmp]`:
  - still required for any match: looks like agent-run serve, live non-zombie,
    never kill self PID
  - **default** (no `--kind`, no `--all`): only serves with **PPID == 1**
  - **`--kind=test-generated`**: argv/exe contain `TestGenerated` (any PPID);
    replaces PPID filter
  - **`--kind=workdir-at-tmp`**: argv/exe contain `/var/folders/` **and** `/T/`
    (macOS Go temp layout; any PPID); replaces PPID filter
  - **multiple `--kind`**: OR of kind predicates (no PPID filter)
  - **`--all`**: every agent-run `__serve*`; **wins over** `--kind` if both set
  - **`--exe PATH`**: always **ANDed** when set
  - **`--dry-run`**: same selection as kill; print only; empty set → clear
    no-match / no-orphans line; exit 0
  - unknown `--kind` value → exit 1, stderr error
  - without `--dry-run`, terminate targets (SIGTERM then SIGKILL); exit 0 when
    terminations were attempted (exit 1 on hard kill failure)
  - successful stdout ends with trailing `\n`
  - kill-orphans help documents `--dry-run`, `--exe`, `--kind`, `--all`
- Open-cleanup harness: after `run --open` with a mock TUI, reclaim must leave
  no live `__serve` for that test session/home.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/pty/
├── DOCTEST.md
├── SETUP.md                              # build agent-run; spawn/kill/liveness helpers
├── help/                                 # discovery surfaces
│   ├── lists-subcommands/                # pty --help lists stats, kill-orphans
│   ├── top-level-mentions-pty/           # agent-run --help mentions pty
│   └── unknown-subcommand/               # pty not-a-cmd → exit 1
├── stats/
│   └── prints-summary/                   # exit 0; limit/serve keywords + trailing \n
├── kill-orphans/                         # always use --exe in live tests when possible
│   ├── help-documents-flags/             # --help: --dry-run, --exe
│   ├── dry-run-lists-serve/              # PPID1 orphan; default dry-run --exe lists PID
│   ├── dry-run-none/                     # --exe unique path, no serves → no-match, exit 0
│   ├── kills-matching-exe/               # PPID1 orphan; kill-orphans --exe; process gone
│   └── filter/                           # selection: default PPID1 / --kind / --all
│       ├── default-ppid1-only/           # PPID1 listed; non-orphan child not listed
│       ├── kind-test-generated/          # --kind=test-generated lists PPID≠1 TG serve
│       ├── kind-workdir-at-tmp/          # --kind=workdir-at-tmp lists PPID≠1 temp-path serve
│       ├── kind-multi-or/                # two kinds OR → both matching serves listed
│       ├── all-includes-non-ppid1/       # --all lists PPID≠1; default does not
│       ├── all-wins-over-kind/           # --all + non-matching kind still lists serve
│       ├── help-documents-kind-and-all/  # --help documents --kind and --all
│       └── unknown-kind-rejected/        # unknown --kind → exit 1, stderr
└── open-cleanup/
    └── open-reclaims-serve/              # run --open + harness reclaim → serve gone
```

Parameter ranking (most → least significant):

1. **Operation surface** — help vs stats vs kill-orphans vs open-cleanup
2. **Help surface** — `pty --help` vs top-level vs unknown subcommand
3. **Kill mode** — dry-run vs real kill vs kill-orphans help
4. **Selection filter** — default PPID1 vs `--kind` vs `--all` vs unknown kind
5. **Target set / markers** — `--exe`, TestGenerated path/argv, workdir-at-tmp path

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/lists-subcommands` | `pty --help` lists `stats` and `kill-orphans`; exit 0; trailing `\n` |
| 2 | `help/top-level-mentions-pty` | `agent-run --help` mentions `pty`; exit 0 |
| 3 | `help/unknown-subcommand` | `pty not-a-real-cmd` → exit 1, stderr error |
| 4 | `stats/prints-summary` | `pty stats` exit 0; stdout has limit/serve-related text + trailing `\n` |
| 5 | `kill-orphans/help-documents-flags` | `pty kill-orphans --help` documents `--dry-run` and `--exe` |
| 6 | `kill-orphans/dry-run-lists-serve` | Spawn PPID1 test `__serve`; default dry-run `--exe` lists PID; still alive |
| 7 | `kill-orphans/dry-run-none` | dry-run `--exe` unique path with no serves → no-match; exit 0; trailing `\n` |
| 8 | `kill-orphans/kills-matching-exe` | Spawn PPID1 serve; `kill-orphans --exe` → process gone; exit 0 |
| 9 | `kill-orphans/filter/default-ppid1-only` | PPID1 + non-orphan child same `--exe`; default dry-run lists only PPID1 |
| 10 | `kill-orphans/filter/kind-test-generated` | PPID≠1 serve with `TestGenerated` marker; kind lists it; default does not |
| 11 | `kill-orphans/filter/kind-workdir-at-tmp` | PPID≠1 serve under `/var/folders/…/T/`; kind lists; default does not |
| 12 | `kill-orphans/filter/kind-multi-or` | TG + workdir markers; both kinds on CLI → both PIDs listed |
| 13 | `kill-orphans/filter/all-includes-non-ppid1` | PPID≠1 serve; `--all` dry-run lists; default does not |
| 14 | `kill-orphans/filter/all-wins-over-kind` | PPID≠1 without TG marker; `--all --kind=test-generated` still lists it |
| 15 | `kill-orphans/filter/help-documents-kind-and-all` | `kill-orphans --help` documents `--kind` and `--all` |
| 16 | `kill-orphans/filter/unknown-kind-rejected` | `--kind=not-a-real-kind` → exit 1, stderr error |
| 17 | `open-cleanup/open-reclaims-serve` | `run --open` with mock TUI; after harness reclaim, no live serve for session |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/pty
doctest test ./cmd/agent-run/tests/pty
doctest test -v ./cmd/agent-run/tests/pty/stats/prints-summary
doctest test -v ./cmd/agent-run/tests/pty/kill-orphans/kills-matching-exe
doctest test -v ./cmd/agent-run/tests/pty/kill-orphans/filter/...
doctest test -v ./cmd/agent-run/tests/pty/open-cleanup/open-reclaims-serve
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// ServeSpawnSpec describes a test __serve process started before kill-orphans.
type ServeSpawnSpec struct {
	// Label keys Response.SpawnPIDs (e.g. "ppid1", "child", "tg", "tmp").
	Label string
	// Orphan true: double-fork so PPID becomes 1. false: direct child (PPID ≠ 1).
	Orphan bool
	// SessionID optional; empty generates a unique id. May include "TestGenerated"
	// when the kind marker should appear in argv rather than the binary path.
	SessionID string
	// PathMarker:
	//   "test-generated"  — binary under <tmpdir>/TestGeneratedCase/bin/agent-run
	//   "workdir-at-tmp"  — binary under <tmpdir>/bin (macOS /var/folders/…/T/…)
	//   ""                — use req.AgentRun as-is
	PathMarker string
}

// Request is shared by all leaves under this nested root.
type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Args     []string
	Env      []string

	// Mode selects Run strategy:
	//   ""                — exec agent-run with Args
	//   "kill-orphans"    — optional spawn serve(s), then Args (kill-orphans CLI)
	//   "open-cleanup"    — run --open, record serve, reclaim, report liveness
	Mode string

	ExecTimeout time.Duration

	// Kill-orphans flow (Mode == "kill-orphans")
	// Legacy single orphan spawn (double-fork PPID 1) using req.AgentRun.
	// Ignored when SpawnPlan is non-empty.
	SpawnServe     bool
	ServeSessionID string
	ServeHoldSecs  int // child sleep duration; default 120

	// SpawnPlan, when non-empty, starts each listed serve before the CLI.
	SpawnPlan []ServeSpawnSpec

	// FollowUpArgs, when non-empty, runs a second agent-run invocation after
	// the primary Args (same env/dir). Used to contrast default dry-run vs
	// --kind / --all dry-run in one leaf.
	FollowUpArgs []string

	// Open-cleanup flow (Mode == "open-cleanup")
	OpenInstantAttach bool
	GrokTTYCommand    string
	Prompt            string

	// Populated during Run for asserts
	SpawnedServePID   int
	SpawnedSessionID  string
	TerminalSessionID string
}

// Response captures CLI output and serve liveness around kill/cleanup.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error

	ServePID         int
	ServeAliveBefore bool
	ServeAliveAfter  bool
	// SpawnPIDs maps ServeSpawnSpec.Label → PID when SpawnPlan was used.
	// Also set for legacy SpawnServe under label "orphan" when plan is empty.
	SpawnPIDs map[string]int

	// Follow-up CLI (when Request.FollowUpArgs set)
	FollowUpStdout   string
	FollowUpStderr   string
	FollowUpExitCode int

	TerminalSessionID  string
	RegistryListenAddr string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "kill-orphans":
		return runKillOrphansFlow(t, req)
	case "open-cleanup":
		return runOpenCleanupFlow(t, req)
	default:
		return runAgentRun(t, req, req.Args...)
	}
}
```
