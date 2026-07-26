# Grok-TTY Chat Tail — Incomplete Web Chat Fix Tests

Doc-style tests for the grok-tty web chat incompleteness bug: producer grok tail must
stay alive after initial `updates.jsonl` sync (`tailState.streamed`), consumer
`WatchEvents` must ignore session status, and web SSE must use page-lifetime watch.

# DSN (Domain Specific Notion)

**Participants**

- **Grok on-disk session** — `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/updates.jsonl`
  ACP stream (`user_message_chunk`, `agent_thought_chunk`, `tool_call`,
  `tool_call_update`, `agent_message_chunk`, `turn_completed`).
- **Producer (grok tail)** — `pkgs/agenttty` tails `updates.jsonl`, converts via
  `grok_session.Converter`, appends to `events.jsonl`. With `KeepTerminalAlive`
  (`--keep-tty`), tail lifetime must follow TTY/registry — **not** end when
  `waitForPersistentTurnRemote` returns on first sync (`tailState.streamed`).
- **Session store** — `AGENT_RUN_HOME/sessions/grok-tty/<id>/events.jsonl` canonical
  NDJSON timeline; `meta.json` status is metadata only.
- **WatchEvents (`pkgs/agentevents`)** — tails `events.jsonl` from byte offset until
  client `ctx` cancelled; must **not** stop when `meta.status != "running"`.
- **agent-run web SSE** — `GET .../events/stream` uses `WatchEvents`; chat page
  should open SSE on session mount for TTY runners (not gate on `status === 'running'`).
- **CLI `sessions --print`** — follow path uses `WatchEvents` until Ctrl+C regardless
  of session status.
- **llm-mock-run-grok** — chrome hook holds PTY; delayed `updates.jsonl` scheduler
  simulates tool-using turn race (`run ls and pwd`).
- **Test harness** — session-scoped binary build; isolated `AGENT_RUN_HOME`; ACP
  line builders; events.jsonl ordering probes; HTTP/SSE/CLI probes.

**Behaviors**

- **Producer P1** — keep-tty run seeds user+think+pending `tool_call`, then delayed
  `tool_call_update(completed)` + `agent_message_chunk` + `turn_completed` →
  `events.jsonl` has completed tool, assistant marker, `done` **after** assistant.
- **Producer P2** — same race but completion append arrives **after** initial batch
  already synced to `events.jsonl` (longer delay; assert pending tool visible first).
- **Consumer C1** — `WatchEvents` on `finished` session delivers newly appended line
  while watch ctx alive.
- **Consumer C2** — `sessions <runner>/<id> --print` on finished session tails
  appended line before timeout/Ctrl+C.
- **Web W1** — POST grok-tty session with delayed assistant; SSE delivers marker.
- **Web W2** — SSE stays connected after status `finished`; late append received.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/grok-tty-chat-tail/
├── DOCTEST.md
├── SETUP.md
├── producer/
│   ├── SETUP.md
│   └── keep-tty/
│       ├── SETUP.md
│       ├── delayed-tool-completion-in-events/     # P1: delayed tool completion + assistant
│       └── append-after-initial-stream-sync/      # P2: completion after streamed initial batch
├── consumer/
│   ├── SETUP.md
│   └── cli-follow-delivers-after-finished/        # C2: sessions --print tails finished session
└── web/
    ├── SETUP.md
    ├── sse-delivers-delayed-assistant/            # W1: web POST + delayed assistant in SSE
    └── sse-stays-open-after-finished/             # W2: SSE receives append after status finished
```

Sibling tree for C1 (separate `Run` contract):

```
pkgs/agentevents/tests/watchevents/
├── DOCTEST.md
├── SETUP.md
└── finished-status/
    ├── SETUP.md
    └── delivers-appended-line-while-watching/     # C1: WatchEvents ignores status gate
```

Parameter ranking (most → least significant):

1. **Layer** — producer tail lifetime vs consumer WatchEvents vs web SSE vs CLI follow.
2. **Session lifecycle** — keep-tty running tail vs finished-session consumer watch.
3. **Append timing** — immediate partial seed vs delayed completion after streamed sync.
4. **Transport** — `events.jsonl` file vs SSE vs CLI stdout.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `producer/keep-tty/delayed-tool-completion-in-events` | P1: delayed tool completion + assistant marker + ordered `done` in `events.jsonl` |
| 2 | `producer/keep-tty/append-after-initial-stream-sync` | P2: pending tool synced first; delayed completion still reaches `events.jsonl` |
| 3 | `consumer/cli-follow-delivers-after-finished` | C2: `sessions --print` on finished session delivers appended assistant line |
| 4 | `web/sse-delivers-delayed-assistant` | W1: web SSE delivers delayed assistant marker from grok-tty POST |
| 5 | `web/sse-stays-open-after-finished` | W2: SSE receives event appended after session status `finished` |
| 6 | `pkgs/agentevents/tests/watchevents/finished-status/delivers-appended-line-while-watching` | C1: `WatchEvents` delivers append on finished session while ctx alive |

Regression (cross-tree, not duplicated):

| # | External | Description |
|---|----------|-------------|
| P3 | `grok-discovery-race/keep-tty/delayed-session-streams` | Discovery not cancelled by chrome; marker still streams |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/grok-tty-chat-tail                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/grok-tty-chat-tail
doctest test --label-all ./cmd/agent-run/tests/grok-tty-chat-tail

# Vet + RED before implement
doctest vet ./cmd/agent-run/tests/grok-tty-chat-tail
doctest vet ./pkgs/agentevents/tests/watchevents
doctest test ./cmd/agent-run/tests/grok-tty-chat-tail
doctest test ./pkgs/agentevents/tests/watchevents

# Focused leaves
doctest test -v ./cmd/agent-run/tests/grok-tty-chat-tail/producer/keep-tty/delayed-tool-completion-in-events
doctest test -v ./cmd/agent-run/tests/grok-tty-chat-tail/producer/keep-tty/append-after-initial-stream-sync
doctest test -v ./cmd/agent-run/tests/grok-tty-chat-tail/web/sse-delivers-delayed-assistant
doctest test -v ./cmd/agent-run/tests/grok-tty-chat-tail/web/sse-stays-open-after-finished
doctest test -v ./pkgs/agentevents/tests/watchevents/finished-status/delivers-appended-line-while-watching

# P3 regression (must stay GREEN after fix)
doctest test ./cmd/agent-run/tests/grok-discovery-race/keep-tty/delayed-session-streams
```

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	LLMMockRunGrok string
	GrokHome       string
	Env            []string

	Mode     string // producer | sse | sse-finished-append | cli-follow
	Scenario string // leaf slug

	SessionID string
	Prompt    string
	Runner    string
	ChromeHoldSeconds int

	GrokSessionUUID      string
	GrokUpdatesPath      string
	GrokUpdatesSchedules []GrokUpdatesSchedule
	CompletionDelay      time.Duration

	WebToken   string
	WebBaseURL string
	WebCmd     *exec.Cmd
	webStderr  *bytes.Buffer

	SSEAfterOffset int64
	SSEMaxWait     time.Duration

	CLIArgs     []string
	ExecTimeout time.Duration
	Sidecar     func()
}

type GrokUpdatesSchedule struct {
	Delay       time.Duration
	UpdatesPath string
	Lines       []string
	OnFire      func()
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
	Elapsed  time.Duration

	EventsFilePath  string
	EventsFileLines []string
	EventsParsed    []map[string]any

	HasAssistantMarker bool
	HasCompletedTool   bool
	HasPendingToolFirst bool
	DoneAfterAssistant  bool

	SSEEvents []map[string]any
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.Mode {
	case "producer":
		return runProducerProbe(t, req)
	case "sse":
		return runWebDelayedAssistantSSE(t, req)
	case "sse-finished-append":
		return runSSEFinishedAppendProbe(t, req)
	case "cli-follow":
		return runCLIFollowProbe(t, req)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}
```