# Scenario

**Feature**: idle pane never gets an agent

```
Idle=true session + resolve that would hit -> Capture -> Agents empty
```

## Steps

1. Inject one idle session with PID.
2. Resolve inject would hard-hit if called; attach must skip idle first.

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
	req.Snapshot = oneSessionSnap(
		"sess-idle-1",
		"idle-shell",
		"/dev/ttys001",
		boolPtr(true),
		intPtr(1001),
		intPtr(1001),
	)
	// If called despite idle skip, fail loudly via error path (soft) — Assert
	// only requires zero agents; optional record.
	req.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		return nil, fmt.Errorf("resolve must not run for idle session (pid=%d)", pid)
	}
	return nil
}
```
