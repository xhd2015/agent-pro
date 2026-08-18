# Scenario

**Feature**: not sendable is not idle

```
SampleIsIdle({Sendable:false, Screen:idle, InputBox:empty, QueueLen:0}) -> false
```

## Steps

1. Set `Sendable=false`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Sendable = false
	return nil
}
```
