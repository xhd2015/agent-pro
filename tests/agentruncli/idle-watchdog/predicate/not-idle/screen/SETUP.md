# Scenario

**Feature**: screen `busy` | `unknown` | `starting` | `modal` is not idle

```
SampleIsIdle(screen in {busy, unknown, starting, modal}) -> false
```

## Steps

1. Keep sendable + empty box + queue 0.
2. Leaves set Screen.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opPredicate
	req.Sendable = true
	req.InputBox = "empty"
	req.QueueLen = 0
	return nil
}
```
