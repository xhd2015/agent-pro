# agent-run pty Subcommand Tests

Doc-style tests for `agent-run pty` — PTY resource stats, kill of orphan
`__serve*` processes, and harness verification that `run --open` leaves no
serve after reclaim cleanup.

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
  listen_addr). Default kill set is **all** matching serves; tests isolate with
  `--exe PATH` so host brainstorm/seatalk sessions are not destroyed.
- **Harness reclaim** — `t.Cleanup` / `ReclaimSessionID` (or equivalent kill of
  serve PID + registry remove) used by TestGenerated open paths so keep-alive
  serves do not leak after the case ends. Product KeepAlive semantics for real
  users are unchanged.

**Behaviors**

- `agent-run pty --help` lists `stats` and `kill-orphans` (and documents
  `--dry-run` / `--exe` on kill-orphans help).
- `agent-run --help` mentions `pty` alongside existing commands.
- Unknown `pty` subcommand → exit 1, stderr error.
- `agent-run pty stats` prints human-readable PTY limit / masters / free
  estimate (best-effort) and agent-run `__serve` counts/categories; exit 0 on
  success (partial probes OK). Successful stdout ends with trailing `\n`.
- `agent-run pty kill-orphans [--dry-run] [--exe PATH]`:
  - default targets all agent-run `__serve*` processes
  - `--exe PATH` restricts to serves whose executable matches PATH
  - `--dry-run` lists targets without killing; empty set prints a clear
    no-match / no-orphans line; exit 0
  - without `--dry-run`, terminate targets (SIGTERM then SIGKILL); never kill
    self; exit 0 when terminations were attempted (exit 1 on hard kill failure)
  - successful stdout ends with trailing `\n`
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
├── kill-orphans/                         # always use --exe in live tests
│   ├── help-documents-flags/             # kill-orphans --help: --dry-run, --exe
│   ├── dry-run-lists-serve/              # spawn serve; dry-run --exe lists PID; still alive
│   ├── dry-run-none/                     # --exe unique path, no serves → no-match, exit 0
│   └── kills-matching-exe/               # spawn serve; kill-orphans --exe; process gone
└── open-cleanup/
    └── open-reclaims-serve/              # run --open + harness reclaim → serve gone
```

Parameter ranking (most → least significant):

1. **Operation surface** — help vs stats vs kill-orphans vs open-cleanup
2. **Help surface** — `pty --help` vs top-level vs unknown subcommand
3. **Kill mode** — dry-run vs real kill vs kill-orphans help
4. **Target set** — matching `--exe` serve present vs none

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/lists-subcommands` | `pty --help` lists `stats` and `kill-orphans`; exit 0; trailing `\n` |
| 2 | `help/top-level-mentions-pty` | `agent-run --help` mentions `pty`; exit 0 |
| 3 | `help/unknown-subcommand` | `pty not-a-real-cmd` → exit 1, stderr error |
| 4 | `stats/prints-summary` | `pty stats` exit 0; stdout has limit/serve-related text + trailing `\n` |
| 5 | `kill-orphans/help-documents-flags` | `pty kill-orphans --help` documents `--dry-run` and `--exe` |
| 6 | `kill-orphans/dry-run-lists-serve` | Spawn test `__serve` via test binary; dry-run `--exe` lists PID; process still alive |
| 7 | `kill-orphans/dry-run-none` | dry-run `--exe` unique path with no serves → no-match message; exit 0; trailing `\n` |
| 8 | `kill-orphans/kills-matching-exe` | Spawn serve; `kill-orphans --exe` test binary → process gone; exit 0 |
| 9 | `open-cleanup/open-reclaims-serve` | `run --open` with mock TUI; after harness reclaim, no live serve for session |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/pty
doctest test ./cmd/agent-run/tests/pty
doctest test -v ./cmd/agent-run/tests/pty/stats/prints-summary
doctest test -v ./cmd/agent-run/tests/pty/kill-orphans/kills-matching-exe
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
	//   "kill-orphans"    — optional spawn serve, then Args (kill-orphans CLI)
	//   "open-cleanup"    — run --open, record serve, reclaim, report liveness
	Mode string

	ExecTimeout time.Duration

	// Kill-orphans flow (Mode == "kill-orphans")
	SpawnServe     bool
	ServeSessionID string
	ServeHoldSecs  int // child sleep duration; default 120

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

	ServePID          int
	ServeAliveBefore  bool
	ServeAliveAfter   bool
	TerminalSessionID string
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
