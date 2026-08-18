# Scenario

**Feature**: sendable + screen idle + empty box + queue 0 is idle

```
SampleIsIdle({Sendable:true, Screen:idle, InputBox:empty, QueueLen:0}) -> true
```

## Steps

1. Keep default all-hold sample.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Sendable = true
	req.Screen = "idle"
	req.InputBox = "empty"
	req.QueueLen = 0
	return nil
}
```
