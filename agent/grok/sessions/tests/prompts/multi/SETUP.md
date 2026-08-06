# Scenario

**Feature**: ListPrompts multi-session selection by last_active and optional window

```
# multi path
write many sessions (newest-first by last_active_at)
  -> ListPrompts(opts: Now, Recent, RecentSet, Limit, LimitSet)
  -> []SessionPrompts
```

## Preconditions

- Default Op is `list`.
- Fixed Now = 2026-07-03T15:00:00Z from root Setup.
- Newest-first ordering by session `LastActiveAt`.
- When `RecentSet`, only prompts with Timestamp in `[Now-Recent, Now]` count;
  sessions with zero in-window prompts are skipped and do not count toward limit.

## Steps

1. Leaf seeds N sessions with known last_active and prompt timestamps.
2. Sets RecentSet / LimitSet matrix fields.
3. Assert list length, ids order, and per-session prompt filtering.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Op == "" {
		req.Op = "list"
	}
	return nil
}
```
