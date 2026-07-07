# Scenario

**Feature**: CLI-parity events via WatchEvents and StreamPhases:false web runs

```
web run -> events.jsonl (no phase field)
SSE -> logs.WatchLine tail
sessions --print -> WatchEvents
```

## Preconditions

- Web runs use `StreamPhases: false` (same as CLI `run --json`).
- SSE `after` offset uses byte positions compatible with ReadEvents.

## Steps

1. Grouping setup sets `req.Area = "events"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Area = "events"
	return nil
}
```
