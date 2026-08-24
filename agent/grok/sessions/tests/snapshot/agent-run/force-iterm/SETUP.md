## Scenario

`--iterm` skips agent-run even when a hit is available.

```go
import (
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureSnapshotSessionID
	writeProjectSnapshotSession(t, req)
	req.AgentRunByID = map[string]*sessions.AgentRunSnapshotResult{
		req.SessionID: {
			AgentRunSessionID: "ar-should-skip",
			Contents:          "should not appear",
		},
	}
	addLiveGrok(req, 4242, "/dev/ttys148")
	req.ITerm = oneITermTab()
	req.ContentsByID = map[string]iterm2.ContentsResult{
		"w2t1p0": {
			SessionID: "w2t1p0",
			App:       "/Applications/iTerm.app",
			Contents:  "forced iterm pane",
		},
	}
	req.Args = []string{req.SessionID, "--iterm", "--json"}
	return nil
}
```
