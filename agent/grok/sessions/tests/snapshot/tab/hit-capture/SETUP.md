# Scenario

```go
import (
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeSnapshotSession(t, req.GrokHome, fixtureTabSnapshotSessionID, req.ProjectDir, "tab snapshot")
	seedSnapshotTabWindow(req)
	req.ContentsByID = map[string]iterm2.ContentsResult{
		"TAB2-UUID": {SessionID: "TAB2-UUID", App: "/Applications/iTerm.app", Contents: "tab2 pane"},
	}
	req.Args = []string{"--tab", "2"}
	return nil
}
```
