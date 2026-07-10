# Scenario

**Feature**: after auto-attach exits, print `grok-tty: <id>` once on stderr

```
agent-run run --agent-runner grok-tty --open "hi"
  -> (instant attach returns)
  -> stderr contains exactly one "grok-tty: <session-id>" id line
  -> stdout lacks "grok-tty:"
```

## Steps

1. Run open with respond-hi fake TUI and instant attach.
2. Assert single prefixed session id line on stderr.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	setGrokTTYCommand(req, fakeTUIRespondHi())
	return nil
}
```
