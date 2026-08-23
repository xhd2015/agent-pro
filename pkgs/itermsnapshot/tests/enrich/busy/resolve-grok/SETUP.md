# Scenario

**Feature**: busy pane + grok hard hit attaches SessionAgent

```
busy PID=5100 + ResolveFromPID -> Kind=grok SessionID Title Tree
  -> Agents["sess-busy-grok"] set
```

## Steps

1. One busy session `sess-busy-grok` with PID 5100.
2. Resolve returns grok hard hit with Title and Tree.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	const sessID = "sess-busy-grok"
	const pid = 5100
	req.Snapshot = oneBusySession(sessID, "grok-work", "/dev/ttys011", pid)
	req.ResolveFromPID = resolveHit("grok", "abc-grok-session", "My Grok Title", []procresolve.ProcNode{
		{PID: pid, PPID: 1, Role: "input", Cmd: "zsh"},
		{PID: pid + 1, PPID: pid, Role: "grok", Cmd: "grok"},
	})
	return nil
}
```
