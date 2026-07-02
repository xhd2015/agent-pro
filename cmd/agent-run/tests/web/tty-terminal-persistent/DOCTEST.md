# agent-run Web TTY Persistent Terminal Tests

Follow-up doc-style tests for real web `codex-tty` / `grok-tty` sessions where
the chat route id, provider resume id, and backend PTY registry id are distinct.

# DSN (Domain Specific Notion)

**Participants**

- **Web chat session** — browser route and session storage identity such as
  `web_abc123`; it owns chat metadata and timeline events.
- **Provider runner session** — provider-specific resume or transcript identity
  such as a Codex rollout id; it is stored as `runner_session_id` and is not a
  terminal attach key.
- **Terminal session mapping** — durable session metadata field
  `terminal_session_id` that maps a web chat to a PTY registry id such as
  `session-1`.
- **TTY registry** — live registry JSON under
  `AGENT_RUN_HOME/{codex-tty,grok-tty}-registry/<terminal_session_id>.json`,
  containing the PTY listen address.
- **Terminal resolver** — resolves
  `(runner, web_chat_session_id) -> terminal_session_id -> registry entry` and
  reports availability independently from chat turn status.
- **Terminal websocket proxy** — attaches browser clients to the mapped PTY
  without creating a replacement PTY when the existing mapped PTY is live.
- **Chat page** — shows the terminal button whenever terminal status reports
  `available: true`, even when the chat session is `finished`.
- **Follow-up sender** — posts a new user message to the same web chat and
  reuses the mapped PTY instead of allocating a fresh registry id.
- **Test harness** — builds `agent-run`, starts `agent-run web` with isolated
  `AGENT_RUN_HOME`, writes session/registry fixtures with distinct ids, probes
  HTTP/websocket endpoints, and uses `playwright-debug` for the browser icon
  check.

**Behaviors**

- Terminal status uses `terminal_session_id`, not the web chat id and not
  `runner_session_id`.
- A `finished` chat can still have an available terminal while its mapped PTY is
  alive.
- Re-entering the same chat and attaching websocket clients reaches the same
  PTY registry id.
- Follow-up messages in the same chat preserve the terminal mapping and do not
  create `session-2` while `session-1` is live.
- Stale mapped registry entries report `available:false` while still exposing
  that a terminal mapping exists.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web/tty-terminal-persistent/
├── DOCTEST.md
├── SETUP.md
├── identity-mapping/
│   ├── status-uses-terminal-session-id/
│   └── stale-terminal-id-distinct-from-missing-mapping/
├── finished-availability/
│   └── status-available-after-finished/
├── running-availability/
│   └── web-created-running-session-terminal-available/
├── lifecycle-reuse/
│   ├── reenter-websocket-uses-same-pty/
│   ├── follow-up-reuses-same-pty/
│   ├── follow-up-live-pty-keeps-session-running/
│   ├── follow-up-ignores-terminal-control-snapshot/
│   ├── follow-up-clears-stale-terminal-input/
│   └── web-created-session-terminal-stays-available/
└── frontend-ui/
    ├── finished-status-terminal-icon/
    ├── modal-hides-session-id-control-frame/
    ├── modal-renders-real-interactive-terminal/
    ├── running-session-modal-attaches-live-terminal/
    ├── running-session-terminal-icon-before-response/
    ├── real-codex-terminal-stale-input-follow-up/
    ├── terminal-close-follow-up-not-duplicated/
    ├── terminal-close-follow-up-submits-enter/
    └── web-created-finished-session-terminal-icon/
```

Parameter ranking (most -> least significant):

1. **Resolver identity source** — `terminal_session_id` mapping vs stale mapped
   PTY vs absent direct registry id.
2. **Chat lifecycle state** — `finished` chat with live PTY vs follow-up turn.
3. **Attach operation** — HTTP status vs websocket attach vs browser icon.
4. **Specific runner** — representative `codex-tty`; the contract is identical
   for `grok-tty`.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `identity-mapping/status-uses-terminal-session-id` | Status resolves `web_* -> terminal_session_id -> session-1`, ignoring `runner_session_id` |
| 2 | `identity-mapping/stale-terminal-id-distinct-from-missing-mapping` | Stale mapped PTY is unavailable but still reports the stored terminal id |
| 3 | `finished-availability/status-available-after-finished` | Finished chat still reports terminal available while mapped PTY is live |
| 4 | `lifecycle-reuse/reenter-websocket-uses-same-pty` | Two attaches after navigation hit the same mapped PTY registry id |
| 5 | `lifecycle-reuse/follow-up-reuses-same-pty` | Follow-up preserves terminal id and does not create a second registry entry |
| 6 | `frontend-ui/finished-status-terminal-icon` | Browser shows terminal button on a finished chat when terminal status is available |
| 7 | `lifecycle-reuse/web-created-session-terminal-stays-available` | Real web-created codex-tty session stores a terminal id and remains attachable after finish |
| 8 | `frontend-ui/web-created-finished-session-terminal-icon` | Browser shows terminal button on the generated route for a web-created finished codex-tty session |
| 9 | `frontend-ui/modal-hides-session-id-control-frame` | Terminal modal attaches the mapped PTY and does not render ptywrap `session_id` control JSON |
| 10 | `frontend-ui/running-session-terminal-icon-before-response` | Browser shows terminal button for a running codex-tty session before the assistant response finishes |
| 11 | `frontend-ui/modal-renders-real-interactive-terminal` | Terminal modal uses a real terminal emulator surface: ANSI control bytes are interpreted and typed input reaches the PTY |
| 12 | `running-availability/web-created-running-session-terminal-available` | Web-created running codex-tty session reports terminal available before the assistant response finishes |
| 13 | `frontend-ui/running-session-modal-attaches-live-terminal` | Browser terminal modal attaches live tty content while a codex-tty turn is still running |
| 14 | `lifecycle-reuse/follow-up-live-pty-keeps-session-running` | Follow-up written to a live PTY keeps the web chat running instead of marking finished immediately |
| 15 | `frontend-ui/terminal-close-follow-up-not-duplicated` | Closing the terminal after the first tty response does not let a stale SSE stream duplicate the next user message |
| 16 | `frontend-ui/terminal-close-follow-up-submits-enter` | Follow-up after closing the terminal is submitted to the existing TTY, not left typed in the input buffer |
| 17 | `lifecycle-reuse/follow-up-ignores-terminal-control-snapshot` | Live terminal follow-up does not store prompt echo or terminal session/control JSON as assistant response |
| 18 | `lifecycle-reuse/follow-up-clears-stale-terminal-input` | Live terminal follow-up clears stale prompt input before submitting the chat follow-up |
| 19 | `frontend-ui/real-codex-terminal-stale-input-follow-up` | Real Codex terminal close/open follow-up does not submit stale terminal input or persist raw terminal output as chat |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/web/tty-terminal-persistent
doctest test ./cmd/agent-run/tests/web/tty-terminal-persistent
doctest test --label ui-automation ./cmd/agent-run/tests/web/tty-terminal-persistent
doctest test --label "codex && ui-automation" ./cmd/agent-run/tests/web/tty-terminal-persistent
doctest test -v ./cmd/agent-run/tests/web/tty-terminal-persistent/identity-mapping
doctest test -v ./cmd/agent-run/tests/web/tty-terminal-persistent/lifecycle-reuse
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Env      []string

	Mode string // "http" | "ws" | "followup" | "ui"

	WebToken   string
	WebBaseURL string
	WebCmd     *exec.Cmd
	webStdout  *bytes.Buffer
	webStderr  *bytes.Buffer

	Runner            string
	ChatSessionID     string
	RunnerSessionID   string
	TerminalSessionID string
	Status            string
	Prompt            string
	FollowUpPrompt    string

	HTTPMethod string
	HTTPPath   string
	HTTPBody   string
	HTTPAuth   string

	RegistryListenAddr string
	RegistryTranscript string
	RegistryIDsBefore  []string
	PTYInputSeen       *string

	WSPath string
	WSAuth string

	PlaywrightScript string

	PTYConnectionCount *int
}

type Response struct {
	HTTPStatus int
	HTTPBody   string

	FirstHTTPStatus  int
	FirstHTTPBody    string
	SecondHTTPStatus int
	SecondHTTPBody   string
	FollowUpStatus   int
	FollowUpBody     string
	FollowUpSessionStatus int
	FollowUpSessionBody   string
	RegistryIDsAfter []string

	WSOutput          string
	PTYConnectionSeen int
	WSError           string

	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int

	Err error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "", "http":
		return runHTTP(t, req)
	case "ws":
		return runWebSocketTwice(t, req)
	case "followup":
		return runFollowUpReuseProbe(t, req)
	case "ui":
		return runPlaywright(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func commandErrorExit(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return 124
	}
	return 1
}

func drainAndClose(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func runHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	method := strings.ToUpper(strings.TrimSpace(req.HTTPMethod))
	if method == "" {
		method = http.MethodGet
	}
	path := req.HTTPPath
	if path == "" {
		path = terminalStatusPath(req.Runner, req.ChatSessionID)
	}
	auth := req.HTTPAuth
	if auth == "" {
		auth = req.WebToken
	}
	contentType := ""
	if req.HTTPBody != "" {
		contentType = "application/json"
	}
	status, body := doHTTP(t, method, req.WebBaseURL+path, auth, contentType, req.HTTPBody)
	return &Response{HTTPStatus: status, HTTPBody: body}, nil
}

func runPlaywright(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skipf("playwright-debug not on PATH: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "playwright-debug", "run", req.PlaywrightScript)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return &Response{
		PlaywrightStdout: stdout.String(),
		PlaywrightStderr: stderr.String(),
		PlaywrightExit:   commandErrorExit(err),
		Err:              err,
	}, nil
}
```
