# Scenario

**Feature**: snapshot before/after compares for occupy + changed

```
occupied.ExactlyOneMoreSpace(before, after) -> bool
changed.Changed(before, after) -> bool
```

## Preconditions

- Nested root under `tests/tty/detection`.
- Pure in-process; no TTY.

## Steps

1. Grouping sets `Op`.
2. Leaf sets Before/After.
3. Assert checks the boolean.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	_ = req
	return nil
}
```
