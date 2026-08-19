# Scenario

**Feature**: idle from t=0 for timeout → SoftExit once; +grace → Shutdown once

```
Tick idle @0, @5s, @10s -> SoftExit=1, Shutdown=0
Tick idle @15s -> Shutdown=1
```

## Steps

1. Three consecutive idle samples at 0, T/2, T (timeout 10s); grace tick at 15s.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(defaultFakeTimeout / 2),
		idleAt(defaultFakeTimeout),
		idleAt(defaultFakeTimeout + defaultFakeGrace),
	}
	return nil
}
```
