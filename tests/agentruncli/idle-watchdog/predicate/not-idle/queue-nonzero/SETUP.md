# Scenario

**Feature**: non-empty send queue is not idle

```
SampleIsIdle({Sendable:true, Screen:idle, InputBox:empty, QueueLen:1}) -> false
```

## Steps

1. Set `QueueLen=1`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.QueueLen = 1
	return nil
}
```
