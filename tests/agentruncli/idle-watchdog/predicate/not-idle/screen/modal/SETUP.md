# Scenario

**Feature**: screen modal is not idle

```
SampleIsIdle({Screen:modal, sendable, empty box, queue 0}) -> false
```

## Steps

1. Set `Screen=modal`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Screen = "modal"
	return nil
}
```
