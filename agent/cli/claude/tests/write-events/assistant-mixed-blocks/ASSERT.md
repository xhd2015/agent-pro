## Expected
- Exactly 3 AgentEvents are emitted, in order: `message`, `think`, `tool_call`.
- The `message` event text is "hi".
- The `think` event text is "why".
- The `tool_call` event tool is "Bash" and `tool_input.command` is "ls".

```go
import (
	"encoding/json"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Lines) != 3 {
		t.Fatalf("expected 3 AgentEvents (message, think, tool_call), got %d lines: %v", len(resp.Lines), resp.Lines)
	}

	wantTypes := []eventtypes.ActionType{
		eventtypes.ActionMessage,
		eventtypes.ActionThink,
		eventtypes.ActionToolCall,
	}
	for i, want := range wantTypes {
		var ev eventtypes.AgentEvent
		if err := json.Unmarshal([]byte(resp.Lines[i]), &ev); err != nil {
			t.Fatalf("unmarshal event %d: %v", i, err)
		}
		if ev.Type != want {
			t.Fatalf("event %d: type = %q, want %q; all lines = %v", i, ev.Type, want, resp.Lines)
		}
	}

	var msg eventtypes.AgentEvent
	json.Unmarshal([]byte(resp.Lines[0]), &msg)
	if msg.Text != "hi" {
		t.Fatalf("message text = %q, want %q", msg.Text, "hi")
	}

	var think eventtypes.AgentEvent
	json.Unmarshal([]byte(resp.Lines[1]), &think)
	if think.Text != "why" {
		t.Fatalf("think text = %q, want %q", think.Text, "why")
	}

	var call eventtypes.AgentEvent
	json.Unmarshal([]byte(resp.Lines[2]), &call)
	if call.Tool != "Bash" {
		t.Fatalf("tool_call tool = %q, want %q", call.Tool, "Bash")
	}
	if cmd, _ := call.ToolInput["command"].(string); cmd != "ls" {
		t.Fatalf("tool_call input.command = %v, want %q", call.ToolInput["command"], "ls")
	}
}
```
