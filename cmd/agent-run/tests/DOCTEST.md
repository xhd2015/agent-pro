# agent-run CLI Tests

Doc-style tests for `cmd/agent-run` — the user-facing agent runner with TUI,
headless `run`, localhost `web` API, and session storage under `AGENT_RUN_HOME`.

# DSN (Domain Specific Notion)

`agent-run` is the CLI entry. Runner backends come from `agent/cli/registry`
(`codex`, `opencode`, `pi`, `crush`, `grok`, `grok-tty`, `codex-tty`, `cursor`,
`fake-codex`, …). Tests use `fake-codex` for deterministic NDJSON output;
`grok-tty` and `codex-tty` tests live in nested trees with fake TUI hooks.

```
agent-run [--agent-runner RUNNER] [--help]
agent-run web [--port PORT] [--dev] [--token TOKEN] [--no-open] [--agent-runner RUNNER]
agent-run run "prompt" [--json] [--model M] [--session ID] [--auto-session-id] [--agent-runner RUNNER]
agent-run attach <session-id>
agent-run sessions [--json]
agent-run sessions <runner>/<session_id> --print
agent-run status
```

**Web mode** binds `127.0.0.1` only. Default port **8192**. Static SPA assets are
served without auth. API auth depends on `--token`:

| CLI | API auth | Startup stderr |
|-----|----------|----------------|
| (no `--token`) | Open — `/api/agent-run/*` without Bearer | Warning to use `--token` or `--token auto` |
| `--token auto` | Bearer required; random hex generated | `agent-run web token: <hex>` |
| `--token <value>` | Bearer required; value as-is | No token line; value persisted to `auth.token` |

**Storage** (`pkgs/agentstorage`, not agent-hub):

```
AGENT_RUN_HOME/                    # default ~/.agent-run/, temp dir in tests
  config.json
  auth.token
  sessions/<runner>/<session_id>/
    meta.json
    events.jsonl                   # NDJSON types.AgentEvent lines
    messages.jsonl
```

**`run --json`** streams NDJSON `types.AgentEvent` lines to stdout (one per line,
flushed). The same events append to `events.jsonl`. Without `--json`, output is
human-readable via `agent/event/print`.

**`sessions <runner>/<session_id> --print`** loads `meta.json` and `events.jsonl`
from the file store, formats events with `print.FormatState` (trace-style header,
numbered lines, message bodies). When `meta.status` is `running`, the CLI prints
existing events then tails `events.jsonl` until status is no longer `running`.
Missing sessions exit 1; a session positional without `--print` is rejected.

Each test sets `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")` and
builds `agent-run` + `fake-codex` binaries into the temp `bin/` directory.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/
├── DOCTEST.md
├── SETUP.md
├── cli-edge/                        Subcommand: invalid CLI input
│   ├── unknown-subcommand/          exit 1
│   └── unknown-agent-runner/        run with bad runner, stderr mentions unknown
├── help/
│   └── top-level/                   lists web, run, sessions, status, --agent-runner
├── run/
│   ├── json-fake-codex-hi/          run --json --agent-runner fake-codex "hi"
│   ├── events-persisted-to-home/    stdout NDJSON lines match events.jsonl
│   ├── human-readable-no-json/      without --json, stdout not all JSON lines
│   ├── auto-session-id/             nested: --auto-session-id + same-id TTY policy (see auto-session-id/DOCTEST.md)
│   ├── agent-runner-binary/         nested: --agent-runner-binary SPEC (see agent-runner-binary/DOCTEST.md)
│   └── agent-runner-config-home/    nested: --agent-runner-config-home PATH (see agent-runner-config-home/DOCTEST.md)
├── web/                             split: token mode (omit | explicit | auto)
│   ├── timeline/                    session detail events (role, timestamps)
│   │   ├── session-detail-includes-user-prompt/
│   │   ├── follow-up-message-includes-user-prompt/
│   │   ├── message-events-include-timestamp/
│   │   ├── assistant-message-includes-timestamp/
│   │   ├── continuation/
│   │   │   └── follow-up-agent-recalls-first-message/
│   │   └── streaming/
│   │       ├── streaming-message-phases-emitted/
│   │       └── assistant-phases-share-stable-id/
│   ├── stream/                      SSE tail of events.jsonl
│   │   ├── sse-delivers-new-events/
│   │   └── sse-after-offset-skips-prior/
│   ├── process-output/              web process must not leak agent UI to terminal
│   │   ├── web-stdout-silent-on-agent-run/
│   │   ├── startup-listen-line-newline/
│   │   ├── startup-stderr-no-leading-blank-line/
│   │   ├── startup-stderr-no-trailing-whitespace/
│   │   └── startup-auto-no-leading-blank-line/
│   ├── workspace/                   server cwd exposed via status + session meta
│   │   ├── status-includes-workspace/
│   │   └── session-meta-includes-workspace/
│   ├── no-token-health-200/         no --token; GET health without auth → 200
│   ├── no-token-startup-warning/    stderr warns about --token
│   ├── token-auto-generates-and-requires-auth/
│   ├── auth-missing-token-401/      --token test; no Bearer → 401
│   ├── auth-wrong-token-401/        wrong Bearer → 401
│   ├── auth-valid-bearer-200/       Bearer test → 200
│   ├── health-port-zero-starts/     --port 0 + explicit token
│   ├── default-port-8192/           default port + explicit token
│   └── grok-mock-config/            nested: --grok-home + --grok-tty-runner-binary (see grok-mock-config/DOCTEST.md)
├── home/
│   └── agent-run-home-isolation/    no files outside AGENT_RUN_HOME after run
├── sessions/                        split: list vs print
│   ├── list/
│   │   └── json-empty/              sessions --json when empty
│   └── print/                       sessions <runner>/<id> --print
│       ├── finished-with-events/    status=finished; formatted trace stdout
│       ├── no-events/               meta only; "(no events yet)" + Done footer
│       ├── unknown-session/         missing session → exit 1
│       ├── missing-print-flag/      session ref without --print → exit 1
│       ├── malformed-session-ref/   positional without slash → exit 1
│       └── follow-running-appends/  status=running; tail until finished
├── status/
│   └── exits-zero/                  status exits 0
├── grok-tty/                        nested root: grok-tty runner + attach (see grok-tty/DOCTEST.md)
    ├── cli-edge/accepts-grok-tty-runner/
    ├── run/                         stderr id, registry, banner wait, capture
    ├── attach/                      registry lookup + WS attach
    ├── help/lists-attach/
    └── real-grok/                   label: grok — real grok CLI on PATH
└── codex-tty/                       nested root: codex-tty runner + attach (see codex-tty/DOCTEST.md)
    ├── cli-edge/accepts-codex-tty-runner/
    ├── run/                         stderr id, registry, banner wait, scrollback capture
    ├── attach/                      registry lookup + WS attach
    ├── help/                        attach + codex-tty runner docs
    └── real-codex/                  label: codex — real Codex CLI on PATH
└── tty/                             nested root: tty status/attach/send subcommands (see tty/DOCTEST.md)
    ├── help/lists-subcommands/      tty --help lists status, attach, send
    ├── status/                      registry read + screen detection
    ├── attach/                      tty attach + shortcut alias
    ├── attach-shortcut/             agent-run attach delegates to tty attach
    └── send/                        WS prompt inject + response capture
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `cli-edge/unknown-subcommand` | Unknown subcommand exits 1 |
| 2 | `cli-edge/unknown-agent-runner` | `run --agent-runner bogus` exits 1, stderr mentions unknown |
| 3 | `help/top-level` | `--help` lists web, run, sessions, status, `--agent-runner` |
| 4 | `run/json-fake-codex-hi` | `run --json --agent-runner fake-codex "hi"` → valid NDJSON, last event type `done` |
| 5 | `run/events-persisted-to-home` | Same run; `events.jsonl` under home matches stdout lines |
| 6 | `run/human-readable-no-json` | `run` without `--json` uses human-readable print, not all JSON lines |
| 40 | `run/agent-runner-binary/fake-binary-receives-argv` | `--agent-runner-binary` fake script; stderr argv includes grok-tty defaults |
| 41 | `run/agent-runner-binary/inner-model-wins` | Binary spec `--model inner` beats CLI `--model outer` |
| 42 | `run/agent-runner-binary/llm-mock-auto-home` | `llm-mock-run-grok` auto-provisions grok home; discovery streams (no scrollback pollution) |
| 43 | `run/agent-runner-config-home/discovers-session` | `--agent-runner-config-home` + seeded `updates.jsonl` → stderr grok session |
| 44 | `run/agent-runner-config-home/child-env-grok-home-only` | Child env has `GROK_HOME=` only; no `AGENT_RUNNER_CONFIG_HOME=` |
| 45 | `web/grok-mock-config/uses-mock-binary-and-home` | Web `--grok-home` + `--grok-tty-runner-binary`; POST grok-tty uses mock + shared home |
| 7 | `web/no-token-health-200` | `web --port 0` (no `--token`); GET health without auth → 200; no `auth.token` |
| 8 | `web/no-token-startup-warning` | Same startup; stderr mentions `--token` |
| 9 | `web/token-auto-generates-and-requires-auth` | `--token auto`; stderr prints token; no auth → 401; Bearer printed → 200 |
| 10 | `web/auth-missing-token-401` | `web --token test --port 0`; GET health without auth → 401 |
| 11 | `web/auth-wrong-token-401` | GET health with wrong Bearer → 401 |
| 12 | `web/auth-valid-bearer-200` | GET health with `Bearer test` → 200 |
| 13 | `web/health-port-zero-starts` | `web --port 0 --token test` starts and health responds |
| 14 | `web/default-port-8192` | `web --token test` without `--port` binds 8192 |
| 15 | `home/agent-run-home-isolation` | After `run`, all writes stay under `AGENT_RUN_HOME` |
| 16 | `sessions/list/json-empty` | `sessions --json` on empty store returns valid empty JSON |
| 22 | `sessions/print/finished-with-events` | `--print` on finished session with events → trace header, message text, Done footer |
| 23 | `sessions/print/no-events` | `--print` with meta only → `(no events yet)` and `Done (session finished)` |
| 24 | `sessions/print/unknown-session` | unknown runner/id → exit 1, stderr mentions session |
| 25 | `sessions/print/missing-print-flag` | `sessions runner/id` without `--print` → exit 1 |
| 26 | `sessions/print/malformed-session-ref` | `sessions noslash --print` → exit 1 |
| 27 | `sessions/print/follow-running-appends` | running session follows new events until status finished |
| 17 | `status/exits-zero` | `status` exits 0 |
| 18 | `web/timeline/session-detail-includes-user-prompt` | POST session → GET detail events include `role=user` prompt |
| 19 | `web/timeline/follow-up-message-includes-user-prompt` | POST `.../messages` → GET detail includes user follow-up |
| 28 | `web/timeline/message-events-include-timestamp` | POST session → user `message` event has `timestamp` > 0 |
| 29 | `web/timeline/assistant-message-includes-timestamp` | After fake-codex run → assistant `message` event has `timestamp` > 0 |
| 30 | `web/process-output/web-stdout-silent-on-agent-run` | Background run → web process streams lack `💬` / `[done]` agent print |
| 31 | `web/process-output/startup-listen-line-newline` | No `--token` → listen URL line ends with newline (own line) |
| 37 | `web/process-output/startup-stderr-no-leading-blank-line` | No `--token` → stderr does not start with `\n`; first bytes `no API token` |
| 38 | `web/process-output/startup-stderr-no-trailing-whitespace` | No `--token` → stderr ends with `\n` only (no trailing space/tab) |
| 39 | `web/process-output/startup-auto-no-leading-blank-line` | `--token auto` → stderr does not start with `\n`; first bytes `agent-run web token:` |
| 20 | `web/workspace/status-includes-workspace` | GET `/api/agent-run/status` → non-empty `workspace` |
| 21 | `web/workspace/session-meta-includes-workspace` | POST create session → GET detail `session.workspace` = server cwd |
| 32 | `web/timeline/continuation/follow-up-agent-recalls-first-message` | POST hi → follow-up → assistant mentions hi |
| 33 | `web/timeline/streaming/streaming-message-phases-emitted` | Assistant events use phase start/update/end |
| 34 | `web/timeline/streaming/assistant-phases-share-stable-id` | Phased assistant rows share one `id` |
| 35 | `web/stream/sse-delivers-new-events` | SSE after=0 delivers user + assistant events |
| 36 | `web/stream/sse-after-offset-skips-prior` | SSE at EOF offset on finished session → no replay |

## Nested: run auto-session-id

`--auto-session-id` and explicit `--session` same-id policy tests live at
`cmd/agent-run/tests/run/auto-session-id/` (**nested DOCTEST root**). Covers help,
flag mutual exclusion, non-TTY storage slug generation, TTY storage+registry same id,
and collision suffixes. Spec version 0.0.2.

```sh
doctest vet ./cmd/agent-run/tests/run/auto-session-id
doctest test ./cmd/agent-run/tests/run/auto-session-id
doctest test -v ./cmd/agent-run/tests/run/auto-session-id/auto-id/tty/same-id-storage-registry-meta
```

## Nested: run agent-runner-binary / config-home

Grok-tty flag tests under `cmd/agent-run/tests/run/` are **nested DOCTEST roots** (inheritance
firewall). They exercise `--agent-runner-binary` and `--agent-runner-config-home` on
`agent-run run --agent-runner grok-tty` with fake runner scripts (not `AGENT_RUN_GROK_TTY_COMMAND`).

```sh
doctest vet ./cmd/agent-run/tests/run/agent-runner-binary
doctest vet ./cmd/agent-run/tests/run/agent-runner-config-home
doctest test ./cmd/agent-run/tests/run/agent-runner-binary
doctest test ./cmd/agent-run/tests/run/agent-runner-config-home
```

## Nested: web grok-mock-config

Web flag tests for `--grok-home` and `--grok-tty-runner-binary` live at
`cmd/agent-run/tests/web/grok-mock-config/` (nested DOCTEST root).

```sh
doctest vet ./cmd/agent-run/tests/web/grok-mock-config
doctest test ./cmd/agent-run/tests/web/grok-mock-config
```

## Nested: grok-tty (PTY interactive runner)

`grok-tty` tests are a **nested DOCTEST root** (inheritance firewall). Each run spawns
an adhoc ptywrap server on a random port; session ids print to stderr as
`grok-tty: session-N`. Default suite uses `AGENT_RUN_GROK_TTY_COMMAND` fake TUI;
`real-grok/` leaves require `--label grok` and real `grok` on PATH.
Keep-tty leaves test `--keep-tty` flag persistence.

```sh
doctest vet ./cmd/agent-run/tests/grok-tty
doctest test ./cmd/agent-run/tests/grok-tty
doctest test --label grok ./cmd/agent-run/tests/grok-tty/real-grok
doctest test -v ./cmd/agent-run/tests/grok-tty/run/keep-tty-registry-persists
```

## Nested: codex-tty (PTY interactive runner)

`codex-tty` tests are a **nested DOCTEST root** (inheritance firewall). Each run
spawns an adhoc ptywrap server on a random port; session ids print to stderr as
`codex-tty: session-N`. Default suite uses `AGENT_RUN_CODEX_TTY_COMMAND` fake TUI;
`real-codex/` leaves require `--label codex` and real `codex` on PATH.

```sh
doctest vet ./cmd/agent-run/tests/codex-tty
doctest test ./cmd/agent-run/tests/codex-tty
doctest test --label codex ./cmd/agent-run/tests/codex-tty/real-codex
```

## Nested: tty (status / attach / send)

`tty` tests are a **nested DOCTEST root** (inheritance firewall). Tests cover
the `agent-run tty` subcommand group — status, attach, and send — as well as
the `agent-run attach` shortcut alias. Mock registry JSON files and fake
in-process ptywrap HTTP+WebSocket servers provide deterministic test harnesses.

```sh
doctest vet ./cmd/agent-run/tests/tty
doctest test ./cmd/agent-run/tests/tty
doctest test -v ./cmd/agent-run/tests/tty/status/registry-entry-valid/human-readable
doctest test -v ./cmd/agent-run/tests/tty/send/sends-to-live-terminal
```

## Nested: web layout (Playwright)

Mobile layout tests live in a separate root (inheritance firewall):

`cmd/agent-run/tests/web-layout/` — `label: chromium`; viewport 390×844; `playwright-debug` scripts.

```sh
doctest vet ./cmd/agent-run/tests/web-layout
doctest test -v ./cmd/agent-run/tests/web-layout --label chromium
```

## How to Run

```sh
doctest vet ./cmd/agent-run/tests
doctest test -v ./cmd/agent-run/tests
doctest test -v ./cmd/agent-run/tests/run/json-fake-codex-hi
doctest test -v ./cmd/agent-run/tests/web/timeline/message-events-include-timestamp
doctest test -v ./cmd/agent-run/tests/web/timeline/continuation/follow-up-agent-recalls-first-message
doctest test -v ./cmd/agent-run/tests/web/stream
doctest test -v ./cmd/agent-run/tests/web/process-output/web-stdout-silent-on-agent-run
doctest test -v ./cmd/agent-run/tests/web-layout/mobile-message-roles-and-timestamps --label chromium
doctest test -v ./cmd/agent-run/tests/grok-tty/run/captures-tui-output
doctest test --label grok ./cmd/agent-run/tests/grok-tty/real-grok
doctest test -v ./cmd/agent-run/tests/codex-tty/run/captures-tui-output
doctest test --label codex ./cmd/agent-run/tests/codex-tty/real-codex
doctest test -v ./cmd/agent-run/tests/tty/status/registry-entry-valid/human-readable
doctest test -v ./cmd/agent-run/tests/tty/send/sends-to-live-terminal
```

```go
import (
	"bytes"
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot   string
	TempDir    string
	Home       string
	AgentRun   string
	FakeCodex  string
	Args       []string
	Env        []string
	Mode          string // "" | "web"
	WebTokenMode  string // "omit" | "explicit" | "auto" (default explicit)
	WebToken      string // explicit/auto: bearer value; auto: filled from stderr after start
	WebPort       int    // 0 = OS-assigned; -1 = default (8192)
	WebServerStderr string // snapshot from web stderr accumulator at start (legacy)
	webProcessStderr *bytes.Buffer
	webProcessStdout *bytes.Buffer
	HTTPPath      string
	HTTPMethod    string // GET (default) | POST
	HTTPBody      string
	HTTPAuth      string // Bearer value; empty = omit header
	WebCmd        *exec.Cmd
	WebBaseURL    string
	SessionRunner string
	SessionID     string
	CreatePrompt   string
	FollowUpPrompt string
	SSEAfterOffset int64
	SSEMaxWait     time.Duration
	Sidecar        func() // optional; started in goroutine immediately before CLI exec
	ExecTimeout    time.Duration
}

type Response struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Err        error
	HTTPStatus int
	HTTPBody   string
	FilesOutsideHome []string
	EventsFilePath   string
	EventsFileLines  []string
	SSEEvents        []map[string]any
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "web":
		return runWebHTTP(t, req)
	case "sse":
		if req.SSEMaxWait <= 0 {
			req.SSEMaxWait = 45 * time.Second
		}
		events := collectSSESessionEvents(t, req, req.SessionRunner, req.SessionID, req.SSEAfterOffset, req.SSEMaxWait)
		return &Response{SSEEvents: events}, nil
	default:
		return runAgentRun(t, req, req.Args...)
	}
}
```
