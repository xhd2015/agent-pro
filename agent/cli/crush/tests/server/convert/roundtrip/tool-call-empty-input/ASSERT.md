## Expected
- Round-trip produces exactly 1 event (no panic).
- Event type is `"tool_call"` (`ActionToolCall`).
- Event tool name is `"bash"`.
- No fatal error occurred (nil ToolInput does not cause panic during conversion).

## Errors
- No panic.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip tool-call-empty-input failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var events []types.AgentEvent
	if err := json.Unmarshal([]byte(resp.Output), &events); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, resp.Output)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != types.ActionToolCall {
		t.Fatalf("expected type %q, got %q", types.ActionToolCall, events[0].Type)
	}
	if events[0].Tool != "bash" {
		t.Fatalf("expected tool %q, got %q", "bash", events[0].Tool)
	}
}
```
