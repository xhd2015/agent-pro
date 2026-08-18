# agent-run `run` — `--exit-on-idle` / `--idle-timeout` help and parse

Classic TDD doctests for **P1 CLI surface** only: `run -h` / `RunHelpText()`
documents the launch-time idle-exit flags, and parse-only validation rejects
bad `--idle-timeout` values. No keep-alive TTY. No product-binary e2e.

Isolated nested root so sibling `cmd/agent-run/tests` trees stay **GREEN**
while these flags / parse helper are missing (this root is RED / compile-RED
until then).

**Out of scope (P1):** watchdog / clock / `/exit`, `idle-policy.json`,
`OnListening`, local-bot always-on, `agent-run resume` flags, persist on
`SessionMeta`, real iTerm / grok / e2e TTY. Silent no-op *emit* lives in
`tests/agentrunapi/follow-up-idle/` (this tree only proves parse-OK).

# DSN (Domain Specific Notion)

`agent-run run` accepts optional idle-exit flags. Help names them. Parse
fails fast on a bad duration and never opens a TTY.

**Participants**

- **`agent-run run` CLI** — optional `--exit-on-idle` (default off) and
  `--idle-timeout DUR` (default `10m` when idle-exit is on). Documents both
  on `run -h`.
- **`RunHelpText()`** — same `run -h` body Handle prints via flags.Help.
  L2; no binary; no process stdio.
- **`ParseRunIdle`** — CLI-facing parse: raw `--idle-timeout` string plus
  `--exit-on-idle` bool. Invalid string errors before normalize. Negative
  with exit-on-idle on errors. Timeout without exit-on-idle is parse-OK
  (enabled=false). Does not start a TTY.
- **User-facing error** — stderr `Error:` prefix, non-zero. Preferred:
  `Error: --idle-timeout must be a positive duration (got -1s)` and
  `Error: invalid value for --idle-timeout: nope`.

**Behaviors**

- `run -h` / `RunHelpText()` documents `--exit-on-idle`, `--idle-timeout`,
  and default `10m`, including the keep-alive / sendable-prompt wording.
- `--exit-on-idle --idle-timeout=-1s` → parse error, `Error:`, non-zero.
- `--idle-timeout=nope` → parse error, `Error:`, non-zero.
- `--idle-timeout=2s` without `--exit-on-idle` → parse OK, enabled=false,
  no TTY.

## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/run-exit-on-idle/
├── DOCTEST.md
├── SETUP.md
├── help/                                      # L2 RunHelpText
│   └── run-help-lists-flags/                  # documents both flags + 10m
└── parse/                                     # L2 ParseRunIdle (no TTY)
    ├── invalid/
    │   ├── negative-with-exit-on-idle/        # -1s + flag → Error:
    │   └── unparseable/                       # nope → Error:
    └── valid/
        └── timeout-without-exit-on-idle/      # 2s, flag off → parse OK
```

Parameter ranking (most → least significant):

1. **Surface** — help vs parse
2. **Parse outcome** — invalid vs valid
3. **Invalid kind** — negative duration vs unparseable string

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `help/run-help-lists-flags` | `RunHelpText()` documents `--exit-on-idle`, `--idle-timeout`, default `10m` |
| 2 | `parse/invalid/negative-with-exit-on-idle` | `--exit-on-idle` + `-1s` → stderr `Error:`, non-zero |
| 3 | `parse/invalid/unparseable` | `--idle-timeout=nope` → stderr `Error:`, non-zero |
| 4 | `parse/valid/timeout-without-exit-on-idle` | `2s` without `--exit-on-idle` → parse OK, enabled=false |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./cmd/agent-run/tests/run-exit-on-idle
doctest test ./cmd/agent-run/tests/run-exit-on-idle

doctest test -v ./cmd/agent-run/tests/run-exit-on-idle/help
doctest test -v ./cmd/agent-run/tests/run-exit-on-idle/parse
```

Expect **RED** (compile or assert) until `RunHelpText` documents the flags
and `ParseRunIdle` lands. No `label: e2e`.

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

### Planned API addition

```go
// package agentruncli

// ParseRunIdle interprets launch-time idle flags after the CLI reads
// --exit-on-idle and the raw --idle-timeout string (empty = omitted).
// Invalid raw → error before NormalizeIdle.
// !exitOnIdle → enabled=false, no error even if timeout parses (including 2s).
// exitOnIdle && timeout==0 / omitted → enabled=true, d=10m
// exitOnIdle && timeout<0 → error
// Does not start a TTY.
func ParseRunIdle(exitOnIdle bool, timeoutRaw string) (enabled bool, d time.Duration, err error)
```

Preferred error strings (no `Error:` prefix on the returned error; harness /
CLI presentation adds it):

```
--idle-timeout must be a positive duration (got -1s)
invalid value for --idle-timeout: nope
```

Help text that must appear on `run -h` / `RunHelpText()`:

```
--exit-on-idle      exit the keep-alive TTY after idle-timeout at a sendable prompt
                    (no-op unless --open, --detach, or --keep-tty actually keep a TTY)
--idle-timeout DUR  idle window used with --exit-on-idle (default: 10m)
```

```go
import (
	"fmt"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
	"github.com/xhd2015/doctest/session"
)

const (
	opHelp  = "help"
	opParse = "parse"
)

// Request drives one L2 CLI surface: help or parse-only idle flags.
type Request struct {
	Op             string // help | parse
	Args           []string
	ExitOnIdle     bool
	IdleTimeoutRaw string
}

// Response holds observed help text or parse presentation.
type Response struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Enabled   bool
	Timeout   time.Duration
	ErrString string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	switch req.Op {
	case opHelp:
		return runHelp(t, req)
	case opParse:
		return runParse(t, req)
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}

func runHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	_ = req
	text := agentruncli.RunHelpText()
	return &Response{
		Stdout:   text,
		ExitCode: 0,
	}, nil
}

func runParse(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	enabled, d, err := agentruncli.ParseRunIdle(req.ExitOnIdle, req.IdleTimeoutRaw)
	resp := &Response{
		Enabled: enabled,
		Timeout: d,
	}
	if err != nil {
		resp.ErrString = err.Error()
		resp.Stderr = fmt.Sprintf("Error: %s\n", err.Error())
		resp.ExitCode = 1
		return resp, nil
	}
	resp.ExitCode = 0
	return resp, nil
}
```
