# agentrunapi — FollowUpOpts.Color → BuildFollowUpCommand `--color`

Classic TDD doctests for **P1 library surface**: when `FollowUpOpts.Color` is
true, `BuildFollowUpCommand` emits `--color` in the run argv (before `--` /
prompt). When false, no `--color` token.

Isolated nested root so parent `wait-driver` / classify trees stay **GREEN**
until the implementer adds `Color bool` to `FollowUpOpts` (this root is RED /
compile-RED until then).

# DSN (Domain Specific Notion)

**Participants**

- **Caller** builds a ForceNew follow-up line via `BuildFollowUpCommand`.
- **`FollowUpOpts.Color`** — boolean; true means the child `agent-run run`
  should force TTY color env (`--color`).
- **`BuildFollowUpCommand`** — pure: opts → single shell-quoted line; must
  include `--color` as its own token when Color is true; must omit it when false.

**Behaviors**

```
BuildFollowUpCommand(Color:true,  Open, SessionID, Prompt)
  -> line contains token --color

BuildFollowUpCommand(Color:false, Open, SessionID, Prompt)
  -> line has no --color token
```

## Version

0.0.2

## Decision Tree

```
tests/agentrunapi/follow-up-color/
├── DOCTEST.md
├── SETUP.md
├── color-true/     # L1: Color true → --color present
└── color-false/    # L2: Color false → --color absent
```

Parameter ranking: **Color true vs false** (only factor).

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `color-true` | `FollowUpOpts.Color` true → shell line contains `--color` |
| 2 | `color-false` | `FollowUpOpts.Color` false → no `--color` token |

## How to Run

```sh
doctest vet ./tests/agentrunapi/follow-up-color
doctest test ./tests/agentrunapi/follow-up-color

doctest test -v ./tests/agentrunapi/follow-up-color/color-true
doctest test -v ./tests/agentrunapi/follow-up-color/color-false
```

Expect **RED** (compile or assert) until `FollowUpOpts.Color` and emission land.

### Planned API addition

```go
// On agentrunapi.FollowUpOpts:
Color bool // true → emit --color on run argv before -- / prompt
```

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/doctest/session"
)

// Request drives BuildFollowUpCommand with Color focus.
type Request struct {
	SessionID   string
	Prompt      string
	AgentRunner string
	Open        bool
	Detach      bool
	Color       bool
}

// Response is the harness observation.
type Response struct {
	FollowUp  string
	ErrString string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	line, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		SessionID:   req.SessionID,
		Prompt:      req.Prompt,
		AgentRunner: req.AgentRunner,
		Open:        req.Open,
		Detach:      req.Detach,
		Color:       req.Color,
	})
	resp := &Response{FollowUp: line}
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func hasColorToken(line string) bool {
	for _, tok := range strings.Fields(line) {
		raw := strings.Trim(tok, `'"'`)
		if raw == "--color" {
			return true
		}
	}
	return false
}
```
