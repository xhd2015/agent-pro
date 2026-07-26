## Expected
- Round-trip produces exactly 4 events (count preserved).
- Event order is preserved: think, message, tool-call, done.
- Types are correctly mapped: `"think"`, `"message"`, `"tool_call"`, `"done"`.
- Text and tool fields are preserved where applicable.

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
		t.Fatalf("roundtrip multi-event failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output")
	}
	var events []types.AgentEvent
	if err := json.Unmarshal([]byte(resp.Output), &events); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, resp.Output)
	}
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	if events[0].Type != types.ActionThink {
		t.Fatalf("expected events[0].Type %q, got %q", types.ActionThink, events[0].Type)
	}
	if events[0].Text != "thinking" {
		t.Fatalf("expected events[0].Text %q, got %q", "thinking", events[0].Text)
	}
	if events[1].Type != types.ActionMessage {
		t.Fatalf("expected events[1].Type %q, got %q", types.ActionMessage, events[1].Type)
	}
	if events[1].Text != "hello" {
		t.Fatalf("expected events[1].Text %q, got %q", "hello", events[1].Text)
	}
	if events[2].Type != types.ActionToolCall {
		t.Fatalf("expected events[2].Type %q, got %q", types.ActionToolCall, events[2].Type)
	}
	if events[2].Tool != "bash" {
		t.Fatalf("expected events[2].Tool %q, got %q", "bash", events[2].Tool)
	}
	if events[2].ToolInput["cmd"] != "ls" {
		t.Fatalf("expected events[2].ToolInput[\"cmd\"] = %q, got %q", "ls", events[2].ToolInput["cmd"])
	}
	if events[3].Type != types.ActionDone {
		t.Fatalf("expected events[3].Type %q, got %q", types.ActionDone, events[3].Type)
	}
}
```
