# Scenario

**Feature**: per-session `grok-sync.lock` uses non-blocking flock

```
TryAcquireSessionLock (holder)
  -> second TryAcquireSessionLock (LOCK_NB)
  -> second returns acquired=false without blocking
```

## Preconditions

- Lock file path: `<sessionDir>/grok-sync.lock`.

## Steps

1. Leaf configures `Mode=flock-nb`.
2. Assert first acquire succeeds; second fails non-blocking.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "flock-nb"
	return nil
}
```