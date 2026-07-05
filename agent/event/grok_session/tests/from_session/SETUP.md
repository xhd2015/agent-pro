# Scenario

**Feature**: forward conversion from grok session updates.jsonl wire to canonical AgentEvents

```
# each wire line is parsed and fed through a stateful converter
updates.jsonl line -> ParseLine -> Converter.ProcessLine -> []types.AgentEvent

# chunk coalescence and turn boundaries add grok_session extensions
Converter.Flush / turn_completed -> ActionDone + turn_index increment
```

## Preconditions

- `FromUpdatesJSONL` walks `req.WireLines`, coalesces chunks, and flushes at end.
- Each leaf sets `req.Target = "from_session"` and populates synthetic wire lines.

## Steps

1. Set `req.Target = "from_session"`.
2. Leaf SETUPs populate `req.WireLines` with flat or envelope ACP lines.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_session"
	return nil
}
```