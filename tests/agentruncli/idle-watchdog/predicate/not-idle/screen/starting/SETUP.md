# Scenario

**Feature**: screen starting is not idle

```
SampleIsIdle({Screen:starting, sendable, empty box, queue 0}) -> false
```

## Steps

1. Set `Screen=starting`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Screen = "starting"
	return nil
}
```
