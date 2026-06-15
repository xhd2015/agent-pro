## Expected
- Round-trip produces exactly 1 event.
- Event type is `"tool_call"` (`ActionToolCall`).
- Event tool name is `"read"`.
- Event tool input contains `"path": "/foo"`.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip tool-call failed: %v", err)
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
	if events[0].Tool != "read" {
		t.Fatalf("expected tool %q, got %q", "read", events[0].Tool)
	}
	if events[0].ToolInput == nil {
		t.Fatal("expected non-nil ToolInput")
	}
	path, ok := events[0].ToolInput["path"].(string)
	if !ok {
		t.Fatalf("expected ToolInput[\"path\"] to be string, got %T", events[0].ToolInput["path"])
	}
	if path != "/foo" {
		t.Fatalf("expected ToolInput[\"path\"] = %q, got %q", "/foo", path)
	}
}
```
