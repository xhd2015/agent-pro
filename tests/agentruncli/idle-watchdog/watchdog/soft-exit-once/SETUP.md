# Scenario

**Feature**: SoftExit fires only once even if Tick continues after timeout

```
idle @0, @10s, @11s, @12s, @15s, @16s
  -> SoftExit=1 (at 10s), Shutdown=1 (at 15s)
```

## Steps

1. Idle ticks at 0, timeout, +1s, +2s, +grace, +grace+1s.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + time.Second),
		idleAt(defaultFakeTimeout + 2*time.Second),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
		idleAt(defaultFakeTimeout + defaultFakeGrace + time.Second),
	}
	return nil
}
```
