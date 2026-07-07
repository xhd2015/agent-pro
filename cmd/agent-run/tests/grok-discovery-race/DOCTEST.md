# Grok-TTY Discovery Race Tests

Doc-style tests for the `grok-tty` + `--keep-tty` (web `KeepTerminalAlive`) bind race:
PTY scrollback chrome must not end `waitForPersistentTurnRemote` before grok session
discovery streams `updates.jsonl`. Repro harness mirrors
`script/debug/grok-tty-discovery-cancel`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --agent-runner grok-tty --keep-tty` with
  `--agent-runner-binary llm-mock-run-grok` and `--agent-runner-config-home`
  (empty `GROK_HOME` at start for failure leaves).
- **llm-mock-run-grok** — shell hook via `LLM_MOCK_RUN_GROK_COMMAND` prints
  real-like grok PTY chrome (worktree header, box drawing, `Grok Build` footer)
  and holds the PTY open with `sleep`.
- **Grok on-disk session** — `$GROK_HOME/sessions/<encoded-cwd>/<uuid>/updates.jsonl`
  may appear late (5–15s) during the PTY run; authoritative ACP stream once present.
- **TTY runner (`pkgs/agenttty`)** — with `KeepTerminalAlive`, grok turn completion
  must use `extraComplete()` (`tailState.streamed`) only, not scrollback
  `persistentTurnComplete` from chrome.
- **Discovery goroutine** — `DiscoverSession` polls until session dir matches cwd +
  prompt; canceled only when parent wait ends (must not cancel at ~1s from false
  scrollback completion).
- **Session store** — `AGENT_RUN_HOME/sessions/grok-tty/<id>/events.jsonl` records
  `think` "Resolve session id...", streamed assistant from `updates.jsonl`, or
  `error` "Cannot resolve session id: …".
- **Test harness** — session-scoped binary build; delayed session scheduler;
  real-like chrome hook; events.jsonl polling with think→error timing.

**Behaviors**

- **Delayed session success** — empty `GROK_HOME` at start; chrome PTY + `--keep-tty`;
  session dir materializes after 5–15s; discovery keeps polling; marker streams to
  `events.jsonl`; no `context canceled` error in the first 10s of discovery.
- **Chrome false-complete failure** — empty `GROK_HOME` + chrome; discovery must poll
  longer than ~3s; must not emit `Cannot resolve session id: context canceled` at
  ~1.2s from scrollback false completion (bug repro).

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/grok-discovery-race/
├── DOCTEST.md
├── SETUP.md
└── keep-tty/                                    # KeepTerminalAlive (web path)
    ├── SETUP.md
    ├── delayed-session-streams/                 # delayed updates.jsonl + chrome → marker in events
    └── chrome-wait-exceeds-discovery-window/    # empty GROK_HOME + chrome → wait >3s, no early cancel
```

Parameter ranking (most → least significant):

1. **Keep-tty path** — `--keep-tty` enables `waitForPersistentTurnRemote` + `extraComplete`.
2. **Grok session timing** — delayed dir (5–15s) vs permanently missing (empty home).
3. **PTY chrome** — real-like hook holds PTY and can falsely satisfy scrollback completion.
4. **Outcome** — streamed marker vs bind error after extended discovery window.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `keep-tty/delayed-session-streams` | Delayed `updates.jsonl` (8s) + chrome + `--keep-tty` → `DELAYED_SESSION_MARKER` in events; no early `context canceled` |
| 2 | `keep-tty/chrome-wait-exceeds-discovery-window` | Empty `GROK_HOME` + chrome + `--keep-tty` → discovery wait >3s without ~1s `context canceled` |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/grok-discovery-race
doctest test ./cmd/agent-run/tests/grok-discovery-race
doctest test -v ./cmd/agent-run/tests/grok-discovery-race/keep-tty/delayed-session-streams
doctest test -v ./cmd/agent-run/tests/grok-discovery-race/keep-tty/chrome-wait-exceeds-discovery-window
doctest test ./cmd/agent-run/tests/enhance-chat
doctest test ./cmd/agent-run/tests/grok-tty/run/discovery-polls-until-session-appears
go run ./script/debug/grok-tty-discovery-cancel -scenario=delayed-session
go run ./script/debug/grok-tty-discovery-cancel
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
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	LLMMockRunGrok string
	GrokHome       string
	Env            []string

	Scenario    string // delayed-session-streams | chrome-wait-exceeds-discovery-window
	SessionID   string
	Prompt      string
	ChromeHoldSeconds int

	GrokSessionUUID      string
	GrokUpdatesPath      string
	GrokUpdatesSchedules []GrokUpdatesSchedule

	ExecTimeout time.Duration
}

type GrokUpdatesSchedule struct {
	Delay       time.Duration
	UpdatesPath string
	OnFire      func()
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Elapsed  time.Duration
	Err      error

	EventsFilePath  string
	EventsFileLines []string
	EventsParsed    []map[string]any

	ThinkTimestamp    int64
	ErrorTimestamp    int64
	ThinkToErrorGap   time.Duration
	HasDelayedMarker  bool
	HasContextCancel  bool
	EarlyContextCancel bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runKeepTTYEventsProbe(t, req)
}
```