## Expected

- Persisted `events.jsonl` contains assistant message events.
- No event line includes a `phase` field (CLI-parity / StreamPhases:false).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	path, lines := readEventsJSONL(t, req.Home, req.Runner, req.SessionID)
	events := parseEventLines(t, lines)
	hasAssistant := false
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			hasAssistant = true
		}
	}
	if !hasAssistant {
		t.Fatalf("expected assistant message in %s: %v", path, events)
	}
	if eventsHavePhaseField(events) {
		t.Fatalf("web run emitted phased assistant events; events=%v", events)
	}
}
```
