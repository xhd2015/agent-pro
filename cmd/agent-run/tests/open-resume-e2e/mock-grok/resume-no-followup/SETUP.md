# Scenario

**Feature**: after /exit, resume without followup reopens (no prompt required)

```
open Paris → /exit → exited → resume --open (no followup)
  -> not "prompt is required"; exit 0 preferred; not already in use
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "resume-no-followup"
	req.SessionID = "e2e-resume-nofollow"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222201"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "" // reopen only
	req.HelloMarker = defaultHelloMarker
	req.ResumeOpen = true
	configureMockGrokEnv(t, req)
	return nil
}
```
