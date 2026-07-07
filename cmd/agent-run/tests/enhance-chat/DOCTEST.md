# Enhance Chat — Pure Session Events Tests

Doc-style tests for removing PTY scrollback fallback from TTY runner event emission.
Grok session binding emits explicit `think` progress and `error` events to
`events.jsonl`; the web chat SSE timeline and React UI render progress/error cards
instead of PTY chrome as assistant messages.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run web** — `POST /api/agent-run/sessions {runner: grok-tty}` with
  `--grok-home`, `--grok-tty-runner-binary`, `--agent-runner grok-tty` on `--port 0`.
- **llm-mock-run-grok** — built from `./agent/llm/llm-mock/llm-mock-run-grok`; shell
  hook via `LLM_MOCK_RUN_GROK_COMMAND` simulates grok TTY + optional `updates.jsonl` seed.
- **Grok on-disk session** — `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/updates.jsonl`
  is the authoritative ACP stream when binding succeeds.
- **TTY runner (`pkgs/agenttty`)** — before `DiscoverSession`, emits
  `ActionThink{Text: "Resolve session id..."}`; on failure emits
  `ActionError{Text: "Cannot resolve session id: " + err}`; never writes PTY scrollback
  or snapshot text to `events.jsonl`.
- **Session store** — `AGENT_RUN_HOME/sessions/grok-tty/<id>/events.jsonl` (NDJSON
  `types.AgentEvent`, single source of truth).
- **WatchEvents / SSE** — web `GET .../events/stream` tails `events.jsonl` via
  `pkgs/agentevents.WatchEvents`; delivers all event types including `think` and `error`.
- **Chat page (React)** — renders `think` / `tool_call` as progress cards
  (`data-testid="progress-card"`), `error` as error card (`data-testid="error-card"`),
  `message` rows as bubbles; terminal modal unchanged (PTY inspect on failure).
- **Test harness** — session-scoped `agent-run` + `llm-mock-run-grok` build; isolated
  `AGENT_RUN_HOME`; success hook seeds `updates.jsonl`; failure hook prints PTY chrome
  but omits grok session dir; HTTP/SSE/Playwright probes.

**Behaviors**

- **Binding success** — mock hook seeds `updates.jsonl` → discovery succeeds →
  `events.jsonl` contains user message, `think` "Resolve session id...", optional
  buffering think, protocol assistant message, `done`; no PTY chrome strings.
- **Binding failure** — empty/wrong `GROK_HOME` or hook skips session dir → discovery
  fails → `think` then `error` "Cannot resolve session id: …" then `done`; no assistant
  `message` with scrollback text (`hi`, box-drawing, banners).
- **SSE parity** — stream delivers `think` and `error` typed rows (not filtered to
  `message` only).
- **UI parity** — chat page shows progress card during/after resolve and error card on
  bind failure; terminal button still available for PTY inspect.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/enhance-chat/
├── DOCTEST.md
├── SETUP.md
├── events/
│   ├── SETUP.md
│   ├── binding-progress-emitted/       # think "Resolve session id..." in events.jsonl
│   ├── binding-failure-emits-error/    # error event, no assistant scrollback fallback
│   └── no-pty-chrome-in-events/        # no box-drawing/banner strings in events.jsonl
├── web/
│   ├── SETUP.md
│   └── sse-emits-think-and-error/      # SSE delivers think + error types on bind failure
└── frontend-ui/
    ├── SETUP.md
    ├── progress-card-on-resolve/       # Playwright: progress-card visible (success path)
    └── error-card-on-bind-failure/     # Playwright: error-card on bind failure
```

Parameter ranking (most → least significant):

1. **Assertion surface** — `events.jsonl` file vs SSE stream vs browser DOM.
2. **Binding outcome** — success (mock seeds updates) vs failure (no grok session dir).
3. **Event type** — `think` progress vs `error` vs absence of PTY chrome assistant text.
4. **Transport** — file read after finish vs live SSE vs Playwright poll.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `events/binding-progress-emitted` | Successful bind writes `think` "Resolve session id..." to `events.jsonl` |
| 2 | `events/binding-failure-emits-error` | Failed bind writes `error` "Cannot resolve session id: …"; no assistant fallback |
| 3 | `events/no-pty-chrome-in-events` | `events.jsonl` lacks PTY chrome substrings even when TUI printed them |
| 4 | `web/sse-emits-think-and-error` | SSE stream includes `think` and `error` event types on bind failure |
| 5 | `frontend-ui/progress-card-on-resolve` | Chat page renders `data-testid="progress-card"` for resolve think event |
| 6 | `frontend-ui/error-card-on-bind-failure` | Chat page renders `data-testid="error-card"` with resolve error text |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/enhance-chat
doctest test ./cmd/agent-run/tests/enhance-chat
doctest test -v ./cmd/agent-run/tests/enhance-chat/events
doctest test -v ./cmd/agent-run/tests/enhance-chat/web
doctest test --label ui-automation ./cmd/agent-run/tests/enhance-chat/frontend-ui
doctest test ./cmd/agent-run/tests/grok-tty/run/fallback-scrollback-when-no-session
doctest test ./cmd/agent-run/tests/web-cli-subset/events
```

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot            string
	TempDir             string
	Home                string
	AgentRun            string
	LLMMockRunGrok      string
	GrokHome            string
	GrokTTYRunnerBinary string
	Env                 []string

	Area            string // events | web | frontend-ui
	Action          string // leaf slug
	Mode            string // events | sse | ui
	BindingOutcome  string // success | failure

	WebToken   string
	WebBaseURL string
	WebCmd     *exec.Cmd
	webStderr  *bytes.Buffer

	Runner    string
	SessionID string
	Prompt    string

	GrokSessionUUID string

	SSEAfterOffset int64
	SSEMaxWait     time.Duration

	PlaywrightScript string
}

type Response struct {
	HTTPStatus int
	HTTPBody   string

	SSEEvents []map[string]any

	EventsFilePath  string
	EventsFileLines []string
	EventsParsed    []map[string]any

	PlaywrightStdout string
	PlaywrightStderr string
	PlaywrightExit   int

	Err error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "events", "":
		return runEventsProbe(t, req)
	case "sse":
		return runSSEProbe(t, req)
	case "ui":
		return runPlaywrightProbe(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}
```