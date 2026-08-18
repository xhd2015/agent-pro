# Scenario

**Feature**: `exit_on_idle=false` writes OK; Tick never exits

```
WriteIdlePolicy({false, 10s}) -> Read found=true, ExitOnIdle=false
NewIdleWatchdog.Tick(idle past timeout+grace) -> SoftExit 0, Shutdown 0
```

## Steps

1. Write `ExitOnIdle=false`.
2. Read, then Tick idle at 0, 10s, 15s.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WritePolicy = true
	req.ExitOnIdle = false
	req.IdleTimeout = 10 * time.Second
	req.TickAfterPolicy = true
	req.SessionID = "sess-policy-off"
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
