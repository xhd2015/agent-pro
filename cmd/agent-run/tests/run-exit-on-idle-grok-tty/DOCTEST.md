# agent-run `run` — `--exit-on-idle` on live grok-tty chrome (`llm-mock-run-grok`)

Classic TDD e2e for the crime scene: after a finished Grok turn the keep-alive
TTY must classify as **idle + empty composer** and **exit** when
`--exit-on-idle --idle-timeout` elapses. Isolated nested root so sibling
`cmd/agent-run/tests` trees stay GREEN.

Uses the product mock runner **`llm-mock-run-grok`** (built from
`./agent/llm/llm-mock/llm-mock-run-grok`) plus `LLM_MOCK_RUN_GROK_COMMAND`
chrome — the same hook style as `grok-discovery-race`. Does **not** install a
stand-in `grok`/`codex` binary.

**Out of scope:** L2 help/parse (`run-exit-on-idle/`), L2 `Tick` fakes
(`tests/agentruncli/idle-watchdog/`), real PATH `grok`, local-bot.

# DSN (Domain Specific Notion)

A detach keep-alive grok-tty session with modern boxed chrome must look idle
to `tty status` and the serve idle watchdog must `/exit` after the timeout.

**Participants**

- **`agent-run run --detach --exit-on-idle --idle-timeout=2s`** — writes
  `idle-policy.json`, starts `__serve` + grok-tty PTY.
- **`llm-mock-run-grok`** — `--agent-runner-binary`. Hook
  `LLM_MOCK_RUN_GROK_COMMAND` prints finished-turn chrome (`Worked for`, boxed
  `│ ❯ … │`, `always-approve`, `Shift+Tab:mode`) and holds the PTY.
- **`agent-run tty status --json`** — `screen_status`, `input_box`,
  `sendable`, `tcp_reachable`.
- **Idle watchdog** — `SampleIsIdle` requires sendable + `screen==idle` +
  `input_box==empty` + empty queue; then timeout + 5s grace.

**Behaviors**

- Shortly after detach (sendable), `screen_status` is `idle` and `input_box`
  is `empty` (not `banner` / `occupied` on an empty boxed composer).
- After `idle-timeout` + grace, the TTY is gone (not reachable / not live).

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run-exit-on-idle-grok-tty/
├── DOCTEST.md
├── SETUP.md
└── detach/                                      # --detach keep-alive (same serve watchdog as --open)
    ├── SETUP.md
    ├── live-chrome-classifies-idle/             # T+settle: screen idle, box empty
    └── live-chrome-exits/                       # after timeout+grace: not live
```

Parameter ranking (most → least significant):

1. **Keep-alive shape** — `--detach` (serve idle watchdog)
2. **Chrome** — live boxed empty composer (no `GROK_TTY_BANNER`)
3. **Observation** — classify at settle vs gone after timeout

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `detach/live-chrome-classifies-idle` | After detach settle: `screen_status=idle`, `input_box=empty`, sendable |
| 2 | `detach/live-chrome-exits` | After 2s idle + 5s grace: session not live |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./cmd/agent-run/tests/run-exit-on-idle-grok-tty
doctest test --label e2e ./cmd/agent-run/tests/run-exit-on-idle-grok-tty

doctest test -v --label e2e ./cmd/agent-run/tests/run-exit-on-idle-grok-tty/detach/live-chrome-classifies-idle
doctest test -v --label e2e ./cmd/agent-run/tests/run-exit-on-idle-grok-tty/detach/live-chrome-exits
```

Leaves are `label: e2e` (process boundary: `agent-run` + `llm-mock-run-grok` PTY).
Expect **RED** until grok-tty `DetectScreenStatus` / `DetectInputBox` treat
this chrome as idle + empty (then the watchdog can exit).

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

// Request drives one detach + tty-status observation.
type Request struct {
	RepoRoot       string
	TempDir        string
	Home           string
	AgentRun       string
	LLMMockRunGrok string
	GrokHome       string
	Env            []string

	SessionID    string
	Prompt       string
	IdleTimeout  time.Duration
	ObserveAfter time.Duration // sleep after detach before final status
	Op           string        // classify | exit
}

// Response is detach + one tty status snapshot at ObserveAfter.
type Response struct {
	DetachStdout string
	DetachStderr string
	DetachExit   int
	PolicyJSON   string

	StatusHuman  string
	StatusJSON   string
	ScreenStatus string
	InputBox     string
	Sendable     bool
	SendableState string
	TCPReachable bool
	Alive        bool
	StatusExit   int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	return runDetachThenStatus(t, req)
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
