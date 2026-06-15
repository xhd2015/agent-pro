## Expected
- `UnwrapEvent` returns a non-nil `*crush_types.Event`.
- `Event.Type` is `EventAgentEvent` (`"agent_event"`).
- `Event.Payload` decodes to an `AgentEventPayload` with:
  - `type`: `"error"`
  - `error`: `"rate limit exceeded"`
  - `run_id`: `"run_001"`

```go
import (
	"encoding/json"
	"testing"

	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var event crush_types.Event
	if err := json.Unmarshal([]byte(resp.Output), &event); err != nil {
		t.Fatalf("failed to parse output as Event: %v\noutput: %s", err, resp.Output)
	}
	if event.Type != crush_types.EventAgentEvent {
		t.Fatalf("expected EventType %q, got %q", crush_types.EventAgentEvent, event.Type)
	}
	var ae crush_types.AgentEventPayload
	if err := json.Unmarshal(event.Payload, &ae); err != nil {
		t.Fatalf("failed to parse Payload as AgentEventPayload: %v\ndata: %s", err, string(event.Payload))
	}
	if ae.Type != "error" {
		t.Fatalf("expected AgentEventPayload.Type %q, got %q", "error", ae.Type)
	}
	if ae.Error != "rate limit exceeded" {
		t.Fatalf("expected AgentEventPayload.Error %q, got %q", "rate limit exceeded", ae.Error)
	}
	if ae.RunID != "run_001" {
		t.Fatalf("expected AgentEventPayload.RunID %q, got %q", "run_001", ae.RunID)
	}
}
```
