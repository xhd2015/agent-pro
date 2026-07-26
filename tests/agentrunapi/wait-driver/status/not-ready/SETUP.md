# Scenario

**Feature**: starting + sendable no is not ready

```
IsSessionReadyFromStatus(starting + no) -> false
```

## Steps

1. Use not-ready fixture stdout.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StatusStdout = statusNotReadyFixture()
	return nil
}
```
