# Scenario

**Feature**: non-open headless still hard-waits inject-ready banner before inject

```
agent-run run --agent-runner grok-tty "hi"
  + delayed GROK_TTY_BANNER
  -> wait banner → inject → exit 0
```

## Preconditions

- No `--open`; no INSTANT attach env (irrelevant for non-open attach).
- Fake TUI delays banner then reads prompt (existing waits-for-banner style).

## Steps

1. Grouping sets grok-tty non-open base args.
2. Leaf installs delayed-banner fake TUI + prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	clearOpenInstantAttach(req)
	req.Runner = "grok-tty"
	req.Args = []string{"run", "--agent-runner", "grok-tty"}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
