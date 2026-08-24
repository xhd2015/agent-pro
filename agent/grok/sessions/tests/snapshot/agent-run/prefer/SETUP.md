## Scenario

Managed Grok id hits agent-run snapshot; Contents must not run.

```go
import (
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureSnapshotSessionID
	writeProjectSnapshotSession(t, req)
	req.AgentRunByID = map[string]*sessions.AgentRunSnapshotResult{
		req.SessionID: {
			AgentRunSessionID: "ar-fixture-session",
			Contents:          "agent-run single frame\n│ ❯ │",
		},
	}
	// iTerm fixtures present but must be unused when prefer hits.
	addLiveGrok(req, 4242, "/dev/ttys148")
	req.ITerm = oneITermTab()
	req.Args = []string{req.SessionID, "--json"}
	return nil
}
```
