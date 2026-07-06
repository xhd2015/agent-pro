# Scenario

**Bug**: tail must not stop after first `turn_completed`; multi-turn sessions continue streaming

```
turn 1 (user + assistant + turn_completed) already on disk
  -> tail bootstrap reads turn 1
  -> turn 2 lines appended after turn_completed
  -> emit turn 2 events before ctx cancel
```

## Preconditions

- `TailUpdatesFromOffset` watches until `ctx.Done()`, not until `turn_completed`.
- Turn 2 content is appended only after turn 1 `turn_completed` is on disk.

## Steps

1. Grouping leaves seed turn 1 in `InitialLines`.
2. Schedule turn 2 user + assistant with a unique marker after a short delay.
3. Assert turn 2 marker appears in collected events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StartOffset = 0
	return nil
}
```