## Scenario

Ambiguous grok→agent-run mapping warns and falls back to iTerm.

```go
import (
	"fmt"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureSnapshotSessionID
	writeProjectSnapshotSession(t, req)
	req.AgentRunErr = fmt.Errorf("ambiguous grok-session-id %s: multiple matches: a, b", req.SessionID)
	addLiveGrok(req, 4242, "/dev/ttys148")
	req.ITerm = oneITermTab()
	req.ContentsByID = map[string]iterm2.ContentsResult{
		"w2t1p0": {
			SessionID: "w2t1p0",
			App:       "/Applications/iTerm.app",
			Contents:  "after ambiguous warn",
		},
	}
	req.Args = []string{req.SessionID}
	return nil
}
```
