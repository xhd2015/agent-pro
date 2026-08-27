# agent-run `run` — `--exit-on-idle` on live Codex TUI (`llm-mock-run-codex`)

Coverage backfill: a detach keep-alive **codex-tty** session with
`--exit-on-idle` must let the serve watchdog decide idleness via
runner-agnostic resting-snapshot change + **space probe**
(`pkgs/tty/detection`), not via `DetectInputBox` / screen classifiers.

Uses real **`llm-mock-run-codex`** + sibling **`llm-mock`** + PATH **`codex`**.
No stand-in `codex` script, no `LLM_MOCK_RUN_CODEX_COMMAND` /
`AGENT_RUN_CODEX_TTY_COMMAND` fake chrome. Isolated nested root so sibling
`cmd/agent-run/tests` trees stay untouched.

**Out of scope:** Grok (`run-exit-on-idle-grok-tty`), L2 Tick fakes
(`tests/agentruncli/idle-watchdog`), L2 `DetectInputBox` fixtures, SPL
`debugging open`, treating `medium`/`max` as occupancy.

# DSN (Domain Specific Notion)

A detach keep-alive **codex-tty** session with `--exit-on-idle` must let the
serve watchdog use resting-snapshot stability + universal space-probe occupy.

**Participants**

- **`agent-run run --detach --exit-on-idle --idle-timeout=10s`** — writes
  `idle-policy.json`, starts `__serve` + real Codex TUI.
- **`llm-mock-run-codex` + sibling `llm-mock`** — `--agent-runner-binary`;
  mock Responses API (`reply with exactly: pong` → `pong`).
- **PATH `codex`** — live TUI (skip if missing). Isolated
  `AGENT_RUN_HOME` / `CODEX_HOME` / `LLM_MOCK_CODEX_HOME`.
- **Idle watchdog** — `pkgs/tty/detection/idle` samples at 0 / 5s / 10s.
  Occupancy is space → exactly-+1-space compare → DEL (not `tty status input_box`).
- **Composer occupancy** — placeholder/empty → three idle samples → SoftExit
  `/exit` → not live. Real no-submit draft → occupied → hits reset → stays live.

**Behaviors**

- Finished mock turn with empty/placeholder composer: after timeout + grace +
  probe slack the TTY is gone.
- After ready, a distinctive no-submit draft stays in the composer: after the
  same window the TTY is still live.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run-exit-on-idle-codex-tty/
├── DOCTEST.md
├── SETUP.md
└── detach/                         # --detach keep-alive (same serve watchdog as --open)
    ├── SETUP.md
    ├── placeholder-exits/          # finished mock turn, empty/placeholder → not live
    └── draft-holds/                # after ready, distinctive draft no-submit → still live
```

Parameter ranking (most → least significant):

1. **Composer occupancy class** — placeholder/empty vs real draft
2. **Keep-alive shape** — `--detach` (serve idle watchdog)
3. **Observation** — gone after timeout+grace vs still live

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `detach/placeholder-exits` | After 10s idle + 5s grace + probe slack: session not live |
| 2 | `detach/draft-holds` | After ready, no-submit draft: session still live past the same window |

## How to Run

```sh
# From the agent-pro module root (needs real `codex` on PATH):
GOWORK=off doctest vet ./cmd/agent-run/tests/run-exit-on-idle-codex-tty
GOWORK=off doctest test --label e2e ./cmd/agent-run/tests/run-exit-on-idle-codex-tty

GOWORK=off doctest test -v --label e2e \
  ./cmd/agent-run/tests/run-exit-on-idle-codex-tty/detach/placeholder-exits
GOWORK=off doctest test -v --label e2e \
  ./cmd/agent-run/tests/run-exit-on-idle-codex-tty/detach/draft-holds
```

Leaves are `label: e2e` (process boundary: `agent-run` + real Codex via
`llm-mock-run-codex`). Skip when `codex` is not on PATH.

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// Request drives one detach + occupancy observation.
type Request struct {
	RepoRoot        string
	TempDir         string
	Home            string
	CodexHome       string
	Workspace       string
	AgentRun        string
	LLMMock         string
	LLMMockRunCodex string
	MockConfigFile  string
	Env             []string

	SessionID    string
	Prompt       string
	IdleTimeout  time.Duration
	ObserveAfter time.Duration
	Op           string // placeholder-exit | draft-hold
	Draft        string // no-submit composer text for draft-hold
}

// Response is detach + optional draft inject + tty liveness at ObserveAfter.
type Response struct {
	DetachStdout string
	DetachStderr string
	DetachExit   int
	PolicyJSON   string

	DraftInjected bool
	DraftStdout   string
	DraftStderr   string
	DraftSnapshot string

	StatusJSON   string
	TCPReachable bool
	Alive        bool
	StatusExit   int
	PID          int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	return runDetachThenObserve(t, req)
}

type ttyStatusJSON struct {
	ScreenStatus  string `json:"screen_status"`
	InputBox      string `json:"input_box"`
	Sendable      bool   `json:"sendable"`
	SendableState string `json:"sendable_state"`
	TCPReachable  bool   `json:"tcp_reachable"`
	PID           int    `json:"pid"`
}
```
