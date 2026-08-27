# Scenario

Pane inventory has empty cwd; TITLE and WORKSPACE come from summary.json
beside the open-file hard-hit session dir (`DiskCwd`).

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

const fixtureListLiveDiskTitle = "from-disk-title"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	sessionDir := writeListLiveSession(t, req.GrokHome, fixtureListLiveSID, fixtureListLiveDiskCWD, fixtureListLiveDiskTitle)
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	pointOpenFileAtSession(req, 5001, sessionDir)
	// Empty pane cwd — force summary.json meta fallback.
	req.PaneByTTY["/dev/ttys148"] = sessions.LivePaneInfo{}
	req.DiskCwd = true
	req.CwdBySession = nil
	req.Args = nil
	return nil
}
```
