# Scenario

**Feature**: RunInteractiveDetach defaults empty AgentRunner to grok-tty

```
RunInteractiveDetach(AgentRunner="") -> --agent-runner=grok-tty in launch argv
```

## Steps

1. Explicit empty runner; unique session for clarity.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "interactive_detach"
	req.SessionID = "sess-id-runner"
	req.Prompt = "default runner"
	req.AgentRunner = ""
	return nil
}
```
