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
  a shared `GROK_HOME` when no `--agent-runner-config-home` is set. Under Option C (no hooks)
  it runs **real `grok`**, which panics on invalid UTF-8 argv (`std::env::args`).

**Behaviors**

- Without `AGENT_RUN_GROK_TTY_COMMAND`, grok-tty spawns `SPEC` (resolved path + user flags +
  defaults) inside the PTY.
- Binary spec `--model inner` prevents CLI `--model outer` from being appended.
- `llm-mock-run-grok` without config home triggers auto-provision so discovery matches the
  child `GROK_HOME`; streamed events must not fall back to scrollback `GROK_HOME=` pollution.
- Invalid UTF-8 PROMPT: direct `llm-mock-run-grok` must crash real `grok` within **3s**;
  `agent-run` + same binary must **not** crash (normalize) and snapshot shows the message
  (`invalid-utf8-open-prompt/*`, label `real-grok`).

## Version

0.0.2

## Decision Tree

```
run/agent-runner-binary/
├── DOCTEST.md
├── SETUP.md
├── fake-binary-receives-argv/     # fake script records argv; defaults visible in stderr
├── inner-model-wins/              # spec --model inner beats CLI --model outer
├── llm-mock-auto-home/            # llm-mock-run-grok auto-provisions grok home; discovery OK
└── invalid-utf8-open-prompt/      # group: real llm-mock-run-grok + real grok, 3s budgets
    ├── direct-llm-mock-crash/     # A: mock alone → env.rs crash ≤3s
    └── agent-run-normalized/      # B: agent-run wrap → no crash ≤3s + snapshot message
```

Parameter ranking (most → least significant):

1. **Binary spec shape** — argv-recording script vs model override vs llm-mock shortcut
2. **CLI model flag** — present only for inner-model-wins leaf
3. **Config home** — explicit only in llm-mock-auto-home negative (must be absent)
4. **PROMPT UTF-8 path** — direct mock crash vs agent-run normalize (real grok)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `fake-binary-receives-argv` | Fake script via `--agent-runner-binary`; stderr shows argv with grok-tty defaults |
| 2 | `inner-model-wins` | Spec includes `--model inner`; CLI `--model outer` not appended |
| 3 | `llm-mock-auto-home` | `llm-mock-run-grok` without config home; grok session discovered; no scrollback `GROK_HOME=` assistant pollution |
| 4 | `invalid-utf8-open-prompt/direct-llm-mock-crash` | Direct `llm-mock-run-grok` + invalid UTF-8 → env.rs crash ≤3s (`label: real-grok`) |
| 5 | `invalid-utf8-open-prompt/agent-run-normalized` | `agent-run` + same binary/prompt → no crash ≤3s; snapshot has message (`label: real-grok`) |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/run/agent-runner-binary                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/run/agent-runner-binary
doctest test --label-all ./cmd/agent-run/tests/run/agent-runner-binary

doctest vet ./cmd/agent-run/tests/run/agent-runner-binary
doctest test ./cmd/agent-run/tests/run/agent-runner-binary
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/fake-binary-receives-argv
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/inner-model-wins
doctest test -v ./cmd/agent-run/tests/run/agent-runner-binary/llm-mock-auto-home
doctest test --label real-grok -v ./cmd/agent-run/tests/run/agent-runner-binary/invalid-utf8-open-prompt
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
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
	// WallClockLimit: if >0, start the command and stop waiting after this
	// duration (native Go execCmdWallClock). Surviving the window returns exit 0.
	WallClockLimit time.Duration
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```