# agentrunbridge — RunOpts.Color → BuildArgs `--color`

Classic TDD doctests for **P1 library surface**: when `RunOpts.Color` is true,
`BuildArgs` emits `--color` on the run argv (after open/detach flags, before
`-e` / `--` / prompt).

Isolated nested root so parent `tests/agentrunbridge` stays **GREEN** until the
implementer adds `Color bool` to `RunOpts` (this root is RED / compile-RED
until then).

# DSN (Domain Specific Notion)

**Participants**

- **Caller** maps structured `RunOpts` to agent-run CLI argv via `BuildArgs`.
- **`RunOpts.Color`** — boolean; true means emit `--color` so agent-run forces
  TTY child color env.
- **`BuildArgs`** — pure: opts → `[]string` starting with `run` (no binary name).

**Behaviors**

```
BuildArgs(RunOpts{Open, Color:true, …})
  -> … --open --color -- <prompt>
```

## Version

0.0.2

## Decision Tree

```
tests/agentrunbridge/color-flag/
├── DOCTEST.md
├── SETUP.md
└── emits-color/     # Color true → exact argv includes --color
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `emits-color` | Open-profile `Color` true → argv has `--color` before `--` |

## How to Run

```sh
doctest vet ./tests/agentrunbridge/color-flag
doctest test ./tests/agentrunbridge/color-flag
doctest test -v ./tests/agentrunbridge/color-flag/emits-color
```

Expect **RED** until `RunOpts.Color` + `BuildArgs` emission land.

### Planned API addition

```go
// On agentrunbridge.RunOpts:
Color bool // true → emit --color after open/detach, before -e / --
```

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunbridge"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	Prompt           string
	SessionID        string
	AgentRunner      string
	AutoSendOrResume bool
	NewTerminal      bool
	Open             bool
	Color            bool
}

type Response struct {
	Args      []string
	ErrString string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	args := agentrunbridge.BuildArgs(agentrunbridge.RunOpts{
		Prompt:           req.Prompt,
		SessionID:        req.SessionID,
		AgentRunner:      req.AgentRunner,
		AutoSendOrResume: req.AutoSendOrResume,
		NewTerminal:      req.NewTerminal,
		Open:             req.Open,
		Color:            req.Color,
	})
	return &Response{Args: args}, nil
}
```
