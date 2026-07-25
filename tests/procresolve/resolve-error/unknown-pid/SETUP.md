# Scenario

**Feature**: pid not present in ListProcs returns pid-not-found error

```
ListProcs = [pid 1 init only]
ResolveFromPID(99999) -> error ("pid not found")
```

## Preconditions

- Snapshot does not contain 99999.
- OpenFiles irrelevant.

## Steps

1. Set `PID=99999`.
2. Install only an unrelated init-like proc (or empty snapshot). Prefer a
   non-empty snapshot that simply omits the target pid.

## Context

- Error text must include `pid not found` (case-sensitive substring).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 99999
	req.Procs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/init"},
	}
	req.OpenFiles = map[int][]string{}
	return nil
}
```
