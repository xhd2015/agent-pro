# Scenario

**Feature**: ReconcileOnce heals sessions with empty events.jsonl

```
meta.status=finished + initial_prompt + grok updates on disk
  -> ReconcileOnce
  -> discovery + sync -> events.jsonl populated
```

## Preconditions

- No `events.jsonl` at start for heal leaf.
- Grok updates pre-seeded with prompt matching `meta.initial_prompt`.

## Steps

1. Grouping leaves choose heal vs skip-when-worker-active mode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.ReconcileTimeout <= 0 {
		req.ReconcileTimeout = 10 * time.Second
	}
	return nil
}
```