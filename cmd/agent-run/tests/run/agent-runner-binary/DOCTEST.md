# agent-run run --agent-runner-binary Tests

Doc-style tests for `agent-run run --agent-runner grok-tty --agent-runner-binary SPEC`.
`SPEC` is a bare binary path or shell-style `binary flag...`; grok-tty appends default
flags unless the inner spec already supplies them. Inner `--model` / `--resume` win over
CLI `--model` / `--session`. `AGENT_RUN_GROK_TTY_COMMAND` is unset so the binary spec
drives the PTY argv (not the legacy hook).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --agent-runner grok-tty --agent-runner-binary SPEC` passes the
  spec to grok-tty; blocks until the PTY session exits.
- **PTY child** — interactive runner inside ptywrap; argv built from `SPEC` plus grok-tty
  defaults (`--always-approve`, `--permission-mode=bypassPermissions`, optional `--model`,
  `--resume`) when not already present in `SPEC`.
- **Fake runner script** — deterministic shell binary replacing `grok` for argv and
  scrollback assertions.
- **llm-mock-run-grok** — shortcut orchestrator used as `--agent-runner-binary`; auto-provisions
  a shared `GROK_HOME` when no `--agent-runner-config-home` is set.

**Behaviors**

- Without `AGENT_RUN_GROK_TTY_COMMAND`, grok-tty spawns `SPEC` (resolved path + user flags +
  defaults) inside the PTY.
- Binary spec `--model inner` prevents CLI `--model outer` from being appended.
- `llm-mock-run-grok` without config home triggers auto-provision so discovery matches the
  child `GROK_HOME`; streamed events must not fall back to scrollback `GROK_HOME=` pollution.

## Version

0.0.2

## Decision Tree

```
run/agent-runner-binary/
├── DOCTEST.md
├── SETUP.md
├── fake-binary-receives-argv/     # fake script records argv; defaults visible in stderr
├── inner-model-wins/              # spec --model inner beats CLI --model outer
└── llm-mock-auto-home/            # llm-mock-run-grok auto-provisions grok home; discovery OK
```

Parameter ranking (most → least significant):

1. **Binary spec shape** — argv-recording script vs model override vs llm-mock shortcut
2. **CLI model flag** — present only for inner-model-wins leaf
3. **Config home** — explicit only in llm-mock-auto-home negative (must be absent)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fake-binary-receives-argv` | Fake script via `--agent-runner-binary`; stderr shows argv with grok-tty defaults |
| 2 | `inner-model-wins` | Spec includes `--model inner`; CLI `--model outer` not appended |
| 3 | `llm-mock-auto-home` | `llm-mock-run-grok` without config home; grok session discovered; no scrollback `GROK_HOME=` assistant pollution |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/agent-runner-binary
doctest test ./cmd/agent-run/tests/run/agent-runner-binary
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/fake-binary-receives-argv
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/inner-model-wins
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/llm-mock-auto-home
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot          string
	TempDir           string
	Home              string
	AgentRun          string
	LLMMockRunGrok    string
	Args              []string
	Env               []string
	AgentRunnerBinary string
	RunnerScriptPath  string
	ArgvLogPath       string
	ArgvProbePath     string
	GrokHomeProbePath string
	GrokHome          string
	GrokSessionUUID   string
	GrokUpdatesPath   string
	Prompt            string
	ExecTimeout       time.Duration
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