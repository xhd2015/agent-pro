# Scenario

**Feature**: idle from t=0 for timeout → SoftExit once; +grace → Shutdown once

```
Tick idle @0, @10s -> SoftExit=1, Shutdown=0
Tick idle @15s -> Shutdown=1
```

## Steps

1. Idle samples at 0, 10s, 15s (timeout 10s, default grace 5s).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
