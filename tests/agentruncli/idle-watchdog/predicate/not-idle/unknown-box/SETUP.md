# Scenario

**Feature**: unknown input box is not idle

```
SampleIsIdle({Sendable:true, Screen:idle, InputBox:unknown, QueueLen:0}) -> false
```

## Steps

1. Set `InputBox=unknown`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InputBox = "unknown"
	return nil
}
```
