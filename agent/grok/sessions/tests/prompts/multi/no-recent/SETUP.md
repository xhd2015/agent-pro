# Scenario

**Feature**: multi without --recent uses session limit defaults (not time window)

```
# !RecentSet
ListPrompts -> newest N sessions by last_active; ALL user prompts per session
default N=10 when !LimitSet; N from Limit when LimitSet
```

## Preconditions

- `RecentSet=false`, `Recent=0`.
- No time filtering of individual prompts.

## Steps

1. Leaf configures Limit/LimitSet and fixtures.
2. Assert session count and ordering.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RecentSet = false
	req.Recent = 0
	return nil
}
```
