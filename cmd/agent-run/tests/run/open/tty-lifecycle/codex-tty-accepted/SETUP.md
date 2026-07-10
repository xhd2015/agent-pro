# Scenario

**Feature**: `--open` is accepted for `codex-tty` (TTY runner, not rejected as non-TTY)

```
agent-run run --agent-runner codex-tty --open "hi"
  -> not rejected as unknown / non-TTY
  -> exit 0 with instant attach
```

## Steps

1. Override grouping runner to `codex-tty`.
2. Set `AGENT_RUN_CODEX_TTY_COMMAND` fake TUI + instant attach.
3. Run `--open` with a short prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "codex-tty"
	req.OpenInstantAttach = true
	req.Prompt = "hi"
	setCodexTTYCommand(req, fakeTUIRespondHi())
	// Clear grok-specific env if grouping set it.
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Args = []string{"run", "--agent-runner", "codex-tty", "--open", req.Prompt}
	return nil
}
```
