# agent-run run --session-id-from-prompt Tests

Doc-style tests for `agent-run run --session-id-from-prompt` and the related same-id
policy for explicit `--session` on TTY runners.

`--session-id-from-prompt` generates a human-readable session id from the prompt slug
(`<slug>-YYYYMMDD-HHMMSS`, then `-1`, `-2`, … on collision) and uses that **same
string** for agent storage and, on TTY runners, the terminal registry id.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` parses `--session-id-from-prompt` / `--session`, resolves the
  prompt, chooses a session id, then invokes the selected runner.
- **Slug generator** — turns the prompt into a base slug (lowercase, non-alnum →
  `-`, collapse/trim `-`, empty → `sess` or `task`, truncate ≤ 128 runes), appends
  local `YYYYMMDD-HHMMSS`, then numeric suffixes on conflict.
- **Agent storage** — `AGENT_RUN_HOME/sessions/<runner>/<id>/` with `meta.json`
  (`session_id`, optional `terminal_session_id`) and event files.
- **Terminal registry** (TTY only) — e.g. `AGENT_RUN_HOME/grok-tty-registry/<id>.json`;
  stderr line `grok-tty: <id>` (or `codex-tty: <id>`).
- **fake-codex** — non-TTY deterministic runner for storage-only checks.
- **Fake TUI** — `AGENT_RUN_GROK_TTY_COMMAND` replaces `grok` for grok-tty PTY runs.

**Behaviors**

```
# auto: generate id, storage + (TTY) registry share it
agent-run run --session-id-from-prompt "prompt"
  -> slugify(prompt) + local timestamp [+ -N]
  -> sessions/<runner>/<id>/
  -> [TTY] registry <id>.json + stderr "<runner>: <id>"

# explicit: user-chosen id also used as TTY registry id
agent-run run --session my-task "prompt"
  -> sessions/<runner>/my-task/
  -> [TTY] registry my-task.json + stderr "<runner>: my-task"

# mutual exclusion
agent-run run --session X --session-id-from-prompt "p" -> error, exit ≠ 0
agent-run run --session-id X --session-id-from-prompt "p" -> error, exit ≠ 0

# --session-id alias of --session (AR1)
agent-run run --session-id my-task "prompt"
  -> same storage id as --session my-task

# help
agent-run run --help -> documents --session-id-from-prompt and --session-id
```

Conflict (id taken) when storage already has `sessions/<runner>/<id>/` (meta
exists), or for TTY a live registry entry exists for that id.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run/session-id-from-prompt/
├── DOCTEST.md
├── SETUP.md
├── help/
│   ├── run-help-lists-flag/              # run --help documents --session-id-from-prompt
│   └── run-help-lists-session-id/        # run --help documents --session-id alias (AR1)
├── flag-conflict/
│   ├── session-and-auto/                 # --session + --session-id-from-prompt → error
│   └── session-id-and-auto/              # --session-id + --session-id-from-prompt → error
├── auto-id/                              # --session-id-from-prompt set; --session unset
│   ├── non-tty/                          # fake-codex (storage only)
│   │   ├── storage-id-from-prompt-slug/  # storage dir + id shape from prompt
│   │   ├── punctuation-slug/             # "Hello, World!!" → hello-world-…
│   │   ├── long-prompt-truncates-base/   # base ≤ 128 runes before timestamp
│   │   ├── empty-slug-uses-fallback-base/# "!!!" → sess|task + timestamp
│   │   └── storage-collision-suffix/     # pre-seed storage → -N suffix
│   └── tty/                              # grok-tty (storage + registry same id)
│       ├── same-id-storage-registry-meta/# stderr, storage, meta, registry agree
│       └── storage-collision-suffix/     # TTY auto-id also gets -N on storage clash
└── explicit-session/                     # --session / --session-id set; auto unset
    ├── non-tty-storage-id/               # storage uses explicit --session id
    ├── non-tty-session-id-alias/         # AR1 --session-id same as --session
    └── tty-same-id-registry-meta/        # TTY registry/stderr/meta use same explicit id
```

Parameter ranking (most → least significant):

1. **Session id policy** — help | flag conflict | auto-id | explicit `--session`
2. **Runner class** — non-TTY (`fake-codex`) vs TTY (`grok-tty`)
3. **Slug / collision edge** — base derive, punctuation, length, empty slug, collision

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-flag` | `run --help` lists `--session-id-from-prompt`; stdout ends with `\n` |
| 2 | `help/run-help-lists-session-id` | `run --help` lists `--session-id` alias (AR1) |
| 3 | `flag-conflict/session-and-auto` | `--session` + auto → exit ≠ 0 |
| 4 | `flag-conflict/session-id-and-auto` | `--session-id` + auto → exit ≠ 0 |
| 5 | `auto-id/non-tty/storage-id-from-prompt-slug` | Storage under `sessions/fake-codex/<slug-ts…>/`; id matches shape |
| 6 | `auto-id/non-tty/punctuation-slug` | `"Hello, World!!"` → base `hello-world` + timestamp |
| 7 | `auto-id/non-tty/long-prompt-truncates-base` | Base portion ≤ 128 runes before `-YYYYMMDD-HHMMSS` |
| 8 | `auto-id/non-tty/empty-slug-uses-fallback-base` | Punctuation-only prompt → base `sess` or `task` |
| 9 | `auto-id/non-tty/storage-collision-suffix` | Pre-seed storage for nearby timestamps → id ends with `-N` |
| 10 | `auto-id/tty/same-id-storage-registry-meta` | stderr `grok-tty: <id>` == storage dir == meta == registry |
| 11 | `auto-id/tty/storage-collision-suffix` | TTY auto-id collision appends `-N`; storage and registry share it |
| 12 | `explicit-session/non-tty-storage-id` | `--session my-task` → `sessions/fake-codex/my-task/` |
| 13 | `explicit-session/non-tty-session-id-alias` | AR1 `--session-id my-task` same storage as `--session` |
| 14 | `explicit-session/tty-same-id-registry-meta` | `--session my-task` on grok-tty → same id on stderr/storage/registry/meta |

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./cmd/agent-run/tests/run/session-id-from-prompt                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./cmd/agent-run/tests/run/session-id-from-prompt
doctest test --label-all ./cmd/agent-run/tests/run/session-id-from-prompt

doctest vet ./cmd/agent-run/tests/run/session-id-from-prompt
doctest test ./cmd/agent-run/tests/run/session-id-from-prompt
doctest test -v ./cmd/agent-run/tests/run/session-id-from-prompt/auto-id/non-tty/storage-id-from-prompt-slug
doctest test -v ./cmd/agent-run/tests/run/session-id-from-prompt/auto-id/tty/same-id-storage-registry-meta
doctest test -v ./cmd/agent-run/tests/run/session-id-from-prompt/explicit-session/tty-same-id-registry-meta
doctest test -v ./cmd/agent-run/tests/run/session-id-from-prompt/flag-conflict/session-and-auto
```

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
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
	Runner     string // "fake-codex" | "grok-tty" | ""
	// GrokTTYCommand is AGENT_RUN_GROK_TTY_COMMAND when using grok-tty.
	GrokTTYCommand string
	KeepTTY        bool
	ExecTimeout    time.Duration
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
