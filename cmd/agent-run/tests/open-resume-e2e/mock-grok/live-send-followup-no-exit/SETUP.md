# Scenario

**Feature**: while session still live after open, send followup without /exit

```
open Paris → send "hello" (no /exit)
  -> HELLO_RESUME_MARKER or hello in snapshot/events
  -> still not required to be exited
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "live-send-followup-no-exit"
	req.SessionID = "e2e-live-send"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222203"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	configureMockGrokEnv(t, req)
	return nil
}
```
