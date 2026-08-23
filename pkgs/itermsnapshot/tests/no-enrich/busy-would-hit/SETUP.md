# Scenario

**Feature**: NoEnrich ignores busy pane even when resolve would hard-hit

```
# busy session + grok resolve available, but NoEnrich set by parent
NoEnrich + busy Snapshot + resolveHit(grok) -> Capture -> Agents empty
```

## Steps

1. Inject one busy session with PID.
2. Provide ResolveFromPID that would return a grok hard hit (must not attach).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Snapshot = oneBusySession(
		"sess-busy-noenrich",
		"busy-pane",
		"/dev/ttys010",
		4201,
	)
	req.ResolveFromPID = resolveHit("grok", "grok-sess-should-not-attach", "title", []procresolve.ProcNode{
		{PID: 4201, PPID: 1, Role: "input", Cmd: "agent-run"},
	})
	return nil
}
```
