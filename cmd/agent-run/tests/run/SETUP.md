# Scenario

**Subcommand**: `run` — headless one-shot agent invocation

## Preconditions

- `agent-run` and `fake-codex` binaries are built (inherited from root `SETUP.md`).
- `AGENT_RUN_HOME` is isolated per test.
- Grouping `Setup` prefixes `req.Args` with `run` and `--agent-runner fake-codex`.

## Steps

1. Grouping `Setup` sets common `run` args prefix.
2. Leaf `Setup` adds `--json` or prompt-specific flags.
3. `Run` executes `agent-run` and captures stdout/stderr.
4. `Assert` checks NDJSON stream, persistence, or human-readable output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "fake-codex"}
	return nil
}
```