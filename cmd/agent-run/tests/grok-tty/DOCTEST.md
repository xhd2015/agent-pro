# grok-tty agent-run Tests

Doc-style tests for `agent-run run --agent-runner grok-tty` and `agent-run attach
<session-id>`. Each run spawns an **adhoc per-run ptywrap HTTP server** on a random
`127.0.0.1:0` port (no `agent-term serve`). Session ids are printed to stderr as
`grok-tty: session-N`; hidden ports are recorded under `AGENT_RUN_HOME/grok-tty-registry/`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` blocks until the grok-tty session exits; `attach` looks
  up a live session by id and opens an interactive WebSocket to the hidden port.
- **Adhoc ptywrap server** — ephemeral HTTP listener on `127.0.0.1:<random>`; lives
  only while the parent `run` process is blocking; registers REST + WS handlers via
  `ptywrap.RegisterAPIWithManager`.
- **PTY session** — interactive grok (or fake TUI) inside a pseudo-terminal; not
  piped `-p` / streaming-json mode.
- **Session registry** — JSON files at `AGENT_RUN_HOME/grok-tty-registry/<id>.json`
  mapping session id → `listen_addr`, `pid`, `created_at`; removed when run exits.
- **Capture sidecar** — in-process readonly scrollback poller; accumulates PTY output
  while user attach may run concurrently (multiplex).
- **Grok on-disk session** — under `$GROK_HOME/sessions/<encoded-cwd>/<grok-uuid>/`;
  `summary.json` carries `info.cwd`; `updates.jsonl` is the authoritative ACP stream
  (`user_message_chunk`, `agent_thought_chunk`, `tool_call`, `tool_call_update`,
  `agent_message_chunk`).
- **Live updates tailer** — after prompt inject, discovers the matching grok session dir,
  tails `updates.jsonl`, converts ACP lines to `AgentEvent`s, emits formatted stdout and
  appends `AGENT_RUN_HOME/sessions/grok-tty/.../events.jsonl` incrementally.
- **Fake TUI** — `AGENT_RUN_GROK_TTY_COMMAND` replaces `grok` argv for deterministic
  tests; must print `GROK_TTY_BANNER` before prompt, then read stdin and echo a
  known response.

**Behaviors**

- `run --agent-runner grok-tty "prompt"` starts adhoc server, creates PTY session,
  writes registry entry, prints `grok-tty: session-N` on **stderr only**, waits for
  banner in scrollback, injects `prompt + "\r"`, discovers grok session dir under
  `GROK_HOME` (polls until found or PTY ends — no fixed 30s cap), prints
  `grok-tty: grok session <uuid>` and `grok-tty: grok updates <path>` on stderr
  once discovered, tails `updates.jsonl` and streams formatted events to stdout during
  the run, blocks until PTY exits, persists `runner_session_id` (grok UUID) in
  `meta.json`, tears down server and registry file. Scrollback extraction is
  **fallback only** when discovery fails (stderr warning + end-of-run heuristic).
- `attach <id>` reads registry → `ptyclient.NewClient("http://"+listen_addr)` →
  `Attach` with snapshot; fails with clear error when registry missing or expired.
- Default suite uses fake TUI hook; `real-grok/` leaves require real `grok` on PATH
  (`label: grok`).

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/grok-tty/
├── DOCTEST.md
├── SETUP.md                           # build agent-run, fake TUI, registry + attach helpers
├── cli-edge/
│   └── accepts-grok-tty-runner/       # grok-tty passes validateRunner (not unknown)
├── run/
│   ├── prints-prefixed-id-stderr/     # stderr grok-tty: session-N; not on stdout
│   ├── waits-for-command-exit/        # blocks until hook exits 0
│   ├── creates-registry-entry/        # registry JSON + listen_addr while running
│   ├── captures-tui-output/           # fake TUI Response: hi → stdout + events.jsonl
│   ├── waits-for-banner/              # delayed GROK_TTY_BANNER before prompt inject
│   ├── prompt-submits-on-enter/       # CR-only fake TUI → SUBMITTED:prompt (RED: bare \n)
│   ├── streams-events-before-exit/    # tail updates.jsonl → stdout before fake TUI exits
│   ├── streams-second-turn-after-completed/  # turn 2 stdout after turn_completed (primary bug)
│   ├── discovers-session-by-cwd-and-prompt/  # two grok dirs; prompt match picks correct tail
│   ├── persists-multiple-event-types/ # events.jsonl has user + tool + assistant (+ think)
│   ├── stores-grok-session-id/        # meta.json runner_session_id == grok UUID
│   ├── fallback-scrollback-when-no-session/  # no GROK_HOME dir → warn + scrollback fallback
│   ├── prints-grok-session-on-stderr/   # stderr grok session uuid + updates.jsonl path after discovery
│   ├── stderr-grok-session-before-stdout-stream/  # stderr grok lines before stdout stream marker
│   ├── discovery-polls-until-session-appears/  # delayed session dir (5s) still discovered + streams
│   ├── stdout-streams-formatted-events/  # stdout 💬 user, ⚡ tool, assistant from updates.jsonl
│   ├── rejects-prior-session-same-prompt/  # old "run ls" session ignored; new session tailed
│   ├── keep-tty-registry-persists/       # after --keep-tty run, registry file persists
│   └── keep-tty-terminal-stays-alive/    # after --keep-tty run, ptywrap server still reachable
├── attach/
│   ├── connects-via-registry/         # background run → attach WS probe OK
│   └── missing-session-error/         # unknown id → exit 1, helpful stderr
├── snapshot/
│   └── grok-mock-run-post-turn/       # llm-mock-run-grok + snapshot shows grok UI, not status-bar only (RED)
├── help/
│   ├── lists-attach/                  # --help lists attach subcommand
│   └── lists-send/                    # --help lists send subcommand
└── real-grok/                         # label: grok (skipped by default)
    ├── banner-appears/                # real grok banner detected before inject
    ├── run-simple-prompt/             # "say hi" → exit 0, captured output non-empty
    ├── prompt-executes-not-stuck/     # real grok run ls → listing in scrollback
    ├── attach-while-running/          # background run → attach connects, output visible
    ├── streams-during-run/            # run ls → stdout non-empty before 60s timeout
    └── prints-grok-session-stderr/    # stderr grok session + path; stdout non-empty (`label: grok`)
```

Parameter ranking (most → least significant):

1. **Invocation** — `run --agent-runner grok-tty` vs `attach <id>` vs `--help`
2. **Runner backend** — fake TUI (`AGENT_RUN_GROK_TTY_COMMAND`) vs real `grok` (`label: grok`)
3. **Grok session source** — live `updates.jsonl` tail vs scrollback fallback (no session dir)
4. **Session lifecycle** — running (registry live) vs expired/missing
5. **TUI readiness** — immediate vs delayed `GROK_TTY_BANNER`
6. **Output surface** — stderr session id vs stdout/events capture vs attach scrollback

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `cli-edge/accepts-grok-tty-runner` | `grok-tty` accepted by `validateRunner`; run starts (not unknown runner) |
| 2 | `run/prints-prefixed-id-stderr` | stderr matches `grok-tty: session-\d+`; stdout lacks prefix |
| 3 | `run/waits-for-command-exit` | Blocks until fake TUI exits; exit code 0 |
| 4 | `run/creates-registry-entry` | Registry JSON exists with `listen_addr` while run is active |
| 5 | `run/captures-tui-output` | Fake TUI echoes `Response: hi`; stdout and `events.jsonl` contain `hi` |
| 6 | `run/waits-for-banner` | Delayed `GROK_TTY_BANNER`; run succeeds (waits before prompt inject) |
| 7 | `run/prompt-submits-on-enter` | CR-only fake TUI; `run ls` → `SUBMITTED:run ls` not `UNSUBMITTED` |
| 8 | `attach/connects-via-registry` | Background run; `attach` probe resolves port from registry, WS OK |
| 9 | `attach/missing-session-error` | Unknown/expired id → exit 1, stderr mentions session not found or expired |
| 10 | `help/lists-attach` | `--help` lists `attach` subcommand |
| 10a | `help/lists-send` | `--help` lists `send` subcommand |
| 11 | `real-grok/banner-appears` | Real grok banner detected; no `banner not detected` error (`label: grok`) |
| 12 | `real-grok/run-simple-prompt` | `run --agent-runner grok-tty "say hi"` exit 0; output/events non-empty (`label: grok`) |
| 13 | `real-grok/prompt-executes-not-stuck` | Real grok `run ls`; scrollback has `total`/`drwx` listing (`label: grok`) |
| 14 | `real-grok/attach-while-running` | Background real run → attach connects; grok output visible (`label: grok`) |
| 15 | `run/streams-events-before-exit` | Temp `GROK_HOME` + synthetic `updates.jsonl`; stdout has streamed marker before fake TUI exits |
| 15a | `run/streams-second-turn-after-completed` | Turn 1 `turn_completed` seeded; turn 2 marker on stdout before fake TUI exits |
| 16 | `run/discovers-session-by-cwd-and-prompt` | Two grok session dirs; prompt `run ls` selects correct session to tail |
| 17 | `run/persists-multiple-event-types` | Streamed `events.jsonl` has user + tool_call + assistant (+ think) events |
| 18 | `run/stores-grok-session-id` | `meta.json` `runner_session_id` equals discovered grok UUID |
| 19 | `run/fallback-scrollback-when-no-session` | No grok session dir → stderr warns; scrollback fallback still emits assistant text |
| 20 | `real-grok/streams-during-run` | `run ls`; stdout non-empty before 60s timeout (`label: grok`) |
| 21 | `run/prints-grok-session-on-stderr` | Fake `GROK_HOME` session; stderr contains `grok session <uuid>` and `grok updates <path>` after discovery |
| 22 | `run/stderr-grok-session-before-stdout-stream` | Stderr grok session lines appear before stdout stream marker (ordering) |
| 23 | `run/discovery-polls-until-session-appears` | Session dir created 5s after prompt inject; discovery still finds it and streams |
| 24 | `run/stdout-streams-formatted-events` | Stdout contains formatted user (`💬`) and tool/assistant lines from synthetic `updates.jsonl` |
| 25 | `real-grok/prints-grok-session-stderr` | `label: grok`; stderr has `grok session` + path; stdout non-empty (`label: grok`) |
| 26 | `run/rejects-prior-session-same-prompt` | Prior session same prompt + old `created_at` ignored; new session marker streamed |
| 27 | `run/keep-tty-registry-persists` | `run --keep-tty` → registry entry persists after run exit |
| 28 | `run/keep-tty-terminal-stays-alive` | `run --keep-tty` → ptywrap listen_addr still reachable after run exit |
| 29 | `snapshot/grok-mock-run-post-turn` | `agent-run snapshot` after llm-mock grok turn shows prompt/UI, not status-bar only (RED) |

## How to Run

```sh
# CI / default — fake TUI only (no label required)
doctest vet ./cmd/agent-run/tests/grok-tty
doctest test ./cmd/agent-run/tests/grok-tty
doctest test -v ./cmd/agent-run/tests/grok-tty/run/captures-tui-output
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-events-before-exit
doctest test ./cmd/agent-run/tests/grok-tty/run/streams-second-turn-after-completed
doctest test ./cmd/agent-run/tests/grok-tty/run/prints-grok-session-on-stderr
doctest test ./cmd/agent-run/tests/grok-tty/run/discovery-polls-until-session-appears
doctest test ./cmd/agent-run/tests/grok-tty/run/keep-tty-registry-persists

# Design / debug — real grok (requires grok on PATH)
doctest test --label grok ./cmd/agent-run/tests/grok-tty/real-grok
doctest test --label grok ./cmd/agent-run/tests/grok-tty
doctest test -v --label grok ./cmd/agent-run/tests/grok-tty/real-grok/attach-while-running
doctest test --label grok ./cmd/agent-run/tests/grok-tty/real-grok/streams-during-run
doctest test --label grok ./cmd/agent-run/tests/grok-tty/real-grok/prints-grok-session-stderr
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

type GrokTTYRegistryEntry struct {
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
	GrokTTYCommand string // AGENT_RUN_GROK_TTY_COMMAND; empty = production grok argv
	GrokTTYPrompt  string
	GrokTTYSessionID string
	GrokHome           string // GROK_HOME; temp dir for fake grok session tree
	GrokSessionUUID    string // AGENT_RUN_GROK_TTY_GROK_SESSION_ID hook (skip discovery)
	GrokUpdatesPath    string // primary updates.jsonl path for scheduled appends
	GrokUpdatesSchedules []GrokUpdatesSchedule
	StreamProbeSubstring string
	StreamProbeTimeout   time.Duration
	SkipGrokSessionDir   bool // fallback test: do not seed GROK_HOME session dirs
	LLMMockRunGrok         string
	AgentRunnerBinary      string
	AgentRunnerConfigHome  string
	KeepTTY                bool
	SnapshotReadyMarker    string
	SnapshotDelay          time.Duration
	Mode           string // "" | "registry-while-running" | "attach-probe" | "attach-interactive-probe" | "attach-scrollback-probe" | "stream-probe" | "snapshot-probe"
	ExecTimeout    time.Duration
	BackgroundCmd    *exec.Cmd
	BackgroundStderr *bytes.Buffer
	BackgroundStdout *bytes.Buffer
	SkipFakeTUI      bool // real-grok: do not set AGENT_RUN_GROK_TTY_COMMAND
}

type GrokUpdatesSchedule struct {
	Delay       time.Duration
	Lines       []string
	UpdatesPath string
	OnFire      func() // optional: create session dirs or append to multiple paths
}

type Response struct {
	Stdout           string
	Stderr           string
	ExitCode         int
	Err              error
	RegistryEntry    *GrokTTYRegistryEntry
	AttachProbeOK    bool
	AttachProbeErr   string
	AttachScrollback string
	EventsFilePath   string
	EventsFileLines  []string
	BackgroundStderr string
	BackgroundStdout string
	StreamProbeSeen        bool
	StreamProbeBeforeExit  bool
	StreamProbeStdout      string
	MetaJSONPath           string
	MetaRunnerSessionID    string
	SnapshotStdout         string
	SnapshotExitCode       int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "registry-while-running":
		return runRegistryWhileRunning(t, req)
	case "attach-probe":
		return runAttachProbe(t, req)
	case "attach-interactive-probe":
		return runAttachInteractiveProbe(t, req)
	case "attach-scrollback-probe":
		return runAttachScrollbackProbe(t, req)
	case "stream-probe":
		return runStreamProbe(t, req)
	case "snapshot-probe":
		return runSnapshotProbe(t, req)
	default:
		return runAgentRun(t, req, req.Args...)
	}
}
```