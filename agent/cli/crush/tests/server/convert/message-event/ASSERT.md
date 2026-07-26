## Expected
- `UnwrapEvent` returns a non-nil `*crush_types.Event`.
- `Event.Type` is `EventMessage` (`"message"`).
- `Event.Payload` decodes to a `MessagePayload` with:
  - `id`: `"msg_abc123"`
  - `role`: `"assistant"`
  - `session_id`: `"sess_xyz789"`
  - `parts`: two entries — first is `text` type, second is `reasoning` type

## Errors
- No error.

```go
import (
	"encoding/json"
	"testing"

	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
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
	if event.Type != crush_types.EventMessage {
		t.Fatalf("expected EventType %q, got %q", crush_types.EventMessage, event.Type)
	}
	if len(event.Payload) == 0 {
		t.Fatal("expected non-empty Payload")
	}
	var msg crush_types.MessagePayload
	if err := json.Unmarshal(event.Payload, &msg); err != nil {
		t.Fatalf("failed to parse Payload as MessagePayload: %v\ndata: %s", err, string(event.Payload))
	}
	if msg.ID != "msg_abc123" {
		t.Fatalf("expected msg.ID %q, got %q", "msg_abc123", msg.ID)
	}
	if msg.Role != "assistant" {
		t.Fatalf("expected msg.Role %q, got %q", "assistant", msg.Role)
	}
	if msg.SessionID != "sess_xyz789" {
		t.Fatalf("expected msg.SessionID %q, got %q", "sess_xyz789", msg.SessionID)
	}
	if len(msg.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(msg.Parts))
	}
	if msg.Parts[0].Type != crush_types.PartText {
		t.Fatalf("expected part[0].Type %q, got %q", crush_types.PartText, msg.Parts[0].Type)
	}
	if msg.Parts[1].Type != crush_types.PartReasoning {
		t.Fatalf("expected part[1].Type %q, got %q", crush_types.PartReasoning, msg.Parts[1].Type)
	}
}
```
