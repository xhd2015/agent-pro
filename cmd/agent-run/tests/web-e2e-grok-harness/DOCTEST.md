# Web E2E Grok Harness Smoke Tests

Minimal smoke tree verifying the shared Playwright/backend grok mock web harness:
`llm-mock-run-grok` + `LLM_MOCK_RUN_GROK_COMMAND` + `--grok-home` /
`--grok-tty-runner-binary` / `--agent-runner grok-tty`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — started with grok mock flags on `--port 0`.
- **llm-mock-run-grok** — built from `./agent/llm/llm-mock/llm-mock-run-grok`.
- **Mock hook** — `LLM_MOCK_RUN_GROK_COMMAND` shell hook seeds `updates.jsonl` under
  `--grok-home` and prints `GROK_TTY_BANNER` for TTY probes.
- **Session API** — `POST /api/agent-run/sessions {runner: grok-tty}`.

**Behaviors**

- POST `grok-tty` session completes with `status=finished`.
- Assistant text contains deterministic mock marker (not random fake-codex output).
- Session path is under `AGENT_RUN_HOME/sessions/grok-tty/<id>/`.

## Version

0.0.2

## Decision Tree

```
web-e2e-grok-harness/
├── DOCTEST.md
├── SETUP.md
└── harness-smoke-grok-tty-session/
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `harness-smoke-grok-tty-session` | Web grok flags + POST grok-tty → finished + deterministic assistant marker |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web-e2e-grok-harness                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web-e2e-grok-harness
doctest test --label-all ./cmd/agent-run/tests/web-e2e-grok-harness

doctest vet ./cmd/agent-run/tests/web-e2e-grok-harness
doctest test ./cmd/agent-run/tests/web-e2e-grok-harness
```

```go
import (
	"bytes"
	"net/http"
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot            string
	TempDir             string
	Home                string
	AgentRun            string
	LLMMockRunGrok      string
	GrokHome            string
	GrokTTYRunnerBinary string
	WebToken            string
	WebBaseURL          string
	WebCmd              *exec.Cmd
	webProcessStderr    *bytes.Buffer
	SessionRunner       string
	SessionID           string
	CreatePrompt        string
	GrokSessionUUID     string
	Env                 []string
}

type Response struct {
	HTTPStatus int
	HTTPBody   string
	Err        error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runHarnessSmokeProbe(t, req)
}
```