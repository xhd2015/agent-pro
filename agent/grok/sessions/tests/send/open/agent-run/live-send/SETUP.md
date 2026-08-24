# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectSendSession(t, req)
	// Stale/live iTerm host present — must NOT be used when agent-run manages the id.
	addLiveGrok(req, 4242, "/dev/ttys148")
	req.ITerm = oneITermTab()
	req.AgentRunByID = map[string]*sessions.AgentRunOpenResult{
		req.SessionID: {
			AgentRunSessionID: "ar-live-1",
			Mode:              sessions.AgentRunOpenModeSend,
			Delivered:         true,
		},
	}
	// Prefer works without --open for managed --session-id.
	req.Args = []string{"hello", "--session-id", req.SessionID}
	return nil
}
```
