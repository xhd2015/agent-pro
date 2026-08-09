# Scenario

**Feature**: RecentSet with non-positive Recent returns an error

```
RecentSet=true, Recent=0
  -> error (clear message); no panic
```

## Preconditions

- No reliance on fixtures; optional empty home is fine.
- Library must reject before treating zero as "no window".

## Steps

1. Set RecentSet=true, Recent=0.
2. Leave other filters zero.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentSet = true
	req.Recent = 0
	req.Limit = 10
	return nil
}
```
