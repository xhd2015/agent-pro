# agent-run run/resume --detach Tests

Doc-style tests for `agent-run run --detach`, `agent-run resume --detach`, and
`--auto-send-or-resume` paths that honor (or ignore) `--detach`. Modeled after
`tty-watch run --detach` and contrasted with `run --open`:

- PTY runs in a keep-alive daemon (`HeadlessRun` / `__serve__`)
- Parent exits after registry registration + optional **soft** grok bind
  (product budget **1 minute**; miss still exit 0)
- **No** event/message streaming
- **No** interactive attach (`--open`)
- Exclusive with `--open` and `--json`
- Empty prompt allowed (like `--open`)
- Stdout always prints both `session-id:` and `terminal-id:` (no ANSI)
- TTY runners only

Classic TDD: feature not implemented yet — leaves must stay **RED** until
implementer lands `--detach`.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --detach` / `resume --detach` validate flags, start a
  keep-alive TTY daemon, soft-bind grok when possible, print both ids, exit.
- **TTY runners** — `grok-tty`, `codex-tty` (tests use fake TUI hooks). Non-TTY
  runners (`fake-codex`, …) reject `--detach`.
- **Detach path** — `KeepTerminalAlive` implied; **no** auto-attach; **no** human
  or NDJSON event stream to the parent process.
- **Terminal registry** — `AGENT_RUN_HOME/<runner>-registry/<terminal-id>.json`
  with `listen_addr` / `pid`; remains reachable after parent exits.
- **Session store** — flat `sessions/<session-id>/meta.json`;
  `terminal_session_id` set on register; soft bind may set `runner_session_id`;
  `status` stays **`running`** (parent does not wait for a full turn).
- **Soft grok bind** — up to ~1 minute discovery budget; on miss parent still
  exits 0 (unbound soft). Optional stderr:
  `grok-tty: grok session <id>` / `grok-tty: grok updates <path>` on hit.
- **Fake TUI** — `AGENT_RUN_GROK_TTY_COMMAND` / `AGENT_RUN_CODEX_TTY_COMMAND`
  replace the real TUI for deterministic PTY runs.
- **Auto-send-or-resume** — MODE=run/resume honor `--detach` like run/resume;
  MODE=send (live) **ignores** `--detach` with a `note:` on stderr (like live
  `--open`).
- **`--no-submit`** — still requires `--open`; `--detach` does **not** unlock it.

**Behaviors**

```
# help
agent-run run --help    -> documents --detach (+ exclusivity hints OK)
agent-run resume --help -> documents --detach

# reject
agent-run run --detach --open …  -> exit ≠ 0; mutually exclusive
agent-run run --detach --json …  -> exit ≠ 0; mutually exclusive
agent-run run --detach --agent-runner fake-codex … -> exit ≠ 0; TTY required
agent-run run --detach --no-submit … -> exit ≠ 0; --no-submit requires --open
# same exclusive open/json on resume --detach

# empty prompt
agent-run run --agent-runner grok-tty --detach
  -> exit 0; stdout session-id: … + terminal-id: …; no stream/attach

# run --detach with prompt
agent-run run --agent-runner grok-tty --detach "hi"
  -> start keep-alive daemon; soft bind (≤1m, soft miss OK)
  -> print both ids on stdout; exit 0 quickly after registry (+ soft budget)
  -> no 💭/💬/NDJSON / "Resolve session id" noise
  -> registry + TCP alive after parent exits
  -> meta.status = running

# resume --detach
seed bound+exited
  -> agent-run resume --detach <session-id>
  -> reopen daemon with provider resume; no attach
  -> exit 0; both ids; registry alive

# auto
MODE=run    + --detach -> create detached (daemon + both ids)
MODE=resume + --detach -> resume detach (no open/attach)
MODE=send   + --detach -> note: ignored; send proceeds (msg_N)
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run/detach/
├── DOCTEST.md
├── SETUP.md
├── help/                                      # flag discovery
│   ├── run-help-lists-detach/                 # run --help documents --detach
│   └── resume-help-lists-detach/              # resume --help documents --detach
├── reject/                                    # invalid combinations
│   ├── with-open/
│   │   ├── run/                               # run --detach --open
│   │   └── resume/                            # resume --detach --open
│   ├── with-json/
│   │   ├── run/                               # run --detach --json
│   │   └── resume/                            # resume --detach --json
│   ├── non-tty/
│   │   └── fake-codex/                        # --detach + fake-codex
│   └── no-submit-still-requires-open/         # --detach does not unlock --no-submit
├── prompt-policy/
│   └── empty-prompt-allowed/                  # empty prompt OK under --detach
├── tty-lifecycle/                             # run --detach happy path (fake TUI)
│   ├── with-prompt/                           # prompt + exit 0 + both ids
│   ├── prints-both-ids/                       # stdout shape: session-id + terminal-id
│   ├── silence-no-stream/                     # no discovery/event stream noise
│   ├── keep-alive-registry/                   # registry file + TCP after parent exits
│   ├── soft-bind-miss-exit-0/                 # isolated HOME; unbound still exit 0
│   └── status-running-after/                  # meta.status=running; terminal live
├── resume/                                    # resume --detach success
│   ├── empty-followup-reopen/                 # no followup; both ids; no attach
│   └── keep-alive-registry/                   # registry alive after resume detach
└── auto/                                      # --auto-send-or-resume + --detach
    ├── mode-run-creates-detached/             # missing session → detach create
    ├── mode-resume-detaches/                  # bound+exited → resume detach
    └── mode-send-ignores-detach/              # live → note ignore; send proceeds
```

Parameter ranking (most → least significant):

1. **Outcome class** — help | reject | prompt-policy | tty-lifecycle | resume | auto
2. **Reject reason** — open | json | non-TTY | no-submit gate
3. **CLI surface within reject exclusivity** — run vs resume
4. **Lifecycle aspect** — prompt | id print | silence | keep-alive | soft bind | status
5. **Resume aspect** — empty reopen vs keep-alive
6. **Auto mode** — run | resume | send (ignore)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-detach` | `run --help` lists `--detach`; stdout ends with `\n` |
| 2 | `help/resume-help-lists-detach` | `resume --help` lists `--detach`; trailing `\n` |
| 3 | `reject/with-open/run` | `run --detach --open` → exit ≠ 0; mutual exclusion |
| 4 | `reject/with-open/resume` | `resume --detach --open` → exit ≠ 0; mutual exclusion |
| 5 | `reject/with-json/run` | `run --detach --json` → exit ≠ 0; mutual exclusion |
| 6 | `reject/with-json/resume` | `resume --detach --json` → exit ≠ 0; mutual exclusion |
| 7 | `reject/non-tty/fake-codex` | `--detach` + `fake-codex` → exit ≠ 0; TTY required |
| 8 | `reject/no-submit-still-requires-open` | `--detach --no-submit` still requires `--open` |
| 9 | `prompt-policy/empty-prompt-allowed` | `--detach` empty prompt → exit 0; both ids |
| 10 | `tty-lifecycle/with-prompt` | `--detach "hi"` → exit 0; both ids printed |
| 11 | `tty-lifecycle/prints-both-ids` | stdout has `session-id:` and `terminal-id:` lines |
| 12 | `tty-lifecycle/silence-no-stream` | no 💭/💬/NDJSON / Resolve session id |
| 13 | `tty-lifecycle/keep-alive-registry` | registry + TCP reachable after parent exits |
| 14 | `tty-lifecycle/soft-bind-miss-exit-0` | isolated HOME / no grok → still exit 0 |
| 15 | `tty-lifecycle/status-running-after` | meta.status=`running`; terminal still live |
| 16 | `resume/empty-followup-reopen` | `resume --detach <id>` no followup → exit 0; both ids |
| 17 | `resume/keep-alive-registry` | registry alive after resume detach |
| 18 | `auto/mode-run-creates-detached` | auto MODE=run + detach → create daemon + both ids |
| 19 | `auto/mode-resume-detaches` | auto MODE=resume + detach → detach reopen |
| 20 | `auto/mode-send-ignores-detach` | auto MODE=send + detach → note ignore; `msg_N` |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/run/detach                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/run/detach
doctest test --label-all ./cmd/agent-run/tests/run/detach

doctest vet ./cmd/agent-run/tests/run/detach
doctest test ./cmd/agent-run/tests/run/detach
doctest test ./cmd/agent-run/tests/run/detach/...

# help / reject
doctest test -v ./cmd/agent-run/tests/run/detach/help/run-help-lists-detach
doctest test -v ./cmd/agent-run/tests/run/detach/help/resume-help-lists-detach
doctest test -v ./cmd/agent-run/tests/run/detach/reject

# run --detach happy path
doctest test -v ./cmd/agent-run/tests/run/detach/prompt-policy/empty-prompt-allowed
doctest test -v ./cmd/agent-run/tests/run/detach/tty-lifecycle

# resume / auto
doctest test -v ./cmd/agent-run/tests/run/detach/resume
doctest test -v ./cmd/agent-run/tests/run/detach/auto
```

Classic TDD note: leaves that assert new `--detach` success or product rejection
wording (not merely “unknown flag”) are expected **RED** until implementation.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot  string
	TempDir   string
	WorkDir   string
	Home      string
	AgentRun  string
	FakeCodex string
	Args      []string
	Env       []string
	Prompt    string
	Runner    string // "fake-codex" | "grok-tty" | "codex-tty" | ""

	// GrokTTYCommand / CodexTTYCommand map to AGENT_RUN_*_TTY_COMMAND hooks.
	GrokTTYCommand  string
	CodexTTYCommand string

	// Mode selects Run post-processing:
	//   ""                       plain exec
	//   "detach-registry-after"  parse stdout ids + load registry entry
	//   "read-meta"              load sessions/<SessionID>/meta.json (or parsed session-id)
	//   "read-meta+registry"     both meta and registry
	Mode        string
	ExecTimeout time.Duration

	// Session seed (flat sessions/<SessionID>/meta.json)
	SessionID         string
	RunnerSessionID   string
	TerminalSessionID string
	MetaStatus        string
	Workspace         string
	Model             string
	InitialPrompt     string
	SeedMeta          bool
	WriteRegistry     bool
	RegistryPID       int
	RegistryClosedPort bool

	// Live send path (MODE=send ignore detach)
	StartFakePTYWrap     bool
	FakePTYWrapPort      int
	FakePTYWrapScrollback string
	FakePTYInjectLog     *[]string
	FollowupPrompt       string

	// Soft-bind isolation
	NoGrokHomeEnv bool
}

type RegistryEntry struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
}

type Response struct {
	Stdout         string
	Stderr         string
	ExitCode       int
	Err            error
	SessionID      string
	TerminalID     string
	RegistryEntry  *RegistryEntry
	MetaAfter      map[string]any
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
