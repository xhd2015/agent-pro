# Scenario

**Feature**: `--open --no-submit` on a TTY runner injects without auto-submit

```
agent-run run --agent-runner grok-tty --open --no-submit "draft"
  -> silent while starting/attached
  -> inject prompt with suffixCR=false (no trailing \r)
  -> auto-attach (AGENT_RUN_OPEN_ATTACH_INSTANT=1 in tests)
  -> stderr once "grok-tty: <id>" after attach returns
  -> CR-sensitive fake TUI does not print SUBMITTED:draft
```

## Preconditions

- Default runner for this branch: `grok-tty` with CR-sensitive fake TUI.
- `OpenInstantAttach` enabled so CI has no interactive controlling TTY.
- Non-empty prompt is required for the inject-without-CR proof leaf.

## Steps

1. Grouping installs grok-tty defaults + instant attach + common open+no-submit args.
2. Leaf specializes prompt, CR-sensitive TUI, snapshot mode, and asserts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.OpenInstantAttach = true
	req.Prompt = "draft"
	setGrokTTYCommand(req, fakeTUIRequiresCR())
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"--open",
		"--no-submit",
		req.Prompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
