# Terminal Poll Tests

Playwright network-monitor tests for `frontend-agent-run` session-page terminal
status polling. Asserts `/api/agent-run/sessions/{runner}/{id}/terminal` is used
only during discovery — not perpetually after `terminal_session_id` or
`available: true` is known.

# DSN (Domain Specific Notion)

**Participants**

- **Chat page** — React session detail route; loads session detail once, tails
  events via SSE while `running`, and may probe terminal availability for TTY
  runners.
- **Session detail API** — `GET /api/agent-run/sessions/{runner}/{id}` returns
  metadata including optional `terminal_session_id` and `status`.
- **Terminal status API** — `GET /api/agent-run/sessions/{runner}/{id}/terminal`
  returns `{available, terminal_session_id}` by resolving the mapped PTY registry.
- **Terminal resolver** — maps web chat id → `terminal_session_id` → live
  `{runner}-registry/<id>.json` listen address.
- **PTY registry** — JSON files under `AGENT_RUN_HOME/{grok-tty,codex-tty}-registry/`;
  may appear after a delay during early discovery.
- **Test harness** — builds `agent-run`, starts fake ptywrap + `agent-run web`,
  seeds session/registry fixtures, runs **playwright-debug** with
  `page.on('request'|'response')` counters for terminal GET traffic.

**Behaviors (sound fix)**

- When session detail already exposes `terminal_session_id`, the chat page sets
  terminal available from detail — **no** repeating `/terminal` poll loop.
- When TTY runner and no `terminal_session_id` yet, poll `/terminal` slowly
  (2–5s) until `available: true`, then **stop**.
- No perpetual 500ms/5s poll after terminal is known.
- Terminal button remains visible when terminal is available on finished chats
  (regression guard).

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/terminal-poll/
├── DOCTEST.md
├── SETUP.md                                    # build, ptywrap, fixtures, network monitors
├── known-terminal-id/                          # session detail already has terminal_session_id
│   ├── SETUP.md
│   ├── no-repeat-poll/                         # finished grok-tty; 8s passive watch; terminal GET ≤1
│   └── terminal-button-visible/                # same seed; terminal button enabled (ui-automation)
└── discovery-needed/                           # no terminal_session_id in session detail initially
    ├── SETUP.md
    └── poll-stops-after-available/             # running; registry after 1s; bounded polls then stop
```

Parameter ranking (most → least significant):

1. **Terminal id source** — known from session detail vs must discover via `/terminal`
2. **Chat lifecycle** — `finished` (passive watch) vs `running` (discovery window)
3. **Registry timing** — present at load vs delayed appearance
4. **Assertion surface** — network GET count vs DOM terminal button

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `known-terminal-id/no-repeat-poll` | Finished `grok-tty` with `terminal_session_id` + live registry; 8s passive watch → terminal GET **≤ 1** |
| 2 | `discovery-needed/poll-stops-after-available` | Running session without `terminal_session_id`; registry after 1s → terminal GET **> 0** during discovery, **stops** after `available: true`, total **≤ 8** in 8s (not 16+) |
| 3 | `known-terminal-id/terminal-button-visible` | Finished session with known mapping; enabled Terminal button still visible (regression) |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/terminal-poll
doctest test -v ./cmd/agent-run/tests/terminal-poll
doctest test -v ./cmd/agent-run/tests/terminal-poll/known-terminal-id/no-repeat-poll --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/terminal-poll/discovery-needed/poll-stops-after-available --label 'chromium && slow'
doctest test -v ./cmd/agent-run/tests/terminal-poll/known-terminal-id/terminal-button-visible --label 'ui-automation'
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type Request struct {
	TempDir    string
	Home       string
	RepoRoot   string
	AgentRun   string
	Port       int
	Token      string
	BaseURL    string
	Env        []string
	Runner            string
	ChatSessionID     string
	RunnerSessionID   string
	TerminalSessionID string
	Status            string
	Prompt            string
	RegistryListenAddr string
	RegistryTranscript string
	RegistryDelay      time.Duration // non-zero: write registry after delay during Run
	Scenario          string
	PlaywrightScript  string
	WatchWindowMs     int

	webCmd *exec.Cmd
}

type Response struct {
	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int
	Err              error
}

func runPlaywrightScript(t *testing.T, script string) (string, string, int, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "playwright-debug", "-e", script)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout := outBuf.String()
	stderr := errBuf.String()
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return stdout, stderr, exitErr.ExitCode(), nil
		}
		if ctx.Err() != nil {
			return stdout, stderr, -1, ctx.Err()
		}
		return stdout, stderr, -1, runErr
	}
	return stdout, stderr, 0, nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	requirePlaywright(t)

	if req.webCmd == nil && req.Port > 0 {
		if err := startWebBackground(t, req); err != nil {
			return nil, err
		}
	}
	if req.BaseURL == "" && req.Port > 0 {
		req.BaseURL = fmt.Sprintf("http://127.0.0.1:%d", req.Port)
	}

	if req.RegistryDelay > 0 && req.RegistryListenAddr != "" {
		startDelayedRegistryWriter(t, req)
	}

	if strings.TrimSpace(req.PlaywrightScript) == "" {
		return nil, fmt.Errorf("PlaywrightScript is empty — leaf Setup must set it")
	}

	stdout, stderr, exitCode, err := runPlaywrightScript(t, req.PlaywrightScript)
	resp := &Response{
		PlaywrightStdout: stdout,
		PlaywrightStderr: stderr,
		PlaywrightExit:   exitCode,
		Err:              err,
	}
	return resp, err
}
```