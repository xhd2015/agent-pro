# Scenario

**Feature**: busy pane + codex hard hit attaches SessionAgent

```
busy PID=5200 + ResolveFromPID Kind=codex -> Agents["sess-busy-codex"]
```

## Steps

1. One busy session `sess-busy-codex` with PID 5200.
2. Resolve returns codex hard hit (Title may be empty).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	const sessID = "sess-busy-codex"
	const pid = 5200
	req.Snapshot = oneBusySession(sessID, "codex-work", "/dev/ttys012", pid)
	req.ResolveFromPID = resolveHit("codex", "codex-sess-xyz", "", []procresolve.ProcNode{
		{PID: pid, PPID: 1, Role: "input", Cmd: "bash"},
		{PID: pid + 1, PPID: pid, Role: "codex", Cmd: "codex"},
	})
	return nil
}
```
