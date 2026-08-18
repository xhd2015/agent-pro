# Scenario

**Feature**: `SampleIsIdle` — all four conditions vs any fail

```
SampleIsIdle(sendable, screen, box, queue) -> true only when all hold
```

## Steps

1. Grouping sets `Op=predicate`.
2. Root already seeded sendable + idle + empty + queue 0.
3. Leaves override the failing factor (or keep all-hold).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opPredicate
	return nil
}
```
