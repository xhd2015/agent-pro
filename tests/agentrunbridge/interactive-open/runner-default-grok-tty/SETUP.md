# Scenario

**Feature**: InteractiveOpen injects grok-tty when AgentRunner empty

```
RunInteractiveOpen(AgentRunner="") -> argv contains --agent-runner=grok-tty
# contrasts with BuildArgs/Run which omit the flag when empty
```

## Preconditions

- AgentRunner empty string.
- Explicit runner contrast leaf vs build-args/runner-empty-omitted.

## Steps

1. Leave AgentRunner empty; set session + prompt.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "interactive_open"
	req.SessionID = "sess-io-runner"
	req.Prompt = "runner default"
	req.AgentRunner = ""
	return nil
}
```
