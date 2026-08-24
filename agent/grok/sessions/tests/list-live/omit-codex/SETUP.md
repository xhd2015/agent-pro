# Scenario

```go
import (
	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Procs = []sessions.FocusProc{
		{PID: 6001, PPID: 1, TTY: "ttys200", Cmd: "/usr/local/bin/codex"},
	}
	req.OpenFiles[6001] = []string{
		"/Users/fixture/.codex/sessions/2026/08/01/rollout-2026-08-01T12-00-00-" + fixtureListLiveSID + ".jsonl",
	}
	req.ITerm = []iterm2.SessionRef{
		{WindowID: "1", TabIndex: 1, SessionID: "iterm-codex", TTY: "/dev/ttys200"},
	}
	req.Args = []string{}
	return nil
}
```
