# Scenario

**Feature**: WaitReady polls StatusFn until ready or timeout

```
WaitReady(SessionID, StatusFn, Timeout, PollInterval)
  -> success | timeout error
```

## Steps

1. Set mode `wait_ready`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "wait_ready"
	return nil
}
```
