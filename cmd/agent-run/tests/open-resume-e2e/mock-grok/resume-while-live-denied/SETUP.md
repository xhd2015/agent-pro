# Scenario

**Feature**: resume while agent still live (no /exit) is denied

```
open Paris → resume --open "hello" (without /exit)
  -> exit ≠ 0; not "already in use"; steer to still active / use send
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "resume-while-live-denied"
	req.SessionID = "e2e-resume-live-deny"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222205"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	req.ExpectResumeDenied = true
	req.ResumeOpen = true
	configureMockGrokEnv(t, req)
	return nil
}
```
