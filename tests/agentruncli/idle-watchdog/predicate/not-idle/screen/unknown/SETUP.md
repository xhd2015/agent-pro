# Scenario

**Feature**: screen unknown is not idle

```
SampleIsIdle({Screen:unknown, sendable, empty box, queue 0}) -> false
```

## Steps

1. Set `Screen=unknown`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Screen = "unknown"
	return nil
}
```
