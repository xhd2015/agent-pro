# agent-run run --no-submit Tests

Doc-style tests for `agent-run run --no-submit`: with required `--open`, put a
non-empty draft into the TTY input **without** auto-submitting a model turn, so
the user can edit before sending after auto-attach.

**Core oracle (Option C)**: real Grok TUI under `llm-mock-run-grok` with **no**
`LLM_MOCK_RUN_GROK_COMMAND` and **no** `AGENT_RUN_GROK_TTY_COMMAND`. Primary
proof is **absence of mock HTTP chat / session `user_message` for the draft** —
not a fake `SUBMITTED:` marker from a bash `read` TUI.

Out of scope: standalone draft semantics without `--open`, non-TTY draft mode,
changing `--keep-tty` independently of `--open`, full send-queue matrix, web UI
draft.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --no-submit` validates flag pairing, then (with
  `--open`) starts the open lifecycle and stages the draft without submit.
- **`--open`** — required companion of `--no-submit`; keep-alive TTY open +
  auto-attach; prints terminal id after attach (`grok-tty: <id>`).
- **TTY runners** — `grok-tty` (and other TTY backends); only these may use
  `--open` / `--no-submit`. Non-TTY runners (`fake-codex`, …) reject the open
  family.
- **Argv + inject policy (`agenttty.RunHeadless`)** — for a **new** session,
  default still puts a non-empty prompt on trailing positional argv (real Grok
  treats that as the first submitted turn). Under **`NoSubmit`**, product must
  **not** put the draft on argv; under `--open`, inject with `suffixCR=false`
  for new sessions too. Resume follow-ups remain inject-only.
- **Open bind (`agentui` open_bind)** — after open, discovers provider session.
  **Bind policy A**: `--no-submit` ⇒ **soft unbound** (must not hard-fail solely
  because there is no grok session id yet — draft-only, no turn).
- **`llm-mock-run-grok`** — real `grok` TUI + mock HTTP; optional
  `LLM_MOCK_RUN_FLAGS=--log-http <file>.jsonl` records chat/responses
  round-trips. Resolves the mock server as a **same-directory sibling `llm-mock`**
  binary (tests build both). Must run **without** `LLM_MOCK_RUN_GROK_COMMAND` /
  `AGENT_RUN_GROK_TTY_COMMAND` hooks for Option C.
- **Open attach test hook** — `AGENT_RUN_OPEN_ATTACH_INSTANT=1` makes auto-attach
  return immediately so CI leaves complete without an interactive controlling TTY.
- **Control path** — same harness **without** `--no-submit` must produce turn
  evidence (mock HTTP and/or session `user_message`), proving the oracle is live.

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

# with --open --no-submit (Option C core; real grok under llm-mock-run-grok)
agent-run run --agent-runner grok-tty --open --no-submit "draft-…"
  --agent-runner-binary <built>/llm-mock-run-grok
  --agent-runner-config-home <isolated>
  1. validate flags + TTY runner
  2. start keep-alive TTY: real grok via mock binary (no hooks)
  3. do NOT append draft to argv as positional PROMPT
  4. inject draft with suffixCR=false after open-ready
  5. auto-attach (instant hook in tests)
  6. stderr once "grok-tty: <id>"
  7. open bind soft-unbound if no provider session (no hard error for NoSubmit)
  8. after settle: no mock HTTP chat for draft AND no session user_message for draft

# control (same harness, no --no-submit)
agent-run run --agent-runner grok-tty --open "draft-…"
  --agent-runner-binary llm-mock-run-grok …
  -> turn evidence: ≥1 mock HTTP chat and/or user_message for prompt
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
└── with-open/                         # valid --open path (real grok / Option C)
    ├── no-model-turn/                 # --open --no-submit: no HTTP / no user_message
    └── control-model-turn/            # --open without --no-submit: turn evidence (oracle)
```

Parameter ranking (most → least significant):

1. **Outcome class** — help | reject | with-open
2. **Reject reason** — missing `--open` vs non-TTY runner under open family
3. **With-open submit mode** — no-submit (no model turn) vs control (submits turn)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-no-submit` | `run --help` lists `--no-submit`; stdout ends with `\n` |
| 2 | `reject/without-open` | `--no-submit` without `--open` → exit ≠ 0; requires `--open` |
| 3 | `reject/non-tty/fake-codex` | `--open --no-submit` + `fake-codex` → exit ≠ 0; TTY/open error |
| 4 | `with-open/no-model-turn` | Option C: real grok + llm-mock-run-grok; `--open --no-submit`; no mock HTTP / no session user_message for draft; soft unbound OK (`label: real-grok, slow`) |
| 5 | `with-open/control-model-turn` | Control: same harness without `--no-submit`; turn evidence proves oracle (`label: real-grok, slow`) |

## How to Run

```sh
# Structure
doctest vet ./cmd/agent-run/tests/run/no-submit

# Lightweight (help + reject; no real grok)
doctest test ./cmd/agent-run/tests/run/no-submit/help
doctest test ./cmd/agent-run/tests/run/no-submit/reject
doctest test -v ./cmd/agent-run/tests/run/no-submit/help/run-help-lists-no-submit
doctest test -v ./cmd/agent-run/tests/run/no-submit/reject/without-open
doctest test -v ./cmd/agent-run/tests/run/no-submit/reject/non-tty/fake-codex

# Option C core + control (requires real grok on PATH)
doctest test --label real-grok ./cmd/agent-run/tests/run/no-submit/with-open
doctest test -v --label real-grok ./cmd/agent-run/tests/run/no-submit/with-open/no-model-turn
doctest test -v --label real-grok ./cmd/agent-run/tests/run/no-submit/with-open/control-model-turn

# Full tree (with-open leaves skipped unless --label real-grok matches)
doctest test ./cmd/agent-run/tests/run/no-submit
```

```go
import (
	"testing"
	"time"
)

type Request struct {
	RepoRoot  string
	TempDir   string
	Home      string // AGENT_RUN_HOME
	Workspace string // --dir project-like cwd
	AgentRun  string
	FakeCodex string
	// LLMMockRunGrok is the session-built llm-mock-run-grok absolute path.
	LLMMockRunGrok string
	// AgentRunnerBinary is --agent-runner-binary (typically LLMMockRunGrok).
	AgentRunnerBinary string
	// GrokHome is --agent-runner-config-home / isolated GROK_HOME for Option C.
	GrokHome string
	// LogHTTPPath is LLM_MOCK_RUN_FLAGS --log-http target (.jsonl).
	LogHTTPPath string
	Args        []string
	Env         []string
	Prompt      string
	Runner      string // "fake-codex" | "grok-tty" | "codex-tty" | ""
	// NoSubmit selects --no-submit on the open path (with-open leaves).
	NoSubmit bool
	// GrokTTYCommand is AGENT_RUN_GROK_TTY_COMMAND when using a fake TUI (reject helpers only).
	GrokTTYCommand string
	// CodexTTYCommand is AGENT_RUN_CODEX_TTY_COMMAND when using codex-tty.
	CodexTTYCommand string
	// OpenInstantAttach sets AGENT_RUN_OPEN_ATTACH_INSTANT=1 so auto-attach
	// returns immediately (CI-safe; no interactive controlling TTY).
	OpenInstantAttach bool
	// Mode selects Run branch:
	//   ""                     = plain exec
	//   "open-registry-after"  = exec then load registry for stderr session id
	//   "open-snapshot-after"  = exec, load registry, snapshot PTY text
	//   "open-real-grok-after" = exec, settle, read --log-http + scan GROK_HOME for turn evidence
	Mode string
	// SettleAfter is post-exec wait before reading turn oracles (Option C).
	SettleAfter time.Duration
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
	// LogHTTPContent / LogHTTPLines are mock --log-http JSONL after open-real-grok-after.
	LogHTTPContent string
	LogHTTPLines   []string
	// HasMockChatHTTP is true when log-http has a chat/completions (or similar) exchange
	// whose body/path looks like a model turn, especially containing the draft prompt.
	HasMockChatHTTP bool
	// HasUserMessageForPrompt is true when GROK_HOME sessions contain a
	// user_message_chunk for req.Prompt (or the prompt substring).
	HasUserMessageForPrompt bool
	// GrokSessionsScanned counts session dirs inspected under GrokHome.
	GrokSessionsScanned int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
