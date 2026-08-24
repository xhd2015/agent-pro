# Scenario

```go
import "github.com/xhd2015/agent-pro/agent/grok/sessions"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Live grok with open files, but no matching iTerm tab.
	req.Procs = []sessions.FocusProc{
		{PID: 5001, PPID: 1, TTY: "ttys148", Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles[5001] = []string{grokListLivePath(fixtureListLiveSID)}
	req.ITerm = nil
	req.Args = []string{}
	return nil
}
```
