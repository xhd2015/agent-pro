# Scenario

**Feature**: unknown Idle (nil) never gets an agent

```
Idle=nil session + resolve that would hit -> Capture -> Agents empty
```

## Steps

1. Inject one session with Idle=nil (unknown), PID set.
2. Resolve inject would hard-hit if called; attach must skip unknown first.

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
		"sess-unknown-1",
		"unknown-pane",
		"/dev/ttys099",
		nil, // Idle unknown
		intPtr(2002),
		nil,
	)
	req.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		return nil, fmt.Errorf("resolve must not run for unknown Idle (pid=%d)", pid)
	}
	return nil
}
```
