# Scenario

**Feature**: start pid missing from the snapshot is a hard error

```
ListProcs omits start pid
FindAncestorGrok -> ok=false
ResolveFromAncestors -> error ("pid not found")
```

## Preconditions

- Prefer **error** over soft `Kind=none` when the start pid is unknown.

## Steps

1. Leaf sets a PID that does not appear in `Procs`.
2. Assert non-nil error containing `pid not found`.

## Context

- Same error fragment as `ResolveFromPID`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Procs == nil {
		req.Procs = []FixtureProc{}
	}
	return nil
}
```
