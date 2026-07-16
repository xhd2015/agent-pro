# agent-run run --resume-from-grok-session

Doc-style tests for **import-by-Grok-session** on
`agent-run run --resume-from-grok-session <id>`.

**Phases**

| Phase | Scope | Leaves |
|-------|--------|--------|
| **P1** | Validation / error matrix | `empty-id`, `missing-grok-session`, wrong runners, `already-mapped`, `dir-mismatch` |
| **P2** | Headless create + pre-bind + `grok --resume` argv | `headless-resume-argv`, `creates-session-meta`, `session-id-already-exists` |
| **P3** | Help, `--detach`, `--open`, mutex with `--auto-send-or-resume` | `help-documents-flag`, `detach-import`, `open-import`, `mutex-auto-send-or-resume` |

Classic TDD for P3: P1+P2 stay **GREEN**. Open/detach/mutex leave wiring may be incomplete — new P3 product leaves stay **RED** until implementer wires `Open`/`Detach` (and mutex) on the import path. Help may already be GREEN if `run --help` documents the flag.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI (`run`)** — `--resume-from-grok-session <id>` imports an external
  Grok CLI session: validate → CreateSession pre-bind → launch (headless / open /
  detach) with provider `--resume <id>`.
- **Flag value** — non-empty provider session UUID (Grok on-disk id).
- **Runner gate** — omitted `--agent-runner` → `grok-tty`; if set must be exactly
  `grok-tty`.
- **GROK_HOME** — isolated via env; `sessions/<url.PathEscape(cwd)>/<id>/summary.json`.
- **Agent-run store** — flat `sessions/<id>/meta.json` with pre-bound `runner_session_id`.
- **Modes (same exclusivity as normal `run`)** —
  - default headless (P2)
  - `--detach` — keep-alive daemon; print `session-id:` / `terminal-id:`; soft bind OK
  - `--open` — keep-alive + interactive attach (tests use `AGENT_RUN_OPEN_ATTACH_INSTANT=1`)
  - `--detach` and `--open` mutually exclusive (existing run rules)
- **Mutex** — `--resume-from-grok-session` must not combine with `--auto-send-or-resume`.
- **Fake hold runner** — `--agent-runner-binary` that paints `GROK_TTY_BANNER` and sleeps
  long; without detach/open parent would block past test timeout → proves mode wiring.

**Behaviors**

```
# P1 validation (error matrix) — unchanged
# P2 headless create + --resume argv — unchanged

# P3 help
run --help -> documents --resume-from-grok-session; trailing newline

# P3 detach import
seed Grok UUID; hold fake binary
  run --detach --session-id FIXED --agent-runner-binary HOLD
      --resume-from-grok-session UUID
  -> exit 0 (parent returns after registry; does not wait for hold sleep)
  -> stdout: session-id: … / terminal-id: …
  -> meta FIXED: runner_session_id=UUID

# P3 open import
seed Grok UUID; hold fake binary; AGENT_RUN_OPEN_ATTACH_INSTANT=1
  run --open --session-id FIXED --agent-runner-binary HOLD
      --resume-from-grok-session UUID ["prompt"]
  -> exit 0 before hold sleep ends
  -> meta FIXED: runner_session_id=UUID

# P3 mutex
run --auto-send-or-resume --session-id X --resume-from-grok-session UUID
  -> exit ≠ 0; mutually exclusive
```

## Version

0.0.2

## Decision Tree

```
run/resume-from-grok-session/          # nested DOCTEST root
├── DOCTEST.md
├── SETUP.md
│
│  # P1 — validation failure mode
├── empty-id/
├── missing-grok-session/
├── wrong-runner-codex/
├── wrong-runner-grok/
├── already-mapped/
├── dir-mismatch/
│
│  # P2 — headless import outcome
├── headless-resume-argv/
├── creates-session-meta/
├── session-id-already-exists/
│
│  # P3 — product modes + help + mutex
├── help-documents-flag/               # run --help lists the flag
├── detach-import/                     # --detach import smoke
├── open-import/                       # --open + instant-attach hook
└── mutex-auto-send-or-resume/         # exclusive with --auto-send-or-resume
```

Parameter ranking (most → least significant):

1. **Outcome class** — validation (P1) vs headless success/collision (P2) vs mode/help/mutex (P3)
2. **Launch mode** (P3) — help vs detach vs open vs mutex reject
3. **Assertion focus** — ids on stdout (detach) vs meta bind vs exclusive error text
4. **Fixtures** — hold binary duration, fixed session id, open instant env

## Test Index

| # | Leaf | Phase | Description |
|---|------|-------|-------------|
| 1 | `empty-id` | P1 | `--resume-from-grok-session=` → exit ≠ 0 |
| 2 | `missing-grok-session` | P1 | Empty `GROK_HOME`; unknown UUID → exit 1, not found |
| 3 | `wrong-runner-codex` | P1 | `--agent-runner codex-tty` → exit 1, requires grok-tty |
| 4 | `wrong-runner-grok` | P1 | `--agent-runner grok` → exit 1, requires grok-tty |
| 5 | `already-mapped` | P1 | Mapped `runner_session_id` → exit 1, already mapped |
| 6 | `dir-mismatch` | P1 | `--dir` ≠ Grok `info.cwd` → exit 1 |
| 7 | `headless-resume-argv` | P2 | Valid import + argv-recorder → exit 0; `--resume UUID` in probe |
| 8 | `creates-session-meta` | P2 | Valid import → meta `runner=grok-tty`, `runner_session_id`=UUID |
| 9 | `session-id-already-exists` | P2 | `--session-id` already in store → exit 1 |
| 10 | `help-documents-flag` | P3 | `run --help` lists `--resume-from-grok-session` |
| 11 | `detach-import` | P3 | import + `--detach` → exit 0; both ids; meta pre-bound |
| 12 | `open-import` | P3 | import + `--open` + instant attach → exit 0; meta pre-bound |
| 13 | `mutex-auto-send-or-resume` | P3 | flag + `--auto-send-or-resume` → exit ≠ 0 exclusive |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/resume-from-grok-session
doctest test ./cmd/agent-run/tests/run/resume-from-grok-session
# P3 only
doctest test ./cmd/agent-run/tests/run/resume-from-grok-session/help-documents-flag
doctest test ./cmd/agent-run/tests/run/resume-from-grok-session/detach-import
doctest test ./cmd/agent-run/tests/run/resume-from-grok-session/open-import
doctest test ./cmd/agent-run/tests/run/resume-from-grok-session/mutex-auto-send-or-resume
```

```go
import (
	"os/exec"
	"testing"
	"time"
)

type Request struct {
	RepoRoot          string
	TempDir           string
	Home              string // AGENT_RUN_HOME
	GrokHome          string // GROK_HOME
	AgentRun          string
	Args              []string
	Env               []string
	WorkDir           string // process cwd for CLI (cmd.Dir)
	GrokSessionID     string // provider UUID under test
	GrokCWD           string // summary info.cwd (absolute)
	DirFlag           string // --dir value when set
	AgentRunner       string // --agent-runner value; empty = omit flag
	SessionID         string // agent-run --session-id when set
	AgentRunnerBinary string // --agent-runner-binary path
	ArgvProbePath     string // fake runner argv probe file
	RunnerScriptPath  string
	FollowupPrompt    string
	MappedSessID      string // agent-run session_id for already-mapped seed
	DetachFlag        bool   // append --detach
	OpenFlag          bool   // append --open
	AutoSendOrResume  bool   // append --auto-send-or-resume
	OpenInstantAttach bool   // AGENT_RUN_OPEN_ATTACH_INSTANT=1
	ExecTimeout       time.Duration
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
