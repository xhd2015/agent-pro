# Scenario

**Feature**: occupied/busy samples do not count; SoftExit after three later empties

```
occupied @0, occupied @5s
idle @8s, idle @10s, idle @18s -> SoftExit=1 at 18s
idle @23s -> Shutdown=1
```

## Steps

1. Occupied probes hold (no idle hits) through early ticks.
2. First empty at 8s; third empty at 18s → SoftExit; +grace → Shutdown.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	firstIdle := 8 * time.Second
	req.Steps = []TickStep{
		occupiedAt(0),
		occupiedAt(5 * time.Second),
		idleAt(firstIdle),
		idleAt(defaultFakeTimeout),
		idleAt(firstIdle + defaultFakeTimeout),
		idleAt(firstIdle + defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
