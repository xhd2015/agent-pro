# agent-run run --color (TTY child color env force)

Classic TDD doctests for **P1**: public contract of `agent-run run --color`.

When Color is on, the TTY agent-runner **child** process env is forced last
(same argv/env path family as session `-e` / prepend-path):

| Action | Detail |
|--------|--------|
| Unset | `NO_COLOR` |
| Set | `FORCE_COLOR=1`, `CLICOLOR=1`, `CLICOLOR_FORCE=1` |
| TERM | If effective TERM is empty or `dumb` → `TERM=xterm-256color`; else leave TERM |

**Not** persisted on `meta.json`. Does **not** recolor agent-run own stdout/`--json`.
TTY runners only; non-TTY + `--color` → hard error. Without `--color`: no force
from this feature.

Library mirrors (separate roots): `agentrunapi.FollowUpOpts.Color` → `--color` in
`BuildFollowUpCommand`; `agentrunbridge.RunOpts.Color` → `BuildArgs` argv;
`agentrunapi.Opts.Color` honors the same child env on production run paths
(sealed here via CLI).

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run --color` (boolean; help documents it). Parent process
  env for the CLI under test is controlled via `cmd.Env` (not suite `Setenv`).
- **TTY PTY child** — agent runner spawned under TTY path; color policy applied
  **last** on child env (wins over parent env and over user `-e NO_COLOR=…`).
- **Env-logging fake runner** — shell script via `--agent-runner-binary`; dumps
  `NO_COLOR|FORCE_COLOR|CLICOLOR|CLICOLOR_FORCE|TERM` (and PATH) to a probe file,
  prints grok-tty banner so the run completes.
- **Non-TTY runner (fake-codex)** — must hard-error when `--color` is set
  (same family as `-e` / `--prepend-path`).
- **Session meta** — must **not** store a color flag from this feature.

**Behaviors**

```
# help
agent-run run --help -> documents --color

# color ON (TTY + env-logger)
parent NO_COLOR=1 + run --color
  -> child: NO_COLOR unset; FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1

parent TERM=dumb + --color
  -> child TERM=xterm-256color

parent TERM=screen-256color + --color
  -> child TERM left as screen-256color (not blindly rewritten)

--color -e NO_COLOR=1
  -> color wins: child still has NO_COLOR unset

# color ON + non-TTY
--agent-runner fake-codex --color -> exit ≠ 0; TTY-only style error

# color OFF
run without --color (parent NO_COLOR / FORCE_COLOR=0 / CLICOLOR*=0)
  -> no FORCE_COLOR=1 / CLICOLOR=1 force; NO_COLOR not cleared by this feature
  # (TERM may still be set by TTY/PTY layers; color-ON TERM policy is on/* only)
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run-color/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, AGENT_RUN_HOME, env-logger
├── help/                                 # CLI surface
│   └── run-lists-color/                  # H1: run --help mentions --color
├── on/                                   # --color present (force policy)
│   ├── parent-nocolor/                   # C1: parent NO_COLOR cleared + force keys
│   ├── parent-term-dumb/                 # C2: dumb → xterm-256color
│   ├── parent-term-good/                 # C3: good TERM preserved
│   ├── wins-over-e-nocolor/              # C4: --color beats -e NO_COLOR=1
│   └── non-tty-rejected/                 # C6: fake-codex + --color hard error
└── off/                                  # without --color (baseline / no force)
    └── no-force/                         # C5: no force keys; TERM not rewritten
```

Parameter ranking (most → least significant):

1. **CLI surface vs execute** — help discovery vs run behavior
2. **Color ON vs OFF** — presence of `--color` (primary behavioral split)
3. **Within ON** — parent env factor (NO_COLOR / TERM dumb / TERM good / `-e` conflict) vs non-TTY reject
4. **Within OFF** — baseline no-force under hostile parent env

## Test Index

| # | Leaf | Covers | Description |
|---|------|--------|-------------|
| 1 | `help/run-lists-color` | H1 | `run --help` documents `--color` |
| 2 | `on/parent-nocolor` | C1 | Parent `NO_COLOR=1` + `--color` → child unsets `NO_COLOR`; sets force trio |
| 3 | `on/parent-term-dumb` | C2 | Parent `TERM=dumb` + `--color` → child `TERM=xterm-256color` |
| 4 | `on/parent-term-good` | C3 | Parent good TERM + `--color` → TERM not rewritten to `xterm-256color` |
| 5 | `on/wins-over-e-nocolor` | C4 | `--color -e NO_COLOR=1` → color wins (`NO_COLOR` still unset) |
| 6 | `on/non-tty-rejected` | C6 | `fake-codex` + `--color` → non-zero; TTY-only style error |
| 7 | `off/no-force` | C5 | Without `--color`: no `FORCE_COLOR=1` / CLICOLOR force; `NO_COLOR` not cleared |

## How to Run

```sh
# Discovery skips labeled e2e leaves by default.
doctest test ./cmd/agent-run/tests/run-color
doctest test --label e2e ./cmd/agent-run/tests/run-color
doctest test --label-all ./cmd/agent-run/tests/run-color

doctest vet ./cmd/agent-run/tests/run-color
doctest test ./cmd/agent-run/tests/run-color

doctest test -v ./cmd/agent-run/tests/run-color/help/run-lists-color
doctest test -v ./cmd/agent-run/tests/run-color/on/parent-nocolor
doctest test -v ./cmd/agent-run/tests/run-color/on/parent-term-dumb
doctest test -v ./cmd/agent-run/tests/run-color/on/parent-term-good
doctest test -v ./cmd/agent-run/tests/run-color/on/wins-over-e-nocolor
doctest test -v ./cmd/agent-run/tests/run-color/on/non-tty-rejected
doctest test -v ./cmd/agent-run/tests/run-color/off/no-force
```

Expect **RED** on `help/` and `on/*` until implementer lands `--color` + child env
policy. `off/no-force` may already GREEN (baseline seal).

Library companions (isolated nested roots; RED until Color fields land):

```sh
doctest test -v ./tests/agentrunapi/follow-up-color/color-true
doctest test -v ./tests/agentrunapi/follow-up-color/color-false
doctest test -v ./tests/agentrunbridge/color-flag/emits-color
```

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot string
	TempDir  string
	Home     string
	AgentRun string
	Args     []string
	// Env is extra KEY=VALUE for the agent-run process (cmd.Env last-win).
	// Color-related parent keys are stripped from the host base then re-applied
	// from this slice so leaves fully control NO_COLOR / FORCE_* / CLICOLOR* / TERM.
	Env []string

	ExecTimeout time.Duration

	// TTY fake runner + env probe
	EnvProbePath      string
	RunnerScriptPath  string
	AgentRunnerBinary string
	Prompt            string
	SessionID         string

	// Color flag request (leaves append --color to Args when true)
	Color bool

	// Optional parent TERM / NO_COLOR for the agent-run process under test
	ParentTERM    string // if non-empty, set TERM= on agent-run cmd.Env
	ParentNoColor bool   // if true, set NO_COLOR=1 on agent-run cmd.Env
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
