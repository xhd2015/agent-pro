# agent-run run --agent-runner-config-home Tests

Doc-style tests for `agent-run run --agent-runner grok-tty --agent-runner-config-home PATH`.
Session discovery reads `PATH` instead of default `~/.grok`. The PTY child argv is prefixed
with `env GROK_HOME=PATH ...` only — `AGENT_RUNNER_CONFIG_HOME` is **not** passed to the child.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --agent-runner-config-home PATH` sets the grok data directory for
  discovery and child env.
- **PTY child** — argv prefixed with `env GROK_HOME=PATH <binary> ...`; no
  `AGENT_RUNNER_CONFIG_HOME=` in child environment.
- **Seeded grok session** — `$PATH/sessions/<encoded-cwd>/<uuid>/updates.jsonl` supplies the
  authoritative ACP stream for tailing.
- **Fake runner script** — replaces `grok` for banner wait and env logging.

**Behaviors**

- Discovery prints `grok-tty: grok session <uuid>` and tails `updates.jsonl` from the config home.
- Env-logging fake runner stderr shows `GROK_HOME=PATH` and omits `AGENT_RUNNER_CONFIG_HOME=`.

## Version

0.0.2

## Decision Tree

```
run/agent-runner-config-home/
├── DOCTEST.md
├── SETUP.md
├── discovers-session/           # seeded updates.jsonl under config home → stderr grok session
└── child-env-grok-home-only/    # child env has GROK_HOME only (no AGENT_RUNNER_CONFIG_HOME)
```

Parameter ranking (most → least significant):

1. **Assertion target** — session discovery vs child env contract
2. **Runner backend** — fake script (hook unset) with banner + hold/env log

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `discovers-session` | Seed `updates.jsonl` under config home; stderr grok session uuid; streams marker |
| 2 | `child-env-grok-home-only` | Fake env logger; stderr has `GROK_HOME=`; no `AGENT_RUNNER_CONFIG_HOME=` |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/agent-runner-config-home
doctest test ./cmd/agent-run/tests/run/agent-runner-config-home
doctest test -v ./cmd/agent-run/tests/run/agent-runner-config-home/discovers-session
doctest test -v ./cmd/agent-run/tests/run/agent-runner-config-home/child-env-grok-home-only
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot              string
	TempDir               string
	Home                  string
	AgentRun              string
	Args                  []string
	Env                   []string
	AgentRunnerBinary     string
	AgentRunnerConfigHome string
	RunnerScriptPath      string
	GrokSessionUUID       string
	GrokUpdatesPath       string
	Prompt                string
	EnvProbePath          string
	ExecTimeout           time.Duration
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```