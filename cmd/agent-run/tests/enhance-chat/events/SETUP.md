# Scenario

**Feature**: grok session binding lifecycle events in events.jsonl (no PTY fallback)

```
web POST grok-tty -> agenttty emits think/error -> events.jsonl single source of truth
success: mock seeds updates.jsonl
failure: empty GROK_HOME, PTY chrome stays out of events.jsonl
```

## Preconditions

- Web runs use `StreamPhases: false` (same as CLI `run --json`).
- `events.jsonl` is the only chat transcript source; PTY scrollback never appended.

## Steps

1. Grouping setup sets `req.Area = "events"` and `req.Mode = "events"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Area = "events"
	req.Mode = "events"
	return nil
}
```