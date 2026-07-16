# Scenario

**Feature**: `--detach` stays silent (no event stream / discovery think) while parent runs

```
agent-run run --agent-runner grok-tty --detach "hi"
  -> stdout/stderr have no "Resolve session id", 💭, 💬, NDJSON events
  -> allowed: optional soft-bind stderr lines (grok session / updates)
  -> allowed: session-id / terminal-id on stdout
```

## Steps

1. Use grouping detach args with prompt `hi`.
2. Assert combined output has no forbidden stream noise.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	return nil
}
```
