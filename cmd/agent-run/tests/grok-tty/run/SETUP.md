# Scenario

**Subcommand**: `run --agent-runner grok-tty` — adhoc server, registry, capture sidecar

```
agent-run run --agent-runner grok-tty "prompt"
  -> stderr grok-tty: session-N
  -> registry entry while running
  -> wait GROK_TTY_BANNER → inject prompt
  -> discover GROK_HOME session → tail updates.jsonl → stream stdout / events.jsonl
  -> (fallback) capture scrollback → stdout / events.jsonl
```

## Preconditions

- Fake TUI via `AGENT_RUN_GROK_TTY_COMMAND` (inherited unless leaf overrides).
- Streaming leaves seed temp `GROK_HOME` + synthetic ACP `updates.jsonl` via root helpers;
  optional `AGENT_RUN_GROK_TTY_GROK_SESSION_ID` skips discovery.
- Grouping `Setup` prefixes `req.Args` with `run --agent-runner grok-tty`.

## Steps

1. Grouping `Setup` sets common run args prefix.
2. Leaf `Setup` chooses fake TUI variant, prompt, or background mode.
3. `Run` executes blocking `agent-run run` or registry probe while background run lives.
4. `Assert` checks stderr id, registry, banner wait, or captured output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	return nil
}
```