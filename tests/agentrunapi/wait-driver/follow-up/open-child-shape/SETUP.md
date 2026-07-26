# Scenario

**Feature**: open-profile child flags (session, runner, dir, nosubmit, open, prompt)

```
BuildFollowUpCommand(Open, SessionID, AgentRunner, WorkspaceDir, NoSubmit, Prompt)
  -> --session-id=… --agent-runner=… --auto-send-or-resume
     --dir=… --no-submit --open -- <prompt>
  -> no --new-terminal, no --detach
```

## Steps

1. Full open profile with dir and no-submit.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DriverBinary = ""
	req.SessionID = "sess-open-shape"
	req.Prompt = "open child"
	req.AgentRunner = "grok-tty"
	req.WorkspaceDir = "/tmp/ws-open"
	req.NoSubmit = true
	req.Open = true
	req.Detach = false
	return nil
}
```
