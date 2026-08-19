# Scenario

**Feature**: chrome-idle + unchanged jsonl size still SoftExit after 3 hits

```
idle+log 100 @0, @5s, @10s -> SoftExit=1; +grace -> Shutdown=1
```

First Tick is baseline (not a change). Same size on ticks 2 and 3 counts as idle.

## Steps

1. Three chrome-idle samples with the same `LogBytes`.
2. Fourth Tick at timeout+grace for Shutdown.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleLogAt(0, 100),
		idleLogAt(defaultFakeTimeout/2, 100),
		idleLogAt(defaultFakeTimeout, 100),
		idleLogAt(defaultFakeTimeout+defaultFakeGrace, 100),
	}
	return nil
}
```
