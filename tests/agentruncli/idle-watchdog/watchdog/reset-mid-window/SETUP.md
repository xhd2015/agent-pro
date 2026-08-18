# Scenario

**Feature**: idle almost-timeout, then occupied, then idle almost-timeout → no SoftExit

```
idle @0, idle @9s, occupied @9s, idle @18s (timeout 10s) -> SoftExit=0
```

## Steps

1. Idle 9s, occupy (reset), idle another 9s. Each window < 10s.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	nine := 9 * defaultFakeTimeout / 10
	req.Steps = []TickStep{
		idleAt(0),
		idleAt(nine),
		sampleAt(nine, occupiedSample()),
		idleAt(nine + nine),
	}
	return nil
}
```
