# Scenario

**Feature**: two consecutive live sends after open

```
open Paris → send "hello" → send "hello again"
  -> both send exit 0; hello marker present
```

## Steps

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "send-two-followups-while-live"
	req.SessionID = "e2e-two-sends"
	req.GrokSessionUUID = "b2222222-2222-4222-8222-222222222207"
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = "hello"
	req.HelloMarker = defaultHelloMarker
	configureMockGrokEnv(t, req)
	return nil
}
```
