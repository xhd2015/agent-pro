# Scenario

**Feature**: RunInteractiveDetach forwards WorkspaceDir + allow-relocate

```
RunInteractiveDetach(... dir + AllowRelocateResumeSessionDir)
  -> argv has --dir=… --allow-relocate-resume-session-dir --detach
```

## Steps

1. Mode interactive_detach; set dir + allow-relocate; ready status via parent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "interactive_detach"
	req.SessionID = "sess-id-dir"
	req.Prompt = "detach dir"
	req.WorkspaceDir = "/tmp/ws-id"
	req.AllowRelocateResumeSessionDir = true
	req.AgentRunner = "grok-tty"
	return nil
}
```
