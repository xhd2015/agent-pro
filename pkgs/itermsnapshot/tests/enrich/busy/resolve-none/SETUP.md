# Scenario

**Feature**: busy pane + Kind=none is a soft miss (no agent)

```
busy + ResolveFromPID returns Kind=none -> Agents empty (success)
```

## Steps

1. One busy session with PID.
2. Resolve returns Kind none / empty SessionID (soft miss).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	const pid = 5300
	req.Snapshot = oneBusySession("sess-busy-none", "no-agent", "/dev/ttys013", pid)
	req.ResolveFromPID = func(p int) (*procresolve.Result, error) {
		return &procresolve.Result{
			InputPID: p,
			Kind:     "none",
			Tree: []procresolve.ProcNode{
				{PID: p, PPID: 1, Role: "input", Cmd: "python"},
			},
		}, nil
	}
	return nil
}
```
