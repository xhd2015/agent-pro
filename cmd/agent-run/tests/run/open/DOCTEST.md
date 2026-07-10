# agent-run run --open Tests

Doc-style tests for `agent-run run --open`: TTY-only interactive open mode that
starts a keep-alive terminal session, optionally injects a prompt, auto-attaches
silently, and prints the terminal session id **once after** attach exits.

Out of scope: fixing grok session discovery hang / “Resolve session id…” binding.
Under `--open`, discovery must not print progress/errors to the user’s screen.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --open` validates flags/runner/prompt, starts the TTY
  session, auto-attaches, then exits after printing the session id.
- **TTY runners** — `grok-tty`, `codex-tty` (and stub-tty in other trees); only
  these may use `--open`. Non-TTY runners (`fake-codex`, `opencode`, …) reject it.
- **Headless open path** — `KeepTerminalAlive` implied; suppress human/JSON event
  emit to stdout/stderr while starting and while attached.
- **Auto-attach** — same path as `agent-run attach` (`ttywatch.AttachWriter`);
  user exits attach → CLI prints once.
- **Terminal registry** — e.g. `AGENT_RUN_HOME/grok-tty-registry/<id>.json`;
  remains for re-attach/send after open completes.
- **Fake TUI** — `AGENT_RUN_GROK_TTY_COMMAND` / `AGENT_RUN_CODEX_TTY_COMMAND`
  replace the real TUI for deterministic PTY runs.
- **Open attach test hook** — `AGENT_RUN_OPEN_ATTACH_INSTANT=1` makes auto-attach
  return immediately (no interactive TTY required in CI). Implementer must honor
  this env for lifecycle leaves; product attach path is unchanged for humans.

**Behaviors**

```
# help
agent-run run --help -> documents --open; stdout ends with \n

# reject non-TTY
agent-run run --open --agent-runner fake-codex "x" -> exit ≠ 0; clear error

# reject --open + --json
agent-run run --open --json --agent-runner grok-tty "x" -> exit ≠ 0; clear error

# empty prompt policy
agent-run run --agent-runner grok-tty          -> "prompt is required", exit ≠ 0
agent-run run --agent-runner grok-tty --open   -> allowed (no prompt required)

# TTY open lifecycle
agent-run run --agent-runner grok-tty --open ["prompt"]
  1. validate TTY runner; not --json
  2. start keep-alive TTY session (silent: no event stream, no discovery think)
  3. inject prompt only if non-empty
  4. auto-attach (silent while attached)
  5. on attach exit: stderr once "<runner>: <terminal-session-id>"
  6. leave registry/PTY alive for later attach/send
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run/open/
├── DOCTEST.md
├── SETUP.md
├── help/
│   └── run-help-lists-open/           # run --help documents --open; trailing \n
├── reject/                            # invalid --open combinations
│   ├── non-tty/
│   │   └── fake-codex/                # --open + non-TTY → error
│   └── with-json/
│       └── tty/                       # --open + --json + TTY → error
├── prompt-policy/                     # empty prompt rules
│   ├── without-open-empty-required/   # no --open, empty → prompt is required
│   └── with-open-empty-allowed/       # --open, empty + TTY → not prompt-required
└── tty-lifecycle/                     # --open + TTY happy path (instant-attach hook)
    ├── silence-no-stream/             # no discovery/event noise; no pre-attach id only
    ├── prints-id-after-attach/        # after attach returns: stderr runner: <id> once
    ├── keep-alive-registry/           # registry file + listen_addr alive after open
    └── codex-tty-accepted/            # --open accepted for codex-tty (not non-TTY)
```

Parameter ranking (most → least significant):

1. **Outcome class** — help | reject | prompt-policy | tty-lifecycle
2. **Reject reason** — non-TTY runner vs `--open`+`--json`
3. **Prompt policy** — with `--open` vs without
4. **Lifecycle aspect** — silence | post-attach id | keep-alive registry | runner acceptance

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-open` | `run --help` lists `--open`; stdout ends with `\n` |
| 2 | `reject/non-tty/fake-codex` | `--open` + `fake-codex` → exit ≠ 0; clear non-TTY / open error |
| 3 | `reject/with-json/tty` | `--open` + `--json` + TTY → exit ≠ 0; clear conflict error |
| 4 | `prompt-policy/without-open-empty-required` | Empty prompt without `--open` → `prompt is required` |
| 5 | `prompt-policy/with-open-empty-allowed` | `--open` without prompt on TTY → not rejected as prompt required |
| 6 | `tty-lifecycle/silence-no-stream` | Open run silent: no `Resolve session id` / 💭 / event stream; final id only after attach |
| 7 | `tty-lifecycle/prints-id-after-attach` | After attach returns, stderr has `grok-tty: <id>` exactly once |
| 8 | `tty-lifecycle/keep-alive-registry` | After open completes, registry entry exists and TCP `listen_addr` is reachable |
| 9 | `tty-lifecycle/codex-tty-accepted` | `--open` + `codex-tty` not rejected as unknown/non-TTY |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/open
doctest test ./cmd/agent-run/tests/run/open
doctest test -v ./cmd/agent-run/tests/run/open/help/run-help-lists-open
doctest test -v ./cmd/agent-run/tests/run/open/reject/non-tty/fake-codex
doctest test -v ./cmd/agent-run/tests/run/open/reject/with-json/tty
doctest test -v ./cmd/agent-run/tests/run/open/prompt-policy/with-open-empty-allowed
doctest test -v ./cmd/agent-run/tests/run/open/tty-lifecycle/silence-no-stream
doctest test -v ./cmd/agent-run/tests/run/open/tty-lifecycle/prints-id-after-attach
doctest test -v ./cmd/agent-run/tests/run/open/tty-lifecycle/keep-alive-registry
```

```go
import (
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
	Prompt     string
	Runner     string // "fake-codex" | "grok-tty" | "codex-tty" | ""
	// GrokTTYCommand is AGENT_RUN_GROK_TTY_COMMAND when using grok-tty.
	GrokTTYCommand string
	// CodexTTYCommand is AGENT_RUN_CODEX_TTY_COMMAND when using codex-tty.
	CodexTTYCommand string
	// OpenInstantAttach sets AGENT_RUN_OPEN_ATTACH_INSTANT=1 so auto-attach
	// returns immediately (CI-safe; no interactive controlling TTY).
	OpenInstantAttach bool
	// Mode selects Run branch: "" = plain exec; "open-registry-after" = exec then
	// read registry for the session id from stderr.
	Mode        string
	ExecTimeout time.Duration
}

type RegistryEntry struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
}

type Response struct {
	Stdout        string
	Stderr        string
	ExitCode      int
	Err           error
	RegistryEntry *RegistryEntry
	SessionID     string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
