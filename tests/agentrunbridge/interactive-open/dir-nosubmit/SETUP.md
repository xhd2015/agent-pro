# Scenario

**Feature**: InteractiveOpen forwards WorkspaceDir and NoSubmit into argv

```
RunInteractiveOpen(dir, noSubmit=true) -> … --dir=… --no-submit --open -- …
```

## Preconditions

- Non-empty WorkspaceDir; NoSubmit true.

## Steps

1. Set SessionID, Prompt, WorkspaceDir, NoSubmit.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "interactive_open"
	req.SessionID = "sess-io-dir"
	req.Prompt = "dir prompt"
	req.WorkspaceDir = "/work/bridge"
	req.NoSubmit = true
	req.AgentRunner = "grok-tty"
	return nil
}
```
