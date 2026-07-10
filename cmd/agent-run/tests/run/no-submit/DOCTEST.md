# agent-run run --no-submit Tests

Doc-style tests for `agent-run run --no-submit`: with required `--open`, inject a
non-empty prompt into the TTY input box **without** trailing Enter (`\r`), so the
user can edit before submitting after auto-attach.

Out of scope: standalone draft semantics without `--open`, non-TTY draft mode,
changing `--keep-tty` independently of `--open`, attach/web/send-queue flags.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --no-submit` validates flag pairing, then (with
  `--open`) starts the open lifecycle and injects the prompt without submit.
- **`--open`** — required companion of `--no-submit`; provides TTY open
  keep-alive + auto-attach semantics.
- **TTY runners** — `grok-tty`, `codex-tty` (and other TTY backends); only these
  may use `--open` / `--no-submit`. Non-TTY runners (`fake-codex`, …) reject
  the open family of flags.
- **Prompt inject path** — `ttywatch.SendMessage(..., suffixCR=false)` when
  `NoSubmit` is true; default remains `suffixCR=true` (auto-submit on Enter).
- **CR-sensitive fake TUI** — `AGENT_RUN_GROK_TTY_COMMAND` replacement that
  prints `SUBMITTED:<line>` only after Enter (`\r` / line completion). Typing
  without Enter must not produce `SUBMITTED:`.
- **Open attach test hook** — `AGENT_RUN_OPEN_ATTACH_INSTANT=1` makes auto-attach
  return immediately so CI leaves complete without an interactive controlling TTY.
- **Terminal registry / snapshot** — after open completes, registry + listen_addr
  remain; tests may snapshot PTY text to prove inject-without-submit.

**Behaviors**

```
# help
agent-run run --help -> documents --no-submit (pairs with --open); stdout ends with \n

# reject without --open
agent-run run --no-submit --agent-runner grok-tty "x"
  -> exit ≠ 0; error explains --no-submit requires --open

# reject non-TTY (open family)
agent-run run --open --no-submit --agent-runner fake-codex "x"
  -> exit ≠ 0; clear TTY / open error

# with --open --no-submit (core)
agent-run run --agent-runner grok-tty --open --no-submit "draft"
  1. validate --no-submit implies --open; TTY runner; not bare inject-without-open
  2. start keep-alive TTY session (silent open path)
  3. inject prompt with suffixCR=false (no trailing \r)
  4. auto-attach (instant hook in tests)
  5. on attach exit: stderr once "grok-tty: <id>"
  6. PTY scrollback must NOT contain SUBMITTED:draft
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run/no-submit/
├── DOCTEST.md
├── SETUP.md
├── help/
│   └── run-help-lists-no-submit/      # run --help documents --no-submit; trailing \n
├── reject/                            # invalid --no-submit combinations
│   ├── without-open/                  # --no-submit without --open → error
│   └── non-tty/
│       └── fake-codex/                # --open --no-submit + non-TTY → error
└── with-open/                         # valid --open --no-submit path
    └── injects-without-cr/            # draft injected; no auto-submit (no SUBMITTED:)
```

Parameter ranking (most → least significant):

1. **Outcome class** — help | reject | with-open
2. **Reject reason** — missing `--open` vs non-TTY runner under open family
3. **With-open aspect** — inject-without-CR (core product behavior)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-no-submit` | `run --help` lists `--no-submit`; stdout ends with `\n` |
| 2 | `reject/without-open` | `--no-submit` without `--open` → exit ≠ 0; requires `--open` |
| 3 | `reject/non-tty/fake-codex` | `--open --no-submit` + `fake-codex` → exit ≠ 0; TTY/open error |
| 4 | `with-open/injects-without-cr` | `--open --no-submit "draft"`: no `SUBMITTED:draft`; session id after attach |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/no-submit
doctest test ./cmd/agent-run/tests/run/no-submit
doctest test -v ./cmd/agent-run/tests/run/no-submit/help/run-help-lists-no-submit
doctest test -v ./cmd/agent-run/tests/run/no-submit/reject/without-open
doctest test -v ./cmd/agent-run/tests/run/no-submit/reject/non-tty/fake-codex
doctest test -v ./cmd/agent-run/tests/run/no-submit/with-open/injects-without-cr
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
	// Mode selects Run branch:
	//   ""                    = plain exec
	//   "open-registry-after" = exec then load registry for stderr session id
	//   "open-snapshot-after" = exec, load registry, snapshot PTY text
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
	// Snapshot is printable PTY scrollback when Mode == "open-snapshot-after".
	Snapshot string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
