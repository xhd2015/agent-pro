# agent-run run --auto-session-id Tests

Doc-style tests for `agent-run run --auto-session-id` and the related same-id
policy for explicit `--session` on TTY runners.

`--auto-session-id` generates a human-readable session id from the prompt slug
(`<slug>-YYYYMMDD-HHMMSS`, then `-1`, `-2`, … on collision) and uses that **same
string** for agent storage and, on TTY runners, the terminal registry id.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI** — `run` parses `--auto-session-id` / `--session`, resolves the
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
agent-run run --auto-session-id "prompt"
  -> slugify(prompt) + local timestamp [+ -N]
  -> sessions/<runner>/<id>/
  -> [TTY] registry <id>.json + stderr "<runner>: <id>"

# explicit: user-chosen id also used as TTY registry id
agent-run run --session my-task "prompt"
  -> sessions/<runner>/my-task/
  -> [TTY] registry my-task.json + stderr "<runner>: my-task"

# mutual exclusion
agent-run run --session X --auto-session-id "p" -> error, exit ≠ 0

# help
agent-run run --help -> documents --auto-session-id
```

Conflict (id taken) when storage already has `sessions/<runner>/<id>/` (meta
exists), or for TTY a live registry entry exists for that id.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run/auto-session-id/
├── DOCTEST.md
├── SETUP.md
├── help/
│   └── run-help-lists-flag/              # run --help documents --auto-session-id
├── flag-conflict/
│   └── session-and-auto/                 # --session + --auto-session-id → error
├── auto-id/                              # --auto-session-id set; --session unset
│   ├── non-tty/                          # fake-codex (storage only)
│   │   ├── storage-id-from-prompt-slug/  # storage dir + id shape from prompt
│   │   ├── punctuation-slug/             # "Hello, World!!" → hello-world-…
│   │   ├── long-prompt-truncates-base/   # base ≤ 128 runes before timestamp
│   │   ├── empty-slug-uses-fallback-base/# "!!!" → sess|task + timestamp
│   │   └── storage-collision-suffix/     # pre-seed storage → -N suffix
│   └── tty/                              # grok-tty (storage + registry same id)
│       ├── same-id-storage-registry-meta/# stderr, storage, meta, registry agree
│       └── storage-collision-suffix/     # TTY auto-id also gets -N on storage clash
└── explicit-session/                     # --session set; --auto-session-id unset
    ├── non-tty-storage-id/               # storage uses explicit id
    └── tty-same-id-registry-meta/        # TTY registry/stderr/meta use same explicit id
```

Parameter ranking (most → least significant):

1. **Session id policy** — help | flag conflict | auto-id | explicit `--session`
2. **Runner class** — non-TTY (`fake-codex`) vs TTY (`grok-tty`)
3. **Slug / collision edge** — base derive, punctuation, length, empty slug, collision

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-flag` | `run --help` lists `--auto-session-id`; stdout ends with `\n` |
| 2 | `flag-conflict/session-and-auto` | Both flags → exit ≠ 0; clear stderr error |
| 3 | `auto-id/non-tty/storage-id-from-prompt-slug` | Storage under `sessions/fake-codex/<slug-ts…>/`; id matches shape |
| 4 | `auto-id/non-tty/punctuation-slug` | `"Hello, World!!"` → base `hello-world` + timestamp |
| 5 | `auto-id/non-tty/long-prompt-truncates-base` | Base portion ≤ 128 runes before `-YYYYMMDD-HHMMSS` |
| 6 | `auto-id/non-tty/empty-slug-uses-fallback-base` | Punctuation-only prompt → base `sess` or `task` |
| 7 | `auto-id/non-tty/storage-collision-suffix` | Pre-seed storage for nearby timestamps → id ends with `-N` |
| 8 | `auto-id/tty/same-id-storage-registry-meta` | stderr `grok-tty: <id>` == storage dir == meta == registry |
| 9 | `auto-id/tty/storage-collision-suffix` | TTY auto-id collision appends `-N`; storage and registry share it |
| 10 | `explicit-session/non-tty-storage-id` | `--session my-task` → `sessions/fake-codex/my-task/` |
| 11 | `explicit-session/tty-same-id-registry-meta` | `--session my-task` on grok-tty → same id on stderr/storage/registry/meta |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/run/auto-session-id
doctest test ./cmd/agent-run/tests/run/auto-session-id
doctest test -v ./cmd/agent-run/tests/run/auto-session-id/auto-id/non-tty/storage-id-from-prompt-slug
doctest test -v ./cmd/agent-run/tests/run/auto-session-id/auto-id/tty/same-id-storage-registry-meta
doctest test -v ./cmd/agent-run/tests/run/auto-session-id/explicit-session/tty-same-id-registry-meta
doctest test -v ./cmd/agent-run/tests/run/auto-session-id/flag-conflict/session-and-auto
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

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
