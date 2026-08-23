# Scenario

**Feature**: when PID is nil, resolve uses ShellPID (kool parity)

```
busy Idle=false, PID=nil, ShellPID=6100
  -> ResolveFromPID(6100) -> Agents["sess-shell-fallback"]
```

## Steps

1. Busy session with PID nil and ShellPID 6100.
2. Resolve hit on that pid; record call PIDs for assert.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	const sessID = "sess-shell-fallback"
	const shellPID = 6100
	req.Snapshot = oneSessionSnap(
		sessID,
		"shell-only-busy",
		"/dev/ttys015",
		boolPtr(false),
		nil, // PID absent
		intPtr(shellPID),
	)
	calls := make([]int, 0, 1)
	req.ResolveCallPIDs = &calls
	hit := resolveHit("grok", "grok-via-shell", "shell-title", []procresolve.ProcNode{
		{PID: shellPID, PPID: 1, Role: "input", Cmd: "zsh"},
	})
	req.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		*req.ResolveCallPIDs = append(*req.ResolveCallPIDs, pid)
		return hit(pid)
	}
	return nil
}
```
