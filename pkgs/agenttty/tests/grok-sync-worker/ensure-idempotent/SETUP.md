# Scenario

**Feature**: concurrent `EnsureGrokSync` is idempotent — one worker per session

```
two goroutines call EnsureGrokSync concurrently
  -> registry holds one worker
  -> single updates line -> single user event
```

## Preconditions

- Pre-seed one complete ACP line on disk before worker starts.
- Both `EnsureGrokSync` calls use identical `(runner, sessionID)`.

## Steps

1. Pre-seed user + assistant + `turn_completed` in `InitialLines`.
2. Set `ConcurrentEnsure=true`.
3. Assert worker count 1 and no duplicate user event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ConcurrentEnsure = true
	return nil
}
```
