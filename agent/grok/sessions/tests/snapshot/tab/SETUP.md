# Scenario

**Feature**: snapshot via --tab / --tab-index (ResolveFromTab → Contents)

```go
import (
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	fixtureTabSnapshotSessionID = "019f283b-dddd-7ddd-dddd-dddddddddddd"
	pidTabSnapshotGrok          = 8200
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.SessionID = fixtureTabSnapshotSessionID
	return nil
}

func seedSnapshotTabWindow(req *Request) {
	req.CurrentSessionID = "w0t1p0:CURRENT-TAB-UUID"
	req.ControllingTTY = "/dev/ttys101"
	req.ITerm = []iterm2.SessionRef{
		{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101", Name: "current"},
		{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102", Name: "grok-tab"},
		{WindowID: "100", WindowName: "work", TabIndex: 3, SessionID: "w0t3p0:TAB3-UUID", TTY: "/dev/ttys103", Name: "bash-only"},
	}
	req.Procs = []sessions.FocusProc{
		{PID: pidTabSnapshotGrok, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/grok"},
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	req.OpenFiles[pidTabSnapshotGrok] = []string{grokSnapshotPath(fixtureTabSnapshotSessionID)}
}
```
