# Scenario

**Subcommand**: `run --agent-runner codex-tty` — adhoc server, registry, scrollback capture

```
agent-run run --agent-runner codex-tty "prompt"
  -> stderr codex-tty: session-N
  -> registry entry while running
  -> wait CODEX_TTY_BANNER → inject prompt
  -> capture scrollback → stdout / events.jsonl
```

## Preconditions

- Fake TUI via `AGENT_RUN_CODEX_TTY_COMMAND` (inherited unless leaf overrides).
- Grouping `Setup` prefixes `req.Args` with `run --agent-runner codex-tty`.

## Steps

1. Grouping `Setup` sets common run args prefix.
2. Leaf `Setup` chooses fake TUI variant, prompt, or background mode.
3. `Run` executes blocking `agent-run run` or registry probe while background run lives.
4. `Assert` checks stderr id, registry, banner wait, or captured output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "codex-tty"}
	return nil
}
```
