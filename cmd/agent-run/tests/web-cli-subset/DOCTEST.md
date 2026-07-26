# agent-run web CLI subset refactor Tests

Doc-style tests for refactoring `agent-run web` into a strict subset of the CLI:
shared `AttachRelay`, `WatchEvents`, `SendToAgentSession`, and
`agenttty.ResolveByAgentSession` — no parallel web-only PTY proxy, send, or
phased event stream code.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run`, `attach`, `send` / `tty send`, `sessions --print`;
  headless runs use `StreamPhases: false` (same as `run --json`).
- **agent-run web server** — thin HTTP transport over the same libraries: POST
  sessions, POST messages, terminal status, terminal websocket, SSE events.
- **AttachRelay (`pkgs/ttywatch`)** — unified attach relay: dial terminal,
  consume handshake, relay output to sink, forward input/resize from sink.
  Sinks: `TTYAttachSink` (CLI stdin/stdout), `WebSocketAttachSink` (browser WS).
  Web terminal uses `attach_mode=attach` (write + resize), not snapshot proxy.
- **WatchEvents (`pkgs/agentevents`)** — tails `events.jsonl` via
  `logs.WatchLine` with byte-offset resume; stops when `meta.status != running`.
  Web SSE and CLI `sessions --print` both use this.
- **SendToAgentSession** — resolves agent session to live TTY via
  `agenttty.ResolveByAgentSession`, enqueues via `agentsend`, starts drainer.
  Web POST messages on live TTY and CLI `send` share this path.
- **Session store** — `AGENT_RUN_HOME/sessions/<runner>/<id>/meta.json` +
  `events.jsonl` (NDJSON `types.AgentEvent`, single source of truth).
- **TTY registry** — `<runner>-registry/<terminal_session_id>.json` with
  `listen_addr`, `pid`, `session_id`.
- **Send queue** — `send-queue/<runner>/<terminal_session_id>.jsonl`.
- **tty-watch CLI** — standalone attach client; must keep working after
  `AttachRelay` extraction (unchanged attach semantics).
- **Chat page / terminal modal** — React UI renders raw `AgentEvent` rows from
  `events.jsonl` (no phased coalescing); xterm.js terminal modal uses attach
  relay backend.
- **Test harness** — builds `agent-run`, `fake-codex`, `tty-watch`; isolated
  `AGENT_RUN_HOME`; fake ptywrap servers; stub-tty for send isolation;
  HTTP/websocket/SSE/CLI probes; `playwright-debug` for UI leaves.

**Behaviors**

- CLI `agent-run attach` and web terminal WS both route through `AttachRelay`
  with `attach_mode=attach`.
- Web runs do not emit phased assistant events (`phase` absent on stored rows).
- SSE streams raw `events.jsonl` lines as `data: <json>\n\n` via file tail, not
  poll loop; `after` byte offset matches `ReadEvents`.
- SSE emits all event types the CLI writes (not filtered to `ActionMessage` only).
- Web follow-up on live TTY enqueues to `agentsend` and returns accepted.
- Web `codex-tty` POST creates PTY via HeadlessRun; `terminal_session_id` stored;
  terminal stays available while running and after finish (`keep-tty`).
- Follow-up reuses the same PTY registry entry (no second registry id).
- Existing `web/tty-terminal/websocket-proxy` and `web/stream` contracts hold
  with CLI-parity event shape.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/web-cli-subset/
├── DOCTEST.md
├── SETUP.md
├── attach-relay/
│   ├── SETUP.md
│   ├── cli-attach-stdin-stdout-relay/
│   ├── web-ws-attach-relay-roundtrip-resize/
│   ├── web-ws-keyboard-reaches-pty/
│   └── tty-watch-attach-unchanged/
├── send/
│   ├── SETUP.md
│   ├── web-followup-enqueues-send-queue/
│   └── web-followup-accepted-delivered/
├── events/
│   ├── SETUP.md
│   ├── web-run-no-phased-assistant-events/
│   ├── sse-file-tail-appends-live/
│   ├── sse-after-offset-skips-prior/
│   ├── sse-emits-all-cli-event-types/
│   └── sessions-print-after-watchevents/
├── lifecycle/
│   ├── SETUP.md
│   ├── web-codex-tty-stores-terminal-session-id/
│   ├── terminal-available-running-and-finished/
│   └── follow-up-reuses-same-pty/
├── frontend-ui/
│   ├── SETUP.md
│   ├── chat-assistant-no-phased-coalescing/
│   └── terminal-modal-attach-web-session/
└── regression/
    ├── SETUP.md
    ├── websocket-proxy-roundtrip-still-works/
    └── sse-contract-cli-parity-shape/
```

Parameter ranking (most → least significant):

1. **Capability area** — attach relay vs send vs events vs lifecycle vs UI vs regression.
2. **Transport surface** — CLI vs web HTTP/WS/SSE vs tty-watch vs browser.
3. **Session state** — live TTY vs finished session vs running tail.
4. **Runner** — stub-tty (send/isolation) vs codex-tty (lifecycle) vs fake-codex (events).
5. **Specific operation** — round-trip IO, resize, queue enqueue, offset resume, etc.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `attach-relay/cli-attach-stdin-stdout-relay` | CLI attach uses `attach_mode=attach`; stdin reaches PTY, stdout relays output |
| 2 | `attach-relay/web-ws-attach-relay-roundtrip-resize` | Web terminal WS uses attach relay (not snapshot); IO + resize round-trip |
| 3 | `attach-relay/web-ws-keyboard-reaches-pty` | Web terminal WS allows write; keyboard bytes reach upstream PTY |
| 4 | `attach-relay/tty-watch-attach-unchanged` | `tty-watch attach` still forwards stdin after AttachRelay extraction |
| 5 | `send/web-followup-enqueues-send-queue` | Web POST messages on live TTY writes `send-queue/...jsonl` like CLI send |
| 6 | `send/web-followup-accepted-delivered` | Web follow-up returns accepted; message delivered via queue |
| 7 | `events/web-run-no-phased-assistant-events` | Web-created session stores assistant rows without `phase` field |
| 8 | `events/sse-file-tail-appends-live` | SSE receives events appended to `events.jsonl` while tailing (not poll-only) |
| 9 | `events/sse-after-offset-skips-prior` | SSE `after` byte offset skips prior events (ReadEvents compatible) |
| 10 | `events/sse-emits-all-cli-event-types` | SSE delivers non-message event types written by CLI (e.g. `done`) |
| 11 | `events/sessions-print-after-watchevents` | `sessions --print` still tails running session after WatchEvents extract |
| 12 | `lifecycle/web-codex-tty-stores-terminal-session-id` | Web codex-tty POST stores `terminal_session_id` from HeadlessRun |
| 13 | `lifecycle/terminal-available-running-and-finished` | Terminal status available while running and after finish (keep-tty) |
| 14 | `lifecycle/follow-up-reuses-same-pty` | Web follow-up preserves terminal mapping; no second registry entry |
| 15 | `frontend-ui/chat-assistant-no-phased-coalescing` | Chat shows one assistant bubble per stored message (no phased rows) |
| 16 | `frontend-ui/terminal-modal-attach-web-session` | Terminal modal attaches to web-created session PTY |
| 17 | `regression/websocket-proxy-roundtrip-still-works` | `web/tty-terminal/websocket-proxy` round-trip contract with attach mode |
| 18 | `regression/sse-contract-cli-parity-shape` | `web/stream` SSE contract with CLI-parity event shape (no `phase`) |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/web-cli-subset                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/web-cli-subset
doctest test --label-all ./cmd/agent-run/tests/web-cli-subset

doctest vet ./cmd/agent-run/tests/web-cli-subset
doctest test ./cmd/agent-run/tests/web-cli-subset
doctest test -v ./cmd/agent-run/tests/web-cli-subset/attach-relay
doctest test -v ./cmd/agent-run/tests/web-cli-subset/send
doctest test -v ./cmd/agent-run/tests/web-cli-subset/events
doctest test -v ./cmd/agent-run/tests/web-cli-subset/lifecycle
doctest test -v --label ui-automation ./cmd/agent-run/tests/web-cli-subset/frontend-ui
doctest test -v ./cmd/agent-run/tests/web-cli-subset/regression
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
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
	FakeCodex string
	LLMMockRunGrok      string
	GrokHome            string
	GrokTTYRunnerBinary string
	TTYWatch string
	Env      []string

	Area   string // attach-relay | send | events | lifecycle | frontend-ui | regression
	Action string // leaf slug

	Mode string // http | websocket | sse | cli | ui | tty-watch-attach

	WebToken   string
	WebPort    int
	WebBaseURL string
	WebCmd     *exec.Cmd
	webStdout  *bytes.Buffer
	webStderr  *bytes.Buffer

	Runner         string
	SessionID      string
	ChatSessionID  string
	TerminalSessionID string
	Prompt         string
	FollowUpPrompt string

	HTTPMethod string
	HTTPPath   string
	HTTPBody   string
	HTTPAuth   string

	RegistryListenAddr string
	RegistryPID        int
	RegistryServerURL  string
	RegistryTranscript string
	RegistryResize     string
	RegistryInputs     []string

	WSPath       string
	WSAuth       string
	WSInput      string
	WSResizeJSON string
	WSAttachMode string // expected upstream attach_mode query

	CLIArgs       []string
	CLIStdin      string
	ExecTimeout   time.Duration

	SSEAfterOffset int64
	SSEMaxWait     time.Duration

	StubScenarioJSON string
	StubScenarioPath string
	BackgroundCmd    *exec.Cmd

	PlaywrightScript string

	CodexTTYCommand string
	CodexTTYPrompt  string

	Sidecar func()
}

type Response struct {
	HTTPStatus int
	HTTPBody   string

	WSOutput      string
	WSResize      string
	WSError       string
	WSUpstreamURL string

	SSEEvents []map[string]any

	Stdout   string
	Stderr   string
	ExitCode int

	EventsFilePath  string
	EventsFileLines []string
	QueueFilePath   string
	QueueHasMsg     bool
	InjectedText    string

	TerminalSessionID string
	RegistryCount     int
	RegistryIDs       []string

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
	case "sse":
		return runSSE(t, req)
	case "cli":
		return runCLI(t, req)
	case "ui":
		return runPlaywright(t, req)
	case "tty-watch-attach":
		return runTTYWatchAttach(t, req)
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
```