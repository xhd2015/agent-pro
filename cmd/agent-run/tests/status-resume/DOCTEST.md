# agent-run session status + resume + run --open post-exit bind

Doc-style tests for session-level `agent-run status <session-id>`,
`agent-run resume`, and `agent-run run --open` post-attach grok session
discovery/bind. Prefer stub fixtures (seeded meta, registry, fake ptywrap,
argv-recording fake runner) over real grok.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — dispatches `status`, `resume`, and `run --open`. Status is
  a multi-layer probe (not a meta dump only). Resume is a **shortcut of `run`**
  that reuses session meta and injects provider resume (`grok --resume <id>`).
- **Session storage** — `AGENT_RUN_HOME/sessions/<runner>/<session_id>/meta.json`
  (`SessionMeta`: runner, session_id, runner_session_id, terminal_session_id,
  status, workspace, model, …).
- **TTY registry** — `AGENT_RUN_HOME/grok-tty-registry/<terminal_session_id>.json`
  (`pid`, `listen_addr`, `created_at`). Absent or unreachable ⇒ process/terminal
  dead layers.
- **Keep-alive serve process** — registry PID for process layer (`alive|dead|unknown`).
- **ptywrap / terminal** — TCP reachability, screen status, sendable (tests use
  fake in-process HTTP+WebSocket when a live terminal is required).
- **Provider runner (grok)** — bound when `meta.runner_session_id` is non-empty;
  **`runner.exited`** is true when the provider agent has exited (unreachable +
  bound, child gone, resume-footer signals); false when live/sendable/idle.
- **Fake runner / hooks** — `AGENT_RUN_GROK_TTY_COMMAND`,
  `--agent-runner-binary` argv recorder, `AGENT_RUN_OPEN_ATTACH_INSTANT=1`,
  `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` + temp `GROK_HOME` for discovery.

**Behaviors**

```
# help
agent-run --help -> lists resume
agent-run status --help -> documents session-ref / multi-layer probe
agent-run resume --help -> lists --open, session-id, followup

# status
agent-run status -> home: <path>\n  (compat bare)
agent-run status <session-id> -> multi-layer probe
agent-run status <runner>/<session_id> -> unambiguous resolve
agent-run status --json <ref> -> JSON mirror (runner.exited, resume.ready)

# layers (human + JSON)
session / process / terminal / runner (bound|unbound + exited) / resume (ready)

# resume gate
resume ready ⇔ runner_session_id non-empty ∧ runner.exited == true
  exited false (live) -> deny; hint send
  unbound -> deny
  missing session -> exit 1
  no followup without --open -> exit 1

# resume ≡ run shortcut
agent-run resume [flags] <session-id> ["followup"]
  ≡ agent-run run [flags] --session-id=<id> --agent-runner=<meta.runner> …
    with ResumeSessionID = meta.runner_session_id
    argv includes grok --resume <id> (when not overridden by TTY command hook)

# run --open post-exit
after attach returns:
  discovery OK -> stderr:
    grok-tty: <terminal-or-session-id>
    grok-tty: grok session <uuid>
    grok-tty: grok updates <path>
    and persist meta.runner_session_id
  discovery fail -> exit ≠ 0; stderr error "grok session id not resolved"
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/status-resume/
├── DOCTEST.md
├── SETUP.md                                    # build agent-run, seed meta/registry, fake ptywrap
├── help/                                       # command discovery
│   ├── top-level-lists-resume/                 # H1 agent-run --help lists resume
│   ├── status-documents-session-ref/           # H2 status --help session-ref / layers
│   └── resume-lists-open/                      # H3 resume --help lists --open + session-id
├── status/                                     # agent-run status [ref]
│   ├── bare-home/                              # S0 bare status → home path
│   ├── missing-session/                        # S1 unknown id → exit 1
│   ├── session-ref/
│   │   └── runner-slash-id/                    # S6 runner/session resolves uniquely
│   └── layers/                                 # multi-layer probe outcomes
│       ├── bound-exited-resume-ready/          # S2 dead terminal + bound + exited → ready
│       ├── live-bound-not-exited/              # S3 live sendable → exited false, not ready
│       ├── unbound-not-ready/                  # S4 no runner_session_id → unbound
│       └── json-shape/                         # S5 --json has runner.exited + resume.ready
├── resume/                                     # agent-run resume
│   ├── denied/
│   │   ├── not-exited/                         # R1 live bound → exit 1, hint send
│   │   ├── unbound/                            # R2 no runner_session_id
│   │   ├── missing-session/                    # R3 session not found
│   │   └── missing-prompt/                     # R4 no followup without --open
│   ├── headless-followup-when-exited/          # R5 gate open; --resume in argv
│   └── open-flag-accepted/                     # R6 --open known (not unknown-flag)
└── run-open-post-exit/                         # run --open after attach returns
    ├── prints-grok-session-when-resolved/      # O1 bind + stderr + meta persist
    └── errors-when-unresolved/                 # O2 exit ≠ 0, not-resolved error
```

Parameter ranking (most → least significant):

1. **Command surface** — help | status | resume | run-open-post-exit
2. **Status invocation** — bare home | missing | session-ref form | multi-layer state
3. **Runner/resume state** — bound+exited | live not-exited | unbound | JSON format
4. **Resume gate outcome** — denied reason vs headless success vs `--open` acceptance
5. **Open discovery outcome** — resolved vs unresolved

## Test Index

| # | Leaf | Req | Description |
|---|------|-----|-------------|
| 1 | `help/top-level-lists-resume` | H1 | `agent-run --help` stdout contains `resume` |
| 2 | `help/status-documents-session-ref` | H2 | `status --help` mentions session id / layers |
| 3 | `help/resume-lists-open` | H3 | `resume --help` lists `--open` and session-id |
| 4 | `status/bare-home` | S0 | bare `status` → exit 0, `home: …\n` |
| 5 | `status/missing-session` | S1 | unknown session → exit 1, not found |
| 6 | `status/session-ref/runner-slash-id` | S6 | `grok-tty/<id>` resolves same session as bare id |
| 7 | `status/layers/bound-exited-resume-ready` | S2 | bound + exited + dead terminal → `resume.ready: yes` |
| 8 | `status/layers/live-bound-not-exited` | S3 | live idle terminal → `exited: false`, not ready, hint send |
| 9 | `status/layers/unbound-not-ready` | S4 | empty `runner_session_id` → unbound, not ready |
| 10 | `status/layers/json-shape` | S5 | `--json` includes `runner.exited` and `resume.ready` |
| 11 | `resume/denied/not-exited` | R1 | live bound → exit 1, cannot resume / use send |
| 12 | `resume/denied/unbound` | R2 | unbound → exit 1, not bound |
| 13 | `resume/denied/missing-session` | R3 | missing meta → exit 1, not found |
| 14 | `resume/denied/missing-prompt` | R4 | exited bound, no followup, no `--open` → prompt required |
| 15 | `resume/headless-followup-when-exited` | R5 | exited bound + followup → run path with `--resume <id>` |
| 16 | `resume/open-flag-accepted` | R6 | `resume --open …` not rejected as unknown flag |
| 17 | `run-open-post-exit/prints-grok-session-when-resolved` | O1 | after attach, stderr grok session + meta.runner_session_id |
| 18 | `run-open-post-exit/errors-when-unresolved` | O2 | discovery fail → exit ≠ 0, not resolved error |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/status-resume
doctest test ./cmd/agent-run/tests/status-resume

# help
doctest test -v ./cmd/agent-run/tests/status-resume/help/top-level-lists-resume
doctest test -v ./cmd/agent-run/tests/status-resume/help/status-documents-session-ref
doctest test -v ./cmd/agent-run/tests/status-resume/help/resume-lists-open

# status
doctest test -v ./cmd/agent-run/tests/status-resume/status/bare-home
doctest test -v ./cmd/agent-run/tests/status-resume/status/missing-session
doctest test -v ./cmd/agent-run/tests/status-resume/status/session-ref/runner-slash-id
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/bound-exited-resume-ready
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/live-bound-not-exited
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/unbound-not-ready
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/json-shape

# resume
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/not-exited
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/unbound
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/missing-session
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/missing-prompt
doctest test -v ./cmd/agent-run/tests/status-resume/resume/headless-followup-when-exited
doctest test -v ./cmd/agent-run/tests/status-resume/resume/open-flag-accepted

# run --open post-exit
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-post-exit/prints-grok-session-when-resolved
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-post-exit/errors-when-unresolved
```

```go
import (
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Args     []string
	Env      []string

	// Mode selects post-exec enrichment in Run:
	//   ""              — plain CLI exec
	//   "status-json"   — parse stdout as JSON into Response.JSONBody
	//   "read-meta"     — after CLI, load sessions/<runner>/<session>/meta.json
	Mode        string
	ExecTimeout time.Duration

	// Session seed (meta.json under sessions/<Runner>/<SessionID>/)
	SeedMeta          bool
	Runner            string // default grok-tty
	SessionID         string
	MetaStatus        string // running | finished | error
	RunnerSessionID   string // provider (grok) session id — resume key
	TerminalSessionID string
	Workspace         string
	Model             string
	InitialPrompt     string

	// Registry / process / terminal layers
	WriteRegistry         bool
	RegistryPID           int  // 0 = os.Getpid() (alive); >0 use as-is; -1 = dead sentinel
	RegistryClosedPort    bool // listen on closed port (unreachable)
	StartFakePTYWrap      bool
	FakePTYWrapPort       int
	FakePTYWrapScrollback string

	// Open / discovery fixtures
	OpenInstantAttach bool
	GrokTTYCommand    string
	GrokHome          string
	GrokSessionUUID   string
	GrokUpdatesPath   string

	// Resume argv probe (no AGENT_RUN_GROK_TTY_COMMAND — use binary path)
	ArgvProbePath      string
	RunnerScriptPath   string
	AgentRunnerBinary  string
	FollowupPrompt     string
}

type RegistryEntryData struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
	CreatedAt  string `json:"created_at"`
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
	JSONBody map[string]any
	// MetaAfter is filled when Mode == "read-meta".
	MetaAfter map[string]any
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
