# Scenario

**Feature**: hard errors when the input pid cannot be resolved in the snapshot

```
# pid absent from ListProcs
ResolveFromPID(missing) -> error containing "pid not found"
```

## Preconditions

- Prefer **error** over soft `Kind=none` when the pid is completely unknown.

## Steps

1. Leaf sets a PID that does not appear in `Procs`.
2. Assert non-nil error with stable message fragment.

## Context

- Distinguishes “process exists but no session” (`resolve-none`) from
  “caller handed a bogus pid” (`resolve-error`).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Error branch: leaves set PID outside the snapshot.
	if req.Procs == nil {
		req.Procs = []FixtureProc{}
	}
	return nil
}
```
