# Scenario

**Feature**: resume --open --no-submit "hello" injects without requiring missing --open error

```
open Paris → /exit → exited → resume --open --no-submit "hello"
  -> exit 0; not already in use; not "no-submit requires --open"
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "resume-open-no-submit-hello"
	req.SessionID = "e2e-resume-nosubmit"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222202"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	req.NoSubmit = true
	req.ResumeOpen = true
	configureMockGrokEnv(t, req)
	return nil
}
```
