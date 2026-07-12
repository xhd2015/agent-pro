# agent-run session env + prepend_paths

Doc-style tests for session-scoped environment injection on TTY agent runners:

- **`--prepend-path=DIR`** (repeatable) — prepend DIR(s) to the agent runner child `PATH`
- **`-e` / `--env KEY=VALUE`** (repeatable) — set env vars on the agent runner child `Env`
- **`--agent-runner-config-home PATH`** — already exists for CLI; must **persist** on session meta
  and **reapply on resume** (same survival model)

All three survive **resume**: stored on `meta.json`, reapplied when the TTY runner is
re-invoked. Resume CLI flags for paths/env are **additional only** (append to stored
values). Non-TTY runners reject these flags with a hard error.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` / `resume` (and `--auto-send-or-resume` create/resume paths)
  accept `--prepend-path`, `-e`/`--env`, and `--agent-runner-config-home`. Help text
  documents the flags on both subcommands.
- **Session storage** — `AGENT_RUN_HOME/sessions/<session_id>/meta.json` holds
  `prepend_paths` (ordered abs paths), `env` (ordered `KEY=VALUE`), and
  `agent_runner_config_home` (abs scalar).
- **TTY PTY child** — agent runner process spawned under the TTY path. Effective
  `cmd.Env` is base `os.Environ()` plus meta/CLI env (last-win per key), with
  `PATH` prefixed by joined `prepend_paths`, plus `GROK_HOME`/`CODEX_HOME` from
  config home.
- **Env-logging fake runner** — shell script replacing `grok` via
  `--agent-runner-binary`; dumps selected env keys to a probe file and prints a
  grok-tty banner so the run completes.
- **Non-TTY runner (e.g. fake-codex)** — must reject `--prepend-path` / `-e`/`--env`
  with a hard error (TTY-only scope).

**Behaviors**

```
# help
agent-run run --help    -> documents --prepend-path, -e/--env
agent-run resume --help -> same flags

# run (TTY only)
agent-run run --agent-runner grok-tty \
  --prepend-path DIR... -e KEY=VALUE... --agent-runner-config-home PATH \
  --agent-runner-binary <env-logger> "prompt"
  -> child PATH starts with abs DIR(s)
  -> child env has KEY=VALUE (last-win per KEY)
  -> child has GROK_HOME=PATH (not necessarily AGENT_RUNNER_CONFIG_HOME=)
  -> meta.json prepend_paths / env / agent_runner_config_home persisted (abs)

# soft allow
--prepend-path /missing/dir -> exit 0; path still stored and applied to PATH

# non-TTY hard error
--agent-runner fake-codex + --prepend-path|-e -> exit ≠ 0; TTY-only / unsupported

# validation
-e FOO (no =) -> exit ≠ 0
-e =bar (empty key) -> exit ≠ 0
FOO= empty value allowed

# resume
seed/stored meta.prepend_paths + meta.env + meta.agent_runner_config_home
  -> resume without flags reapplies child PATH/env/GROK_HOME
  -> resume --prepend-path /more appends to stored + persists
  -> resume -e NEW=1 appends (last-win) + persists
  -> resume without --agent-runner-config-home keeps stored scalar
```

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/session-env/
├── DOCTEST.md
├── SETUP.md                              # build agent-run, isolated AGENT_RUN_HOME, helpers
├── help/                                 # CLI surface (most discoverable)
│   ├── run-lists-flags/                  # H1: run --help
│   └── resume-lists-flags/               # H2: resume --help
├── run/                                  # create path (TTY child env + meta write)
│   ├── combined-child-and-meta/          # R1+R2+R4+S1: prepend+env+config-home
│   ├── env-last-win/                     # R3: --env A=1 -e A=2 → A=2
│   ├── missing-prepend-soft/             # R5: missing dir soft-allow
│   └── non-tty-rejected/                 # R6: fake-codex + flags hard error
├── resume/                               # reapply / append on resume
│   ├── reapplies-stored/                 # S2+S5: no extra flags → stored values
│   ├── append-prepend-path/              # S3: extra --prepend-path grows list
│   └── append-env/                       # S4: extra -e grows env list
└── validation/                           # flag parse errors
    ├── env-missing-equals/               # V1: -e FOO
    └── env-empty-key/                    # V2: -e =bar
```

Parameter ranking (most → least significant):

1. **CLI surface vs execute vs validation** — help discovery vs behavioral run vs hard errors
2. **Operation mode** — `run` (create + apply + persist) vs `resume` (reapply + optional append)
3. **Within run** — happy combined env contract vs last-win vs soft missing dir vs non-TTY reject
4. **Within resume** — reapply-only vs append prepend vs append env
5. **Within validation** — missing `=` vs empty key

## Test Index

| # | Leaf | Covers | Description |
|---|------|--------|-------------|
| 1 | `help/run-lists-flags` | H1 | `run --help` mentions `--prepend-path` and `-e`/`--env` |
| 2 | `help/resume-lists-flags` | H2 | `resume --help` mentions same flags |
| 3 | `run/combined-child-and-meta` | R1,R2,R4,S1 | PATH prefix + FOO + GROK_HOME on child; meta has all three fields (abs) |
| 4 | `run/env-last-win` | R3 | `--env A=1 -e A=2` → child `A=2` |
| 5 | `run/missing-prepend-soft` | R5 | nonexistent prepend dir: exit 0; PATH still contains abs path |
| 6 | `run/non-tty-rejected` | R6 | non-TTY + `--prepend-path` or `-e` → non-zero; TTY-only message |
| 7 | `resume/reapplies-stored` | S2,S5 | resume without flags reapplies stored PATH/FOO/GROK_HOME |
| 8 | `resume/append-prepend-path` | S3 | resume `--prepend-path` appends; PATH order stored then new; meta grows |
| 9 | `resume/append-env` | S4 | resume `-e NEW=1` adds env; meta grows |
| 10 | `validation/env-missing-equals` | V1 | `-e FOO` → non-zero clear error |
| 11 | `validation/env-empty-key` | V2 | `-e =bar` → non-zero clear error |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/session-env
doctest test ./cmd/agent-run/tests/session-env

doctest test -v ./cmd/agent-run/tests/session-env/help/run-lists-flags
doctest test -v ./cmd/agent-run/tests/session-env/help/resume-lists-flags
doctest test -v ./cmd/agent-run/tests/session-env/run/combined-child-and-meta
doctest test -v ./cmd/agent-run/tests/session-env/run/env-last-win
doctest test -v ./cmd/agent-run/tests/session-env/run/missing-prepend-soft
doctest test -v ./cmd/agent-run/tests/session-env/run/non-tty-rejected
doctest test -v ./cmd/agent-run/tests/session-env/resume/reapplies-stored
doctest test -v ./cmd/agent-run/tests/session-env/resume/append-prepend-path
doctest test -v ./cmd/agent-run/tests/session-env/resume/append-env
doctest test -v ./cmd/agent-run/tests/session-env/validation/env-missing-equals
doctest test -v ./cmd/agent-run/tests/session-env/validation/env-empty-key
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

	ExecTimeout time.Duration

	// TTY fake runner + env probe
	EnvProbePath      string
	RunnerScriptPath  string
	AgentRunnerBinary string
	Prompt            string

	// Paths under test (absolute under TempDir when set by Setup)
	PrependPathDir        string
	PrependPathMore       string
	AgentRunnerConfigHome string
	SessionID             string

	// Resume seed (bound + exited meta under flat sessions/<SessionID>/)
	SeedMeta          bool
	Runner            string
	MetaStatus        string
	RunnerSessionID   string
	TerminalSessionID string
	Workspace         string
	Model             string
	InitialPrompt     string
	// Seeded session-scoped env injection fields (written raw into meta.json)
	SeedPrependPaths []string
	SeedEnv          []string
	SeedConfigHome   string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
