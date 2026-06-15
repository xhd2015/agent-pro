## Expected
- `UnwrapEvent` returns a non-nil `*crush_types.Event`.
- `Event.Type` is `EventRunComplete` (`"run_complete"`).
- `Event.Payload` decodes to a `RunCompletePayload` with:
  - `session_id`: `"sess_xyz789"`
  - `run_id`: `"run_001"`
  - `message_id`: `"msg_def456"`
  - `text`: `"Task completed successfully"`
  - `error`: `""` (empty)
  - `cancelled`: `false`

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
	if event.Type != crush_types.EventRunComplete {
		t.Fatalf("expected EventType %q, got %q", crush_types.EventRunComplete, event.Type)
	}
	var rc crush_types.RunCompletePayload
	if err := json.Unmarshal(event.Payload, &rc); err != nil {
		t.Fatalf("failed to parse Payload as RunCompletePayload: %v\ndata: %s", err, string(event.Payload))
	}
	if rc.SessionID != "sess_xyz789" {
		t.Fatalf("expected SessionID %q, got %q", "sess_xyz789", rc.SessionID)
	}
	if rc.RunID != "run_001" {
		t.Fatalf("expected RunID %q, got %q", "run_001", rc.RunID)
	}
	if rc.MessageID != "msg_def456" {
		t.Fatalf("expected MessageID %q, got %q", "msg_def456", rc.MessageID)
	}
	if rc.Text != "Task completed successfully" {
		t.Fatalf("expected Text %q, got %q", "Task completed successfully", rc.Text)
	}
	if rc.Error != "" {
		t.Fatalf("expected empty Error, got %q", rc.Error)
	}
	if rc.Cancelled {
		t.Fatal("expected Cancelled == false")
	}
}
```
