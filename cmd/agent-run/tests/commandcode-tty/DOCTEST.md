# commandcode-tty agent-run Tests

Doc-style tests for `agent-run run --agent-runner commandcode-tty` and
`agent-run send`. Each run spawns `llm-mock-run-commandcode` in an adhoc PTY
server. Session ids are printed to stderr as `commandcode-tty: session-N`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` validates the runner id, starts a TTY backend, injects
  `-p` flag (headless) or omits it (open), waits for scrollback content, captures
  response text, and persists the session. `send` enqueues a follow-up message and
  waits for delivery to the live PTY.
- **llm-mock-run-commandcode** — shortcut binary that starts an in-process mock HTTP
  server, sets sandbox env vars, isolates HOME, and runs `cmd` in the foreground.
- **cmd** — the Command Code CLI, running inside the PTY.
- **Mock HTTP server** — serves `/alpha/whoami`, `/alpha/generate` (NDJSON),
  `/alpha/lifecycle-events`, `/alpha/fingerprint/record`.

**Behaviors**

- `run --json "Hello"` injects `-p`, waits for scrollback content, captures response.
- `run --open "Hello"` starts keep-alive PTY, prints session id on stderr.
- `send <id> "Hello 2"` enqueues, drains, delivers to live PTY.
- Missing `--agent-runner-binary` fails with clear error.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/commandcode-tty/
├── DOCTEST.md
├── SETUP.md                           # build agent-run, mock binary
├── run-headless/                      # run --json "Hello"
│   ├── SETUP.md
│   └── ASSERT.md
├── open-keep-alive/                   # run --open "Hello"
│   ├── SETUP.md
│   └── ASSERT.md
├── send-followup/                     # open + send + snapshot
│   ├── SETUP.md
│   └── ASSERT.md
└── errors-missing-binary/             # missing --agent-runner-binary
    ├── SETUP.md
    └── ASSERT.md
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `run-headless` | Headless run with `--json` exits 0, emits valid JSON with message text |
| 2 | `open-keep-alive` | `--open` prints session id on stderr, snapshot returns response text |
| 3 | `send-followup` | Send to open session delivers second message, both prompts in scrollback |
| 4 | `errors-missing-binary` | Missing `--agent-runner-binary` fails with clear error |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/commandcode-tty                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/commandcode-tty
doctest test --label-all ./cmd/agent-run/tests/commandcode-tty

doctest vet ./cmd/agent-run/tests/commandcode-tty
doctest test ./cmd/agent-run/tests/commandcode-tty
doctest test -v ./cmd/agent-run/tests/commandcode-tty/run-headless
doctest test -v ./cmd/agent-run/tests/commandcode-tty/send-followup
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot       string
	TempDir        string
	Home           string
	AgentRun       string
	Args           []string
	Env            []string
	MockBinary     string
	SessionID      string
	SkipMockBuild  bool
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Snapshot string
	SessionID string
	SendID   string
	Duration time.Duration
}

func execSnapshot(t *testing.T, agentRun, sessionID string, env ...string) string {
	cmd := exec.Command(agentRun, "tty", "snapshot", sessionID)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	start := time.Now()
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(req.AgentRun, req.Args...)
	if len(req.Env) > 0 {
		cmd.Env = append(os.Environ(), req.Env...)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
			return resp, nil
		}
		return resp, err
	}

	re := regexp.MustCompile(`commandcode-tty:\s*(session-\d+)`)
	if m := re.FindStringSubmatch(resp.Stderr); len(m) > 1 {
		resp.SessionID = m[1]
	}
	return resp, nil
}
```
