# Scenario

**Feature**: zero-window Snapshot attaches no agents

```
Snapshot{Windows:[]} + ResolveFromPID present -> Capture -> Agents empty
```

## Steps

1. Inject Snapshot with empty Windows and zero Summary.
2. Resolve inject still set (must not be called meaningfully; no sessions).

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Snapshot = &snapshot.Snapshot{
		CapturedAt: "2026-07-25T12:00:00Z",
		Host:       "testhost",
		Source:     "iterm2",
		Summary:    snapshot.SnapshotSummary{},
		Windows:    nil,
	}
	req.ResolveFromPID = func(pid int) (*procresolve.Result, error) {
		return nil, fmt.Errorf("resolve should not be called for empty snapshot (pid=%d)", pid)
	}
	return nil
}
```
