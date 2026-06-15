## Expected
- One or more `EventMessage` events are received, each with valid `MessagePayload` (non-empty `id`, `role` = `"assistant"`, `session_id`, text parts).
- Exactly one `EventRunComplete` event is received.
- The total event sequence ends with `run_complete`.

## Side Effects
- Workspace is created (visible on the server).
- A session is created and messages are stored.

```go
import (
	"encoding/json"
	"testing"

	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("send-and-receive failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output with events JSON array")
	}
	var events []crush_types.Event
	if err := json.Unmarshal([]byte(resp.Output), &events); err != nil {
		t.Fatalf("failed to parse events: %v\noutput: %s", err, resp.Output)
	}
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	lastIdx := len(events) - 1
	if events[lastIdx].Type != crush_types.EventRunComplete {
		t.Fatalf("expected last event type %q, got %q", crush_types.EventRunComplete, events[lastIdx].Type)
	}
	hasMessage := false
	for i, e := range events {
		if e.Type == crush_types.EventMessage {
			hasMessage = true
			var msg crush_types.MessagePayload
			if err := json.Unmarshal(e.Payload, &msg); err != nil {
				t.Fatalf("event[%d]: failed to parse payload: %v", i, err)
			}
			if msg.ID == "" {
				t.Fatalf("event[%d]: expected non-empty message ID", i)
			}
			if msg.Role != "assistant" {
				t.Fatalf("event[%d]: expected role 'assistant', got %q", i, msg.Role)
			}
			if msg.SessionID == "" {
				t.Fatalf("event[%d]: expected non-empty session_id", i)
			}
		}
	}
	if !hasMessage {
		t.Fatal("expected at least one EventMessage before run_complete")
	}
}
```
