# Scenario

**Feature**: R1 — non-blocking second flock acquirer skips

```
session dir exists
  -> TryAcquireSessionLock #1 (hold)
  -> TryAcquireSessionLock #2 (LOCK_NB)
  -> #2 acquired=false
```

## Steps

1. Create empty session dir.
2. Acquire lock twice; release first on defer.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "flock-nb"
	req.SessionID = "lock-nb-test"
	return nil
}
```