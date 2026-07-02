# codex-tty agent-run Tests

Doc-style tests for `agent-run run --agent-runner codex-tty` and
`agent-run attach <session-id>`. Each run spawns an **adhoc per-run ptywrap
HTTP server** on a random `127.0.0.1:0` port, never a shared `agent-term serve`.
Session ids are printed to stderr as `codex-tty: session-N`; hidden attach ports
are recorded under a codex-specific registry in `AGENT_RUN_HOME`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` validates the runner id, starts the tty backend, blocks
  until the Codex process exits, and persists the resulting agent-run session.
  `attach` receives only a session id and searches supported tty registries
  deterministically.
- **Adhoc ptywrap server** — ephemeral HTTP listener on `127.0.0.1:<random>`; lives
  only while the parent `run` process is blocking; exposes attach over WebSocket.
- **PTY session** — interactive Codex CLI (or fake TUI hook) inside a pseudo-terminal;
  not a non-interactive JSON or `-p` mode.
- **Session registry** — JSON file mapping session id to `listen_addr`, `pid`, and
  `created_at`; provider-specific enough that codex-tty and grok-tty diagnostics are
  not ambiguous. When the same session id appears in multiple registry dirs, attach
  probes candidates and skips unreachable stale entries before choosing a live one.
- **Capture sidecar** — readonly scrollback poller that collects useful assistant
  text when no structured Codex sidecar stream is available. Codex scrollback capture
  is cleaned before it becomes stdout or an event, so terminal controls, TUI chrome,
  startup status, and prompt history do not leak into headless output.
- **Codex rollout transcript** — JSONL file under
  `<codex-home>/sessions/<YYYY>/<MM>/<DD>/rollout-*-<codex-session-id>.jsonl`.
  The Codex session id is learned from PTY scrollback resume text such as
  `codex resume <uuid>`, then the matching rollout file is tailed for assistant
  records while the PTY is still running.
- **Fake TUI** — `AGENT_RUN_CODEX_TTY_COMMAND` replaces the Codex executable for
  deterministic tests; it prints `CODEX_TTY_BANNER`, then reads the prompt and emits
  controlled output.
- **Real Codex CLI** — optional local binary resolved through `AGENT_RUNNER_CODEX_PATH`,
  `codex_cli_path`, or the normal fallback; real tests are labeled `codex`.

**Behaviors**

- `run --agent-runner codex-tty "prompt"` validates `codex-tty`, starts an adhoc
  ptywrap server, launches interactive Codex in the PTY, writes a registry entry,
  prints `codex-tty: session-N` to **stderr only**, waits for a Codex banner/readiness
  marker, injects `prompt + "\r"` so the prompt is submitted, blocks until Codex exits,
  discovers a Codex rollout transcript from resume scrollback when available, streams
  assistant JSONL records to stdout/events before PTY exit, falls back to cleaned
  scrollback when no transcript is found, and marks the agent-run session finished or
  errored through existing storage.
- `attach <id>` searches live tty registries, resolves the codex-tty registry entry,
  and attaches to the hidden ptywrap listener; stale same-id registry entries are
  skipped; unknown or expired ids exit non-zero with a clear error.
- Default suite uses the fake TUI hook and requires no real Codex installation.
  `real-codex/` leaves are skipped unless `--label codex` is requested.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/codex-tty/
├── DOCTEST.md
├── SETUP.md                           # build agent-run, fake TUI, registry + attach helpers
├── cli-edge/
│   └── accepts-codex-tty-runner/      # codex-tty passes validateRunner (not unknown)
├── run/
│   ├── prints-prefixed-id-stderr/     # stderr codex-tty: session-N; not on stdout
│   ├── waits-for-command-exit/        # blocks until hook exits 0
│   ├── creates-registry-entry/        # registry JSON + listen_addr while running
│   ├── captures-tui-output/           # fake TUI Response: hi → stdout + events.jsonl
│   ├── cleans-codex-scrollback/       # raw Codex TUI scrollback → useful lines only
│   ├── waits-for-banner/              # delayed CODEX_TTY_BANNER before prompt inject
│   ├── prompt-submits-on-enter/       # CR-only fake TUI → SUBMITTED:prompt (RED: bare \n)
│   └── fallback-scrollback-when-no-session/  # scrollback capture works without sidecar stream
├── session-jsonl-streaming/
│   ├── transcript-present/
│   │   ├── discovers-resume-id-agent-message/     # resume UUID -> rollout JSONL -> agent_message
│   │   ├── discovers-active-cwd-before-resume-footer/ # session_meta.cwd -> active rollout before footer
│   │   ├── ignores-stale-matching-cwd-transcript/ # current same-cwd rollout wins over stale transcript
│   │   ├── response-item-assistant-message/       # assistant response_item.message content streams
│   │   ├── streams-before-pty-exit/               # JSONL assistant message appears while PTY sleeps
│   │   ├── deduplicates-task-complete/            # task_complete duplicate is suppressed
│   │   └── ignores-function-call-output/          # tool output is diagnostic, not assistant stdout
│   └── transcript-absent/
│       └── falls-back-to-scrollback/              # no rollout file -> existing scrollback fallback
├── attach/
│   ├── connects-via-registry/         # background run → attach WS probe OK
│   ├── skips-stale-same-id-registry/  # dead grok session-1 does not shadow live codex session-1
│   └── missing-session-error/         # unknown id → exit 1, helpful stderr
├── help/
│   ├── lists-attach/                  # top-level help lists attach subcommand
│   └── lists-codex-tty-runner/        # runner help lists codex-tty backend
└── real-codex/                        # label: codex (skipped by default)
    ├── banner-appears/                # real codex banner detected before inject
    ├── attach-while-running/          # background run → attach connects, output visible
    ├── prompt-executes-not-stuck/     # real codex run ls exits or produces visible output
    └── visible-scrollback-output/     # real prompt produces visible captured scrollback
```

Parameter ranking (most → least significant):

1. **Invocation** — `run --agent-runner codex-tty` vs `attach <id>` vs `--help`
2. **Runner backend** — fake TUI (`AGENT_RUN_CODEX_TTY_COMMAND`) vs real `codex` (`label: codex`)
3. **Session lifecycle** — running (registry live) vs expired/missing
4. **TUI readiness** — immediate vs delayed `CODEX_TTY_BANNER`
5. **Codex transcript source** — rollout JSONL present vs scrollback fallback
6. **Codex JSONL record kind** — assistant message records vs duplicate/noise records
7. **Prompt submission** — carriage return submitted vs typed-but-not-submitted
8. **Output surface** — stderr session id vs stdout/events capture vs attach scrollback

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `cli-edge/accepts-codex-tty-runner` | `codex-tty` accepted by `validateRunner`; run starts (not unknown runner) |
| 2 | `run/prints-prefixed-id-stderr` | stderr matches `codex-tty: session-\d+`; stdout lacks prefix |
| 3 | `run/waits-for-command-exit` | Blocks until fake TUI exits; exit code 0 |
| 4 | `run/creates-registry-entry` | Registry JSON exists with `listen_addr` while run is active |
| 5 | `run/captures-tui-output` | Fake TUI echoes `Response: hi`; stdout and `events.jsonl` contain `hi` |
| 6 | `run/cleans-codex-scrollback` | Raw Codex TUI scrollback keeps result lines and strips controls/chrome/status |
| 7 | `run/waits-for-banner` | Delayed `CODEX_TTY_BANNER`; run succeeds (waits before prompt inject) |
| 8 | `run/prompt-submits-on-enter` | CR-only fake TUI; `run ls` → `SUBMITTED:run ls` not `UNSUBMITTED` |
| 9 | `attach/connects-via-registry` | Background run; `attach` probe resolves port from registry, WS OK |
| 10 | `attach/skips-stale-same-id-registry` | Dead grok-tty `session-1` entry is skipped; live codex-tty `session-1` attaches |
| 11 | `attach/missing-session-error` | Unknown/expired id → exit 1, stderr mentions session not found or expired |
| 12 | `run/fallback-scrollback-when-no-session` | No structured sidecar stream; scrollback fallback still emits assistant text |
| 13 | `help/lists-attach` | Top-level help lists `attach` |
| 14 | `help/lists-codex-tty-runner` | Run help lists `codex-tty` as a supported runner |
| 15 | `real-codex/banner-appears` | Real Codex banner detected; no banner timeout (`label: codex`) |
| 16 | `real-codex/prompt-executes-not-stuck` | Real Codex prompt exits or produces visible scrollback (`label: codex`) |
| 17 | `real-codex/attach-while-running` | Background real run → attach connects; Codex output visible (`label: codex`) |
| 18 | `real-codex/visible-scrollback-output` | Real prompt produces useful captured stdout/scrollback (`label: codex`) |
| 19 | `session-jsonl-streaming/transcript-present/discovers-resume-id-agent-message` | Resume UUID in scrollback selects matching rollout JSONL and streams `event_msg.agent_message` |
| 20 | `session-jsonl-streaming/transcript-present/discovers-active-cwd-before-resume-footer` | Matching `session_meta.cwd` selects the active rollout before a resume footer is printed |
| 21 | `session-jsonl-streaming/transcript-present/ignores-stale-matching-cwd-transcript` | Newest recent same-cwd rollout wins over an older stale transcript |
| 22 | `session-jsonl-streaming/transcript-present/response-item-assistant-message` | `response_item.message` with assistant `output_text` emits an assistant message |
| 23 | `session-jsonl-streaming/transcript-present/streams-before-pty-exit` | Assistant JSONL text reaches stdout before fake Codex exits |
| 24 | `session-jsonl-streaming/transcript-present/deduplicates-task-complete` | Duplicate `task_complete.last_agent_message` does not print the same final answer twice |
| 25 | `session-jsonl-streaming/transcript-present/ignores-function-call-output` | `function_call_output` text is not emitted as an assistant message by default |
| 26 | `session-jsonl-streaming/transcript-absent/falls-back-to-scrollback` | No rollout transcript keeps existing cleaned scrollback fallback |

## How to Run

```sh
# CI / default — fake TUI only (no label required)
doctest vet ./cmd/agent-run/tests/codex-tty
doctest test ./cmd/agent-run/tests/codex-tty
doctest test -v ./cmd/agent-run/tests/codex-tty/run/captures-tui-output
doctest test ./cmd/agent-run/tests/codex-tty/run/fallback-scrollback-when-no-session
doctest test ./cmd/agent-run/tests/codex-tty/session-jsonl-streaming

# Design / debug — real Codex (requires codex on PATH and usable auth)
doctest test --label codex ./cmd/agent-run/tests/codex-tty/real-codex
doctest test --label codex ./cmd/agent-run/tests/codex-tty
doctest test -v --label codex ./cmd/agent-run/tests/codex-tty/real-codex/attach-while-running
doctest test --label codex ./cmd/agent-run/tests/codex-tty/real-codex/visible-scrollback-output
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type CodexTTYRegistryEntry struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
	CreatedAt  string `json:"created_at"`
}

type Request struct {
	RepoRoot   string
	TempDir    string
	Home       string
	AgentRun   string
	Args       []string
	Env        []string
	CodexTTYCommand string // AGENT_RUN_CODEX_TTY_COMMAND; empty = production codex argv
	CodexTTYPrompt  string
	CodexTTYSessionID string
	CodexHome           string // CODEX_HOME; temp dir for fake rollout transcripts
	CodexTranscriptSessionID string
	CodexTranscriptPath      string
	CodexTranscriptSchedules []CodexTranscriptSchedule
	StreamProbeSubstring string
	StreamProbeTimeout   time.Duration
	SkipCodexSessionDir   bool // asserts scrollback-only behavior; no structured sidecar required
	Mode           string // "" | "registry-while-running" | "attach-probe" | "attach-cli-only-probe" | "attach-interactive-probe" | "attach-scrollback-probe" | "codex-jsonl-stream-probe"
	ExecTimeout    time.Duration
	BackgroundCmd    *exec.Cmd
	BackgroundStderr *bytes.Buffer
	BackgroundStdout *bytes.Buffer
	SkipFakeTUI      bool // real-codex: do not set AGENT_RUN_CODEX_TTY_COMMAND
}

type CodexTranscriptSchedule struct {
	Delay time.Duration
	Lines []string
}

type Response struct {
	Stdout           string
	Stderr           string
	ExitCode         int
	Err              error
	RegistryEntry    *CodexTTYRegistryEntry
	AttachProbeOK    bool
	AttachProbeErr   string
	AttachScrollback string
	EventsFilePath   string
	EventsFileLines  []string
	BackgroundStderr string
	BackgroundStdout string
	StreamProbeSeen       bool
	StreamProbeBeforeExit bool
	StreamProbeStdout     string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "registry-while-running":
		return runRegistryWhileRunning(t, req)
	case "attach-probe":
		return runAttachProbe(t, req)
	case "attach-cli-only-probe":
		return runAttachCLIOnlyProbe(t, req)
	case "attach-interactive-probe":
		return runAttachInteractiveProbe(t, req)
	case "attach-scrollback-probe":
		return runAttachScrollbackProbe(t, req)
	case "codex-jsonl-stream-probe":
		return runCodexJSONLStreamProbe(t, req)
	default:
		return runAgentRun(t, req, req.Args...)
	}
}
```
