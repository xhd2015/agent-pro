# Scenario

**Feature**: busy pane + ResolveFromPID error is soft (no agent, no Capture fail)

```
busy + ResolveFromPID error -> Capture success, Agents empty
```

## Steps

1. One busy session with PID.
2. Resolve returns an error.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Snapshot = oneBusySession("sess-busy-err", "resolve-fails", "/dev/ttys014", 5400)
	req.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		return nil, fmt.Errorf("simulated resolve failure for pid %d", pid)
	}
	return nil
}
```
