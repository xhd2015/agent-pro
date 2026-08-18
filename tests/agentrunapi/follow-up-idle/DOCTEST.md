# agentrunapi — FollowUpOpts idle-exit emit (`--exit-on-idle` / `--idle-timeout`)

Classic TDD doctests for **P1 library surface**: launch-time idle-exit fields on
`Opts` / `FollowUpOpts`, `NormalizeIdle`, and `BuildFollowUpCommand` emission.

Isolated nested root so parent `wait-driver` / `follow-up-color` / classify
trees stay **GREEN** until the implementer adds the fields + helper + emit
(this root is RED / compile-RED until then).

**Out of scope (P1):** watchdog / clock / `/exit` / kill serve, `idle-policy.json`,
`OnListening`, input-box occupancy, local-bot `ExecLauncher`, `agent-run resume`
flags, persist on `SessionMeta`, `agentrunbridge.BuildArgs`, real iTerm / grok / e2e TTY.

# DSN (Domain Specific Notion)

Launch-time idle-exit is a pair of follow-up flags. The helper decides whether
they are on; emit writes own tokens only when they are on.

**Participants**

- **Caller** builds a ForceNew follow-up line via `BuildFollowUpCommand` (open
  profile typical) and may also set the same fields on `Opts`.
- **`FollowUpOpts.ExitOnIdle` / `Opts.ExitOnIdle`** — bool; default off. When
  false, timeout is ignored (no error, no emit).
- **`FollowUpOpts.IdleTimeout` / `Opts.IdleTimeout`** — `time.Duration`. Zero
  with exit-on-idle on means default `10m` at normalize/emit time. Negative
  with exit-on-idle on is an API error.
- **`NormalizeIdle`** — pure: `(exitOnIdle, timeout) → (enabled, d, err)`.
  Shared by CLI parse and emit.
- **`DefaultIdleTimeout`** — `10 * time.Minute`. Compact emit form is `10m`
  (not `time.Duration.String()`'s `10m0s`).
- **`BuildFollowUpCommand`** — opts → one shell-quoted line. Emits
  `--exit-on-idle` and `--idle-timeout=<compact>` as own tokens, same family as
  `--color`, before `--` / prompt, only when normalize says enabled. Never
  `--new-terminal`.

**Behaviors**

- `ExitOnIdle=false` (timeout 0, `2s`, or even `-1s`) → enabled false; neither
  `--exit-on-idle` nor `--idle-timeout*` appears.
- `ExitOnIdle=true` and timeout `0` → enabled true, `d=10m`; tokens
  `--exit-on-idle` and `--idle-timeout=10m`.
- `ExitOnIdle=true` and timeout `2m` → `--idle-timeout=2m` (not `10m`).
- `ExitOnIdle=true` and timeout `2s` → `--idle-timeout=2s`.
- `ExitOnIdle=true` and timeout `-1s` → `NormalizeIdle` error; emit also errors
  when normalize sits on the emit path.

## Version

0.0.2

## Decision Tree

```
tests/agentrunapi/follow-up-idle/
├── DOCTEST.md
├── SETUP.md
├── disabled/                 # ExitOnIdle=false → omit both flags (timeout ignored)
│   ├── timeout-zero/         # timeout 0
│   ├── timeout-set/          # timeout 2s still omit
│   └── timeout-negative/     # timeout -1s: no error, still omit
└── enabled/                  # ExitOnIdle=true
    ├── default-timeout/      # 0 → 10m
    ├── timeout-2m/           # explicit 2m
    ├── timeout-2s/           # explicit 2s
    └── timeout-negative/     # -1s → API error
```

Parameter ranking (most → least significant):

1. **ExitOnIdle** — off (silent no-op) vs on (emit / validate)
2. **IdleTimeout** — zero/default | positive 2m | positive 2s | negative

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `disabled/timeout-zero` | `ExitOnIdle=false`, timeout 0 → omit both flags |
| 2 | `disabled/timeout-set` | `ExitOnIdle=false`, timeout `2s` → omit both flags |
| 3 | `disabled/timeout-negative` | `ExitOnIdle=false`, timeout `-1s` → no error; omit both flags |
| 4 | `enabled/default-timeout` | `ExitOnIdle=true`, timeout 0 → `--exit-on-idle --idle-timeout=10m` |
| 5 | `enabled/timeout-2m` | `ExitOnIdle=true`, timeout `2m` → `--idle-timeout=2m` (not `10m`) |
| 6 | `enabled/timeout-2s` | `ExitOnIdle=true`, timeout `2s` → `--idle-timeout=2s` |
| 7 | `enabled/timeout-negative` | `ExitOnIdle=true`, timeout `-1s` → normalize + emit API error |

## How to Run

```sh
# From the agent-pro module root:
doctest vet ./tests/agentrunapi/follow-up-idle
doctest test ./tests/agentrunapi/follow-up-idle

doctest test -v ./tests/agentrunapi/follow-up-idle/disabled
doctest test -v ./tests/agentrunapi/follow-up-idle/enabled
```

Expect **RED** (compile or assert) until `ExitOnIdle` / `IdleTimeout` on
`Opts` and `FollowUpOpts`, `DefaultIdleTimeout`, `NormalizeIdle`, and emit land.

Use `GOWORK=off` when a parent `go.work` would hide the agent-pro module.

### Planned API addition

```go
// package agentrunapi

const DefaultIdleTimeout = 10 * time.Minute

// On agentrunapi.Opts and agentrunapi.FollowUpOpts:
ExitOnIdle  bool
IdleTimeout time.Duration // 0 + ExitOnIdle → 10m at emit/normalize time

// NormalizeIdle is the shared pure helper (CLI + BuildFollowUpCommand).
// !exitOnIdle → enabled=false (timeout ignored, no error even if timeout set)
// exitOnIdle && timeout==0 → enabled=true, d=DefaultIdleTimeout
// exitOnIdle && timeout>0 → enabled=true, d=timeout
// exitOnIdle && timeout<0 → error
func NormalizeIdle(exitOnIdle bool, timeout time.Duration) (enabled bool, d time.Duration, err error)
```

Emit compact duration tokens (`10m`, `2m`, `2s`), not `time.Duration.String()`
(`10m0s`, `2m0s`). Token checks use `strings.Fields` + quote trim — never
substring of `--idle-timeout=10m`.

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/doctest/session"
)

// Request drives NormalizeIdle + BuildFollowUpCommand with idle-exit focus.
type Request struct {
	SessionID   string
	Prompt      string
	AgentRunner string
	Open        bool
	Detach      bool
	ExitOnIdle  bool
	IdleTimeout time.Duration
}

// Response is the harness observation (normalize + emit).
type Response struct {
	FollowUp     string
	ErrString    string
	Enabled      bool
	Normalized   time.Duration
	NormalizeErr string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	// Compile-prove both public structs carry the same launch-time fields.
	_ = agentrunapi.Opts{
		ExitOnIdle:  req.ExitOnIdle,
		IdleTimeout: req.IdleTimeout,
	}

	enabled, normalized, nerr := agentrunapi.NormalizeIdle(req.ExitOnIdle, req.IdleTimeout)
	resp := &Response{
		Enabled:    enabled,
		Normalized: normalized,
	}
	if nerr != nil {
		resp.NormalizeErr = nerr.Error()
	}

	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		SessionID:   req.SessionID,
		Prompt:      req.Prompt,
		AgentRunner: req.AgentRunner,
		Open:        req.Open,
		Detach:      req.Detach,
		ExitOnIdle:  req.ExitOnIdle,
		IdleTimeout: req.IdleTimeout,
	})
	resp.FollowUp = line
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func argvTokens(line string) []string {
	fields := strings.Fields(line)
	out := make([]string, 0, len(fields))
	for _, tok := range fields {
		out = append(out, strings.Trim(tok, `"'`))
	}
	return out
}

func tokensBeforeDashDash(line string) []string {
	toks := argvTokens(line)
	for i, tok := range toks {
		if tok == "--" {
			return toks[:i]
		}
	}
	return toks
}

func sliceHasToken(toks []string, want string) bool {
	for _, tok := range toks {
		if tok == want {
			return true
		}
	}
	return false
}

func hasExactToken(line, want string) bool {
	return sliceHasToken(argvTokens(line), want)
}

func hasIdleTimeoutPrefix(line string) bool {
	for _, tok := range argvTokens(line) {
		if tok == "--idle-timeout" || strings.HasPrefix(tok, "--idle-timeout=") {
			return true
		}
	}
	return false
}
```
