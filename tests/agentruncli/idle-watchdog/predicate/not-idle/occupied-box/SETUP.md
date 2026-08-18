# Scenario

**Feature**: occupied input box is not idle (even if sendable)

```
SampleIsIdle({Sendable:true, Screen:idle, InputBox:occupied, QueueLen:0}) -> false
```

## Steps

1. Set `InputBox=occupied`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InputBox = "occupied"
	return nil
}
```
