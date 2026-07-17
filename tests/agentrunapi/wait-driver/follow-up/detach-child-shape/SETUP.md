# Scenario

**Feature**: detach-profile child flags (no open, no new-terminal)

```
BuildFollowUpCommand(Detach, SessionID, AgentRunner, Prompt, AllowRelocate)
  -> --auto-send-or-resume --detach -- <prompt>
  -> no --open, no --new-terminal
```

## Steps

1. Detach profile with allow-relocate.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.DriverBinary = ""
	req.SessionID = "sess-detach-shape"
	req.Prompt = "detach child"
	req.AgentRunner = "grok-tty"
	req.WorkspaceDir = "/tmp/ws-detach"
	req.AllowRelocateResumeSessionDir = true
	req.Open = false
	req.Detach = true
	return nil
}
```
