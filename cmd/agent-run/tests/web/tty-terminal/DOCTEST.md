# agent-run Web TTY Terminal Tests

Doc-style tests for `agent-run web` support for `codex-tty` and `grok-tty`
runners: runner discovery, terminal availability, authenticated websocket
attach, durable backend terminal reuse, read-only session runner metadata, and
the browser terminal modal.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web server** — serves the React chat UI and authenticated
  `/api/agent-run/*` JSON/SSE/websocket endpoints.
- **Runner catalog** — reports available agent runners to the web home page.
  It must include ordinary runners and tty runners (`codex-tty`, `grok-tty`).
- **Chat session store** — persists fixed session metadata under
  `AGENT_RUN_HOME/sessions/<runner>/<session_id>/meta.json` and timeline events
  under `events.jsonl`.
- **TTY registry** — live registry JSON under
  `AGENT_RUN_HOME/{codex-tty,grok-tty}-registry/<session_id>.json`, containing
  `session_id`, `listen_addr`, `pid`, and `created_at`.
- **Terminal resolver** — maps `(runner, session_id)` to a reachable tty
  registry entry only for tty runners. It never exposes raw filesystem paths or
  arbitrary upstream addresses to the browser.
- **Terminal websocket proxy** — accepts browser websocket attach inside the
  authenticated agent-run API namespace and forwards PTY output, browser input,
  paste bytes, Enter, and resize/control messages to the resolved ptywrap
  websocket.
- **Chat page** — displays transcript and session runner metadata. Existing
  session runner is read-only; terminal icon appears only when the terminal
  resolver reports an attachable live terminal.
- **Terminal modal** — attaches/detaches the browser view. Closing the modal
  does not stop the backend run and does not clear the chat transcript.
- **Test harness** — builds `agent-run`, starts `agent-run web` with isolated
  `AGENT_RUN_HOME`, writes deterministic session/registry fixtures, uses HTTP
  and websocket clients for backend tests, and uses `playwright-debug` for
  browser UI leaves.

**Behaviors**

- `GET /api/agent-run/runners` includes `codex-tty` and `grok-tty`.
- `GET /api/agent-run/sessions/{runner}/{session_id}/terminal` returns
  `available: true` only when the session is tty-backed and the registry entry
  is reachable.
- Missing or stale tty registries return a graceful not-available response.
- Non-tty runners never advertise attachable terminals.
- Terminal websocket attach requires the same auth as the rest of
  `/api/agent-run`; it proxies bytes and resize messages for the same
  `(runner, session_id)`.
- Re-entering the same chat page for a tty session reuses the existing registry
  entry and PTY session instead of creating a new backend terminal.
- During refreshes, already-loaded chat content remains visible.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web/tty-terminal/
├── DOCTEST.md
├── SETUP.md
├── backend-api/
│   ├── runner-list-includes-tty/
│   └── terminal-status/
│       ├── live-tty-registry/
│       ├── missing-registry/
│       ├── stale-registry/
│       └── non-tty-runner/
├── websocket-proxy/
│   ├── round-trip-io/
│   ├── resize-control/
│   └── auth-required/
├── durable-reuse/
│   └── same-session-reattach/
├── chat-runner/
│   └── read-only-existing-session/
└── frontend-ui/
    ├── runner-picker-shows-tty/
    ├── terminal-icon-available/
    ├── terminal-icon-absent-non-tty/
    ├── modal-attach-io/
    ├── modal-close-keeps-chat/
    ├── modal-reattach-after-navigation/
    └── loading-refresh-keeps-transcript/
```

Parameter ranking (most → least significant):

1. **Surface under test** — backend HTTP API vs websocket proxy vs persisted
   session lifecycle vs browser UI.
2. **Runner class** — tty (`codex-tty`, `grok-tty`) vs non-tty (`codex`,
   `grok`, `opencode`).
3. **Registry liveness** — live reachable registry vs missing registry vs stale
   unreachable registry.
4. **Attachment operation** — status probe vs websocket I/O vs resize/control
   forwarding vs authenticated rejection.
5. **Page state** — home runner picker vs existing chat page vs modal open/close
   vs navigation reattach vs refresh/loading.
6. **Specific tty runner** — representative leaves use `codex-tty`; helper
   assertions require the same behavior for both `codex-tty` and `grok-tty`
   where the runner catalog is checked.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `backend-api/runner-list-includes-tty` | Runners API includes `codex-tty` and `grok-tty` |
| 2 | `backend-api/terminal-status/live-tty-registry` | TTY session with reachable registry advertises `available: true` without leaking paths |
| 3 | `backend-api/terminal-status/missing-registry` | TTY session without registry returns graceful unavailable status |
| 4 | `backend-api/terminal-status/stale-registry` | TTY session with unreachable registry is not advertised as attachable |
| 5 | `backend-api/terminal-status/non-tty-runner` | Non-tty session does not advertise terminal attach |
| 6 | `websocket-proxy/round-trip-io` | Browser-side websocket receives PTY bytes and forwards input/Enter |
| 7 | `websocket-proxy/resize-control` | Browser resize/control message reaches the upstream PTY websocket |
| 8 | `websocket-proxy/auth-required` | Terminal websocket attach follows `/api/agent-run` auth rules |
| 9 | `durable-reuse/same-session-reattach` | Same web tty session resolves to the same registry-backed PTY after page navigation |
| 10 | `chat-runner/read-only-existing-session` | Existing chat page exposes runner as read-only metadata and follow-up uses route runner |
| 11 | `frontend-ui/runner-picker-shows-tty` | Home runner picker shows both tty runners |
| 12 | `frontend-ui/terminal-icon-available` | TTY chat with available terminal shows accessible terminal button |
| 13 | `frontend-ui/terminal-icon-absent-non-tty` | Non-tty chat has no attachable terminal affordance |
| 14 | `frontend-ui/modal-attach-io` | Terminal modal opens, renders terminal, receives output, sends keyboard input |
| 15 | `frontend-ui/modal-close-keeps-chat` | Closing modal preserves transcript and does not stop session |
| 16 | `frontend-ui/modal-reattach-after-navigation` | Reopening after leaving and returning attaches to same backend terminal |
| 17 | `frontend-ui/loading-refresh-keeps-transcript` | Refreshing terminal/session state leaves existing transcript visible |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web/tty-terminal                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web/tty-terminal
doctest test --label-all ./cmd/agent-run/tests/web/tty-terminal

doctest vet ./cmd/agent-run/tests/web/tty-terminal
doctest test ./cmd/agent-run/tests/web/tty-terminal
doctest test -v ./cmd/agent-run/tests/web/tty-terminal/backend-api
doctest test -v ./cmd/agent-run/tests/web/tty-terminal/websocket-proxy
doctest test -v ./cmd/agent-run/tests/web/tty-terminal/durable-reuse
doctest test -v ./cmd/agent-run/tests/web/tty-terminal/chat-runner
doctest test -v --label ui-automation ./cmd/agent-run/tests/web/tty-terminal/frontend-ui
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
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Env      []string

	Mode string // "http" | "websocket" | "ui"

	WebToken     string
	WebPort      int
	WebBaseURL   string
	WebCmd       *exec.Cmd
	webStdout    *bytes.Buffer
	webStderr    *bytes.Buffer

	Runner    string
	SessionID string
	Prompt    string

	HTTPMethod string
	HTTPPath   string
	HTTPBody   string
	HTTPAuth   string

	RegistryListenAddr string
	RegistryPID        int
	RegistryServerURL  string
	RegistryWSPath     string
	RegistryTranscript string
	RegistryResize     string

	WSPath       string
	WSAuth       string
	WSInput      string
	WSResizeJSON string

	PlaywrightScript string
}

type Response struct {
	HTTPStatus int
	HTTPBody   string

	WSOutput string
	WSResize string
	WSError   string

	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int

	Err error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Mode {
	case "http", "":
		return runHTTP(t, req)
	case "websocket":
		return runTerminalWebSocket(t, req)
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

func runCommand(ctx context.Context, name string, args ...string) (string, string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), commandErrorExit(err), err
}

func drainAndClose(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}
```
