## Expected
- Mixed grok stream produces exactly 4 AgentEvent lines: 1 think, 2 tool_call, 1 message.
- Tool events use canonical `tool_call` type with normalized tool names `read` and `grep`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Lines) != 4 {
		t.Fatalf("expected 4 AgentEvents (1 think + 2 tool_call + 1 message), got %d lines: %v", len(resp.Lines), resp.Lines)
	}

	var types []eventtypes.ActionType
	var tools []string
	for _, line := range resp.Lines {
		var ev eventtypes.AgentEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		types = append(types, ev.Type)
		if ev.Type == eventtypes.ActionToolCall {
			tools = append(tools, ev.Tool)
		}
	}

	wantTypes := []eventtypes.ActionType{
		eventtypes.ActionThink,
		eventtypes.ActionToolCall,
		eventtypes.ActionToolCall,
		eventtypes.ActionMessage,
	}
	for i, want := range wantTypes {
		if types[i] != want {
			t.Fatalf("event %d: type = %q, want %q; all types = %v", i, types[i], want, types)
		}
	}

	gotTools := strings.Join(tools, ",")
	if gotTools != "read,grep" {
		t.Fatalf("tool_call tools = %q, want read,grep; lines = %v", gotTools, resp.Lines)
	}
}
```