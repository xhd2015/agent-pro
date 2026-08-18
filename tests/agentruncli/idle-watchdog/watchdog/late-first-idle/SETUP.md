# Scenario

**Feature**: clock starts at first idle (slow boot cannot trip timeout)

```
busy @0, idle @8s, idle @10s -> SoftExit=0
idle @18s -> SoftExit=1
idle @23s -> Shutdown=1
```

## Steps

1. Busy at t=0 (clock not started).
2. First idle at 8s; timeout 10s → exit at 18s, shutdown at 23s.

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
		sampleAt(0, busySample()),
		idleAt(firstIdle),
		idleAt(defaultFakeTimeout),
		idleAt(firstIdle + defaultFakeTimeout),
		idleAt(firstIdle + defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
