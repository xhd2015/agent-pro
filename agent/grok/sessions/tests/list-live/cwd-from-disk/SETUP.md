# Scenario

Pane inventory has empty cwd; TITLE and WORKSPACE come from GrokHome summary
via the selective disk meta index (`DiskCwd`).

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

const fixtureListLiveDiskTitle = "from-disk-title"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeListLiveSession(t, req.GrokHome, fixtureListLiveSID, fixtureListLiveDiskCWD, fixtureListLiveDiskTitle)
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	// Empty pane cwd — force disk meta fallback.
	req.PaneByTTY["/dev/ttys148"] = sessions.LivePaneInfo{}
	req.DiskCwd = true
	req.CwdBySession = nil
	req.Args = nil
	return nil
}
```
