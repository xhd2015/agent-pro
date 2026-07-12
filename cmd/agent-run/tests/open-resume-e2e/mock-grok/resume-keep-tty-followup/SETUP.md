# Scenario

**Feature**: resume with followup without --open (keep-tty reopen + send turn)

```
open Paris → /exit → resume (no --open) "hello"
  -> not prompt required; not already in use; exit 0 preferred
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "resume-keep-tty-followup"
	req.SessionID = "e2e-resume-keeptty"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222209"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	req.ResumeNoOpen = true
	configureMockGrokEnv(t, req)
	return nil
}
```
