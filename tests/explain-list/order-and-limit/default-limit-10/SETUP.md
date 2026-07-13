# Scenario

**Feature**: default limit is 10 when --limit is omitted

```
# 12 sessions seeded (question-00 oldest … question-11 newest)
explain list -> 10 shown of 12, limit 10; includes question-11, excludes question-01
```

## Preconditions

- Twelve valid one-turn sessions with increasing timestamps.

## Steps

1. Seed 12 sessions via `seedNSessions`.
2. Run plain `list` (no `--limit`).
3. Assert title and that newest 10 appear / oldest 2 do not.

## Context

- Newest are higher indices from `seedNSessions` (hourBase 10 → times climb).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = seedNSessions(12, 10)
	return nil
}
```
