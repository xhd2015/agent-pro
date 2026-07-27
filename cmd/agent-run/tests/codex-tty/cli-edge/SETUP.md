# Scenario

**Subcommand**: `cli-edge` — codex-tty runner validation

```
agent-run run --agent-runner codex-tty → validateRunner accepts codex-tty (not unknown)
```

## Preconditions

- Fake TUI hook configured (inherited from root `SETUP.md`).
- Unknown runners exit 1 with stderr mentioning `unknown`; codex-tty must not.

## Steps

1. Leaf `Setup` sets `run --agent-runner codex-tty` with a minimal fake TUI prompt.
2. `Run` executes `agent-run`.
3. `Assert` verifies the run is not rejected as an unknown runner.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	return nil
}
```