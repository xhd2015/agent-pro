# Scenario

**Feature**: busy every tick past timeout → no SoftExit

```
Tick busy @0, @10s, @20s -> SoftExit=0, Shutdown=0
```

## Steps

1. Busy samples at 0, timeout, 2×timeout.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		sampleAt(0, busySample()),
		sampleAt(defaultFakeTimeout, busySample()),
		sampleAt(2*defaultFakeTimeout, busySample()),
	}
	return nil
}
```
