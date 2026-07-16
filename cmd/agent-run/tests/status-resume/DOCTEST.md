# agent-run session status + resume + run --open grok bind

Doc-style tests for session-level `agent-run status <session-id>`,
`agent-run resume`, and `agent-run run --open` grok session discovery/bind
(including background bind that continues after detach). Prefer stub fixtures
(seeded meta, registry, fake ptywrap, argv-recording fake runner, delayed
`GROK_HOME` materialization) over real grok.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — dispatches `status`, `resume`, and `run --open`. Status is
  a multi-layer probe (not a meta dump only). Resume is a **shortcut of `run`**
  that reuses session meta and injects provider resume (`grok --resume <id>`).
- **Session storage** — flat `AGENT_RUN_HOME/sessions/<session_id>/meta.json`
  (`SessionMeta`: runner, session_id, runner_session_id, terminal_session_id,
  status, workspace, model, …). Optional durable bind progress (e.g.
  `bind.json` with `in_progress|ok|failed`) so concurrent status can observe
  mid-open binding. Secondary lookup by explicit `--grok-session-id` matches
  `meta.runner_session_id` when `meta.runner` is exactly `grok` or `grok-tty`.
- **TTY registry** — `AGENT_RUN_HOME/grok-tty-registry/<terminal_session_id>.json`
  (`pid`, `listen_addr`, `created_at`). Absent or unreachable ⇒ process/terminal
  dead layers.
- **Keep-alive serve process** — registry PID for process layer (`alive|dead|unknown`).
- **ptywrap / terminal** — TCP reachability, screen status, sendable (tests use
  fake in-process HTTP+WebSocket when a live terminal is required).
- **Provider runner (grok)** — bound when `meta.runner_session_id` is non-empty;
  **`runner.exited`** is true when the provider agent has exited: terminal
  unreachable/missing + bound, **or** keep-alive serve still TCP-reachable but
  agent gone (exit markers in scrollback / `grok --resume` footer / not sendable
  after exit / no child under serve). False only when truly live (sendable yes /
  idle prompt without exit markers / child alive). Do **not** treat
  `terminal.status == reachable` alone as still running.
  While open bind is in flight, runner status may be **`binding`**.
- **Background bind worker** — for `run --open` + `grok-tty`, starts DiscoverSession
  as soon as the open session is up (after PTY start / prompt inject), without
  blocking attach; persists `runner_session_id` when found; joined after detach.
- **Fake runner / hooks** — `AGENT_RUN_GROK_TTY_COMMAND`,
  `--agent-runner-binary` argv recorder, `AGENT_RUN_OPEN_ATTACH_INSTANT=1`,
  `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` + temp `GROK_HOME` for discovery (including
  delayed materialization of `updates.jsonl`). Isolated `HOME` without `GROK_HOME`
  env proves hard-require from non-empty open prompt alone (O1). Session may be
  seeded under a cwd ≠ agent-run workspace so prompt-only fallback binds (O3).

**Behaviors**

```
# help
agent-run --help -> lists resume
agent-run status --help -> documents session-ref / multi-layer probe / --grok-session-id
agent-run resume --help -> lists --open, session-id, followup, --grok-session-id

# status
agent-run status -> home: <path>\n  (compat bare)
agent-run status <session-id> -> multi-layer probe
agent-run status <runner>/<session_id> -> unambiguous resolve (legacy leaf)
agent-run status --grok-session-id ID -> meta-only resolve (runner grok|grok-tty)
  # mutex with positional; 0 matches not found; 2+ ambiguous; non-grok never match
  # CLI --agent-runner ignored for this lookup
agent-run status --json <ref> -> JSON mirror (runner.exited, resume.ready)

# layers (human + JSON)
session / process / terminal / runner (binding|bound|unbound + exited) / resume (ready)

# runner.exited heuristics (status probe)
live idle/sendable terminal -> exited false, resume not ready
zombie serve after /exit (reachable + exit scrollback / sendable no) -> exited true, resume ready when bound
bound + terminal missing|unreachable|process dead -> exited true, resume ready
unbound -> resume not ready

# resume gate
resume ready ⇔ runner_session_id non-empty ∧ runner.exited == true
  exited false (live) -> deny; hint send (not "already in use")
  unbound -> deny
  missing session -> exit 1
  --no-submit without --open -> exit ≠ 0; requires --open

# resume ≡ run shortcut
agent-run resume [flags] <session-id> ["followup"]
agent-run resume [flags] --grok-session-id ID ["followup"]
  ≡ agent-run run [flags] --session-id=<id> --agent-runner=<meta.runner> …
    with ResumeSessionID = meta.runner_session_id
    argv includes grok --resume <id> (when not overridden by TTY command hook)
  # --grok-session-id and positional session-id are mutually exclusive

# resume + zombie terminal registry (after /exit keep-alive still holds id)
bound+exited + registry live (zombie serve / exit scrollback):
  -> reclaim zombie terminal id (tear down serve + remove registry)
  -> reserve same terminal id (primary)
  -> if reclaim fails: allocate new terminal id, update meta.terminal_session_id
  -> must NOT fail: session id "…" already in use
live agent: do not reclaim; error steers to send

# run --open post-exit (legacy finalize path)
after attach returns:
  discovery OK -> stderr:
    grok-tty: <terminal-or-session-id>
    grok-tty: grok session <uuid>
    grok-tty: grok updates <path>
    and persist meta.runner_session_id
  discovery fail -> exit ≠ 0; stderr error "grok session id not resolved"

# run --open background bind (new contract)
run --open:
  create session -> start keep-alive TTY + inject prompt
  -> start background bind worker (DiscoverSession; long budget)
  -> attach (foreground; instant hook in tests)
       concurrent status may see runner: binding then bound
  -> attach returns (detach)
  -> ALWAYS join/wait for bind worker (never soft-exit unbound while pending)
  -> print post-exit report from bind result
  -> exit 0 if bound; exit ≠ 0 if hard-require bind failed
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
│   ├── status-documents-session-ref/           # H2 status --help session-ref / layers / --grok-session-id
│   └── resume-lists-open/                      # H3 resume --help lists --open + session-id + --grok-session-id
├── status/                                     # agent-run status [ref]
│   ├── bare-home/                              # S0 bare status → home path
│   ├── missing-session/                        # S1 unknown id → exit 1
│   ├── session-ref/
│   │   ├── runner-slash-id/                    # S6 runner/session resolves uniquely
│   │   └── grok-session-id/                    # explicit --grok-session-id meta-only lookup
│   │       ├── found-grok-tty/                 # G1 resolve grok-tty (+ --agent-runner ignored)
│   │       ├── found-runner-grok/              # G2 resolve runner=grok
│   │       ├── missing/                        # G3 0 matches → not found
│   │       ├── ambiguous/                      # G4 2+ matches → ambiguous + both ids
│   │       ├── ignores-non-grok/               # G5 codex-tty UUID must not resolve
│   │       ├── mutex-with-positional/          # G6 flag + positional exclusive
│   │       └── empty-flag/                     # G7 empty --grok-session-id= → exit ≠ 0
│   └── layers/                                 # multi-layer probe outcomes
│       ├── bound-exited-resume-ready/          # S2 dead/missing terminal + bound + exited → ready
│       ├── zombie-serve-exited-resume-ready/   # E1 CRITICAL RED: reachable zombie serve → exited true
│       ├── zombie-serve-json/                  # E1 JSON: alive+reachable but exited true + resume ready
│       ├── live-bound-not-exited/              # S3/E2 live sendable → exited false, not ready
│       ├── unbound-not-ready/                  # S4 no runner_session_id → unbound
│       └── json-shape/                         # S5 --json has runner.exited + resume.ready (dead term)
├── resume/                                     # agent-run resume
│   ├── denied/
│   │   ├── not-exited/                         # live bound → exit 1, hint send (not already-in-use)
│   │   ├── unbound/                            # no runner_session_id
│   │   ├── missing-session/                    # session not found
│   │   ├── missing-prompt/                     # no-followup reopen (not prompt-required)
│   │   └── no-submit-without-open/             # R4 --no-submit requires --open
│   ├── zombie-reclaim/                         # bound+exited but registry holds id
│   │   ├── open-hello/                         # R2 CRITICAL: resume --open "hello" reclaims
│   │   └── headless-followup/                  # R1 headless followup reclaims + --resume argv
│   ├── headless-followup-when-exited/          # dead terminal; --resume in argv
│   ├── by-grok-session-id/                     # G8 resume via --grok-session-id (no positional)
│   └── open-flag-accepted/                     # --open known (not unknown-flag)
├── run-open-post-exit/                         # run --open post-attach finalize (B5 regression)
│   ├── prints-grok-session-when-resolved/      # O1 bind + stderr + meta persist
│   └── errors-when-unresolved/                 # O2 exit ≠ 0, not-resolved error
└── run-open-background-bind/                   # background bind + always wait after detach
    ├── success-preseeded/                      # B1 bind ready before/during open → session + meta
    ├── detach-before-bind-always-waits/        # B2 CRITICAL delayed discover past attach
    ├── hard-require-without-grok-home-env/     # O1 non-empty prompt hard-wait (no GROK_HOME env)
    ├── hard-fail-after-full-wait/              # B3 no GROK material → exit ≠ 0
    ├── prompt-fallback-cwd-mismatch/           # O3 DiscoverSession prompt scan when cwd ≠ workspace
    └── status-binding-while-open/              # B4 mid-open status binding|bound
```

Parameter ranking (most → least significant):

1. **Command surface** — help | status | resume | run-open-post-exit | run-open-background-bind
2. **Status invocation** — bare home | missing | session-ref form | multi-layer state
3. **Session-ref form** — compound `runner/id` | `--grok-session-id` (match cardinality / runner allowlist / mutex / empty)
4. **Runner/resume state** — bound+exited (dead term) | zombie serve exited | live not-exited | unbound | JSON format
5. **Resume outcome** — denied reason | zombie reclaim (open / headless) | clean-term success | by-grok-session-id | `--open` flag
6. **Resume flag gates** — `--no-submit` requires `--open`; live deny steers to send
7. **Open discovery outcome** — resolved vs unresolved (post-exit finalize)
8. **Background-bind timing vs detach** — preseeded success | detach-before-bind wait |
   hard fail after wait | concurrent status binding
9. **Hard-require trigger / discovery key** — non-empty prompt without GROK_HOME env (O1) |
   prompt-only match when session cwd ≠ agent-run workspace (O3)

## Test Index

| # | Leaf | Req | Description |
|---|------|-----|-------------|
| 1 | `help/top-level-lists-resume` | H1 | `agent-run --help` stdout contains `resume` |
| 2 | `help/status-documents-session-ref` | H2 | `status --help` mentions session id / layers / `--grok-session-id` |
| 3 | `help/resume-lists-open` | H3 | `resume --help` lists `--open`, session-id, `--grok-session-id` |
| 4 | `status/bare-home` | S0 | bare `status` → exit 0, `home: …\n` |
| 5 | `status/missing-session` | S1 | unknown session → exit 1, not found |
| 6 | `status/session-ref/runner-slash-id` | S6 | `grok-tty/<id>` resolves same session as bare id |
| 6a | `status/session-ref/grok-session-id/found-grok-tty` | G1 | `--grok-session-id` resolves `grok-tty`; CLI `--agent-runner` ignored |
| 6b | `status/session-ref/grok-session-id/found-runner-grok` | G2 | `--grok-session-id` resolves exact `runner=grok` |
| 6c | `status/session-ref/grok-session-id/missing` | G3 | no matching meta → exit 1 not found |
| 6d | `status/session-ref/grok-session-id/ambiguous` | G4 | two matches → exit 1; both agent-run ids listed |
| 6e | `status/session-ref/grok-session-id/ignores-non-grok` | G5 | `codex-tty` UUID must not resolve |
| 6f | `status/session-ref/grok-session-id/mutex-with-positional` | G6 | flag + positional → exit 1 exclusive |
| 6g | `status/session-ref/grok-session-id/empty-flag` | G7 | empty `--grok-session-id=` → exit ≠ 0 |
| 7 | `status/layers/bound-exited-resume-ready` | S2/E3 | bound + exited + dead/missing terminal → `resume.ready: yes` |
| 8 | `status/layers/zombie-serve-exited-resume-ready` | E1 | **CRITICAL RED**: zombie serve (alive+reachable+exit scrollback) → `exited: true`, resume ready |
| 9 | `status/layers/zombie-serve-json` | E1 | zombie serve `--json` → `runner.exited` true + `resume.ready` true while process alive |
| 10 | `status/layers/live-bound-not-exited` | S3/E2/E4 | live idle/sendable → `exited: false`, not ready, hint send |
| 11 | `status/layers/unbound-not-ready` | S4 | empty `runner_session_id` → unbound, not ready |
| 12 | `status/layers/json-shape` | S5 | `--json` includes `runner.exited` and `resume.ready` (dead term seed) |
| 13 | `resume/denied/not-exited` | live | live bound → exit 1, cannot resume / use send; **not** already-in-use |
| 14 | `resume/denied/unbound` | unbound | unbound → exit 1, not bound |
| 15 | `resume/denied/missing-session` | missing | missing meta → exit 1, not found |
| 16 | `resume/denied/missing-prompt` | reopen | exited bound, no followup, no `--open` → not "prompt required" |
| 17 | `resume/denied/no-submit-without-open` | R4 | `resume --no-submit` without `--open` → exit ≠ 0, requires `--open` |
| 18 | `resume/zombie-reclaim/open-hello` | R2 | **CRITICAL RED**: zombie registry + `resume --open id "hello"` → reclaim, not already-in-use |
| 19 | `resume/zombie-reclaim/headless-followup` | R1 | zombie registry + headless followup → reclaim + `--resume` in argv |
| 20 | `resume/headless-followup-when-exited` | clean | dead terminal + followup → run path with `--resume <id>` |
| 20a | `resume/by-grok-session-id` | G8 | `resume --grok-session-id` headless followup; argv has `--resume` UUID |
| 21 | `resume/open-flag-accepted` | open-flag | `resume --open …` not rejected as unknown flag |
| 22 | `run-open-post-exit/prints-grok-session-when-resolved` | O1/B5 | after attach, stderr grok session + meta.runner_session_id |
| 23 | `run-open-post-exit/errors-when-unresolved` | O2/B5 | discovery fail → exit ≠ 0, not resolved error |
| 24 | `run-open-background-bind/success-preseeded` | B1 | preseeded GROK session; open exits with session + meta |
| 25 | `run-open-background-bind/detach-before-bind-always-waits` | B2 | delayed discover after attach; wait ≥ delay; then bind |
| 26 | `run-open-background-bind/hard-require-without-grok-home-env` | O1 | non-empty prompt, **no** `GROK_HOME` env; delayed materialize; hard wait + bind |
| 27 | `run-open-background-bind/hard-fail-after-full-wait` | B3 | empty GROK material; exit ≠ 0; not resolved; no false bind |
| 28 | `run-open-background-bind/prompt-fallback-cwd-mismatch` | O3 | session under other encoded cwd; prompt match still binds |
| 29 | `run-open-background-bind/status-binding-while-open` | B4 | mid-open `status --json` shows runner binding\|bound |

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
doctest test -v ./cmd/agent-run/tests/status-resume/status/session-ref/grok-session-id
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/bound-exited-resume-ready
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/zombie-serve-exited-resume-ready
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/zombie-serve-json
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/live-bound-not-exited
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/unbound-not-ready
doctest test -v ./cmd/agent-run/tests/status-resume/status/layers/json-shape

# resume
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/not-exited
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/unbound
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/missing-session
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/missing-prompt
doctest test -v ./cmd/agent-run/tests/status-resume/resume/denied/no-submit-without-open
doctest test -v ./cmd/agent-run/tests/status-resume/resume/zombie-reclaim/open-hello
doctest test -v ./cmd/agent-run/tests/status-resume/resume/zombie-reclaim/headless-followup
doctest test -v ./cmd/agent-run/tests/status-resume/resume/headless-followup-when-exited
doctest test -v ./cmd/agent-run/tests/status-resume/resume/by-grok-session-id
doctest test -v ./cmd/agent-run/tests/status-resume/resume/open-flag-accepted

# run --open post-exit (regression)
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-post-exit/prints-grok-session-when-resolved
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-post-exit/errors-when-unresolved

# run --open background bind
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/success-preseeded
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/detach-before-bind-always-waits
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/hard-require-without-grok-home-env
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/hard-fail-after-full-wait
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/prompt-fallback-cwd-mismatch
doctest test -v ./cmd/agent-run/tests/status-resume/run-open-background-bind/status-binding-while-open
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
	//   ""                — plain CLI exec
	//   "status-json"     — parse stdout as JSON into Response.JSONBody
	//   "read-meta"       — after CLI, load sessions/<session>/meta.json
	//   "open-status-mid" — start open async, probe status mid-flight, wait open
	Mode        string
	ExecTimeout time.Duration

	// Session seed (meta.json under flat sessions/<SessionID>/; Runner is meta field)
	SeedMeta          bool
	Runner            string // default grok-tty
	SessionID         string
	MetaStatus        string // running | finished | error
	RunnerSessionID   string // provider (grok) session id — resume key / --grok-session-id
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
	// GrokMaterializeDelay: when >0, Run schedules writing the fake GROK session
	// dir (summary + updates.jsonl) after this delay from process start. Used to
	// force detach-before-bind (instant attach returns first; bind must wait).
	GrokMaterializeDelay time.Duration
	// OpenPrompt is the prompt text written into delayed updates.jsonl when
	// GrokMaterializeDelay > 0 (falls back to InitialPrompt / last Args token).
	OpenPrompt string
	// NoGrokHomeEnv: materialize/seed under GrokHome (and optional HOME) but do
	// **not** set GROK_HOME / AGENT_RUN_GROK_TTY_GROK_SESSION_ID / AGENT_RUNNER_CONFIG_HOME
	// on the child. Proves hard-require for non-empty open prompt alone (O1).
	NoGrokHomeEnv bool
	// GrokSessionCwd: when non-empty, seed/materialize session under this path's
	// encoded cwd and summary.info.cwd (may differ from agent-run Workspace).
	// Used by prompt-fallback when session cwd ≠ workspace (O3).
	GrokSessionCwd string

	// Resume argv probe (no AGENT_RUN_GROK_TTY_COMMAND — use binary path)
	ArgvProbePath     string
	RunnerScriptPath  string
	AgentRunnerBinary string
	FollowupPrompt    string

	// StatusProbeArgs for Mode "open-status-mid" (default: status --json <SessionID>).
	StatusProbeArgs []string
	// StatusProbeAfter is how long to wait after session meta appears before
	// probing status (default ~400ms). Keeps the probe inside the bind window.
	StatusProbeAfter time.Duration
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
	// MetaAfter is filled when Mode == "read-meta" or "open-status-mid".
	MetaAfter map[string]any
	// Elapsed is wall time of the primary CLI invocation (open path).
	Elapsed time.Duration
	// Mid-open status probe results (Mode == "open-status-mid").
	StatusProbeStdout string
	StatusProbeStderr string
	StatusProbeExit   int
	StatusProbeJSON   map[string]any
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
