# Scenario

**Feature**: screen busy is not idle

```
SampleIsIdle({Screen:busy, sendable, empty box, queue 0}) -> false
```

## Steps

1. Set `Screen=busy`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Screen = "busy"
	return nil
}
```
