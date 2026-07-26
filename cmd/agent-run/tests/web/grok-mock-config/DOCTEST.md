# agent-run web grok mock config Tests

Doc-style tests for `agent-run web --grok-home PATH --grok-tty-runner-binary SPEC
--agent-runner grok-tty`. Web-created `grok-tty` sessions use the configured mock binary
and shared grok home for session discovery.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web server** — in-memory `webRunConfig` carries `GrokHome` and
  `GrokTTYRunnerBinary`; not persisted to disk beyond runtime.
- **Session API** — `POST /api/agent-run/sessions` with runner `grok-tty` starts a background
  grok-tty run using web flags.
- **Mock runner binary** — wrapper or `llm-mock-run-grok` invoked as grok-tty child; proves
  argv/binary selection via log file or hook.
- **Shared grok home** — `--grok-home` seeds/hosts `sessions/<encoded-cwd>/<uuid>/` for
  discovery under the configured path.

**Behaviors**

- Web startup flags flow into grok-tty `AgentRunnerBinary` and `AgentRunnerConfigHome`.
- POST session completes with grok session data under the specified `--grok-home`.
- Child process uses the mock binary (not default `grok`).

## Version

0.0.2

## Decision Tree

```
web/grok-mock-config/
├── DOCTEST.md
├── SETUP.md
└── uses-mock-binary-and-home/    # web flags → grok-tty session uses mock + shared home
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `uses-mock-binary-and-home` | Web with `--grok-home` + `--grok-tty-runner-binary`; POST grok-tty session; data under home; mock binary probe |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web/grok-mock-config                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web/grok-mock-config
doctest test --label-all ./cmd/agent-run/tests/web/grok-mock-config

doctest vet ./cmd/agent-run/tests/web/grok-mock-config
doctest test ./cmd/agent-run/tests/web/grok-mock-config
doctest test -v ./cmd/agent-run/tests/web/grok-mock-config/uses-mock-binary-and-home
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
	RepoRoot              string
	TempDir               string
	Home                  string
	AgentRun              string
	LLMMockRunGrok        string
	MockWrapperPath       string
	ArgvProbePath         string
	GrokHome              string
	GrokTTYRunnerBinary   string
	WebToken              string
	WebPort               int
	WebBaseURL            string
	WebCmd                *exec.Cmd
	webProcessStderr      *bytes.Buffer
	SessionRunner         string
	SessionID             string
	CreatePrompt          string
	GrokSessionUUID       string
	Env                   []string
	ExecTimeout           time.Duration
}

type Response struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Err        error
	HTTPStatus int
	HTTPBody   string
	ArgvProbe  string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runWebGrokMockProbe(t, req)
}
```