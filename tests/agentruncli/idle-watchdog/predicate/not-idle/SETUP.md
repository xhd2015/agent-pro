# Scenario

**Feature**: any failed idle factor → `SampleIsIdle` false

```
occupied | unknown box | queue>0 | not sendable | screen not idle
  -> SampleIsIdle false
```

## Steps

1. Keep sendable + idle + empty + queue 0 as the base.
2. Leaves flip exactly one factor.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opPredicate
	req.Sendable = true
	req.Screen = "idle"
	req.InputBox = "empty"
	req.QueueLen = 0
	return nil
}
```
