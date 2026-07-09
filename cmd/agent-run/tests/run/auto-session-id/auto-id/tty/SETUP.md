# Scenario

**Feature**: auto-session-id with TTY runner uses the same id for storage and registry

```
agent-run run --agent-runner grok-tty --auto-session-id "prompt"
  -> sessions/grok-tty/<id>/
  -> stderr grok-tty: <id>
  -> grok-tty-registry/<id>.json (with --keep-tty)
  -> meta.terminal_session_id == <id>
```

## Preconditions

- Fake TUI via `AGENT_RUN_GROK_TTY_COMMAND`.
- `--keep-tty` used so registry persists after run for file asserts.

## Steps

1. Prefix `--agent-runner grok-tty --keep-tty`.
2. Install respond-hi fake TUI.
3. Leaves add prompts / collision seeds.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.KeepTTY = true
	req.Args = append(req.Args, "--agent-runner", "grok-tty", "--keep-tty")
	setGrokTTYCommand(req, fakeTUIRespondHi())
	return nil
}
```
