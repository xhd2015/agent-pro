# Scenario

**Feature**: `--open` on TTY runners: silent start, auto-attach, post-exit id, keep-alive

```
agent-run run --agent-runner grok-tty --open "prompt"
  -> silent while starting/attached
  -> auto-attach (AGENT_RUN_OPEN_ATTACH_INSTANT=1 in tests)
  -> stderr once "grok-tty: <id>" after attach returns
  -> registry kept for re-attach/send
```

## Preconditions

- Default runner for this branch: `grok-tty` with fake TUI.
- `OpenInstantAttach` enabled so CI has no interactive controlling TTY.
- Discovery bug out of scope — only silence of discovery progress is required.

## Steps

1. Grouping installs grok-tty fake TUI + instant attach + common `--open` args.
2. Leaves add prompt and specialize asserts (silence / id / registry / codex).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.OpenInstantAttach = true
	req.Prompt = "open-lifecycle"
	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
