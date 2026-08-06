# Scenario

**Feature**: second `kill` on an already-stopped session is idempotent

```
startLiveKillFixture(kill-double-1)
agent-run kill kill-double-1          # first stop (Setup best-effort)
agent-run kill kill-double-1          # Run under test
  -> exit 0
  -> stderr: warning: session kill-double-1 not running
```

## Steps

1. Start live fixture with meta.
2. Best-effort first kill in Setup (succeeds once product lands).
3. Args = second `kill kill-double-1`.
4. Assert exit 0 + warning not running.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	const sid = "kill-double-1"
	startLiveKillFixture(t, req, sid, true)
	req.Mode = "handle"
	firstKillBestEffort(t, req, sid)
	req.Args = []string{"kill", sid}
	return nil
}
```
