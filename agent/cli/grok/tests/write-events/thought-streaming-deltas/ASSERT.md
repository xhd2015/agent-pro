## Expected
- Six grok thought streaming lines coalesce into exactly one `ActionThink` AgentEvent.
- The merged event text is "The user wants me to act".

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
	if len(resp.Lines) != 1 {
		t.Fatalf("expected 1 coalesced think AgentEvent, got %d lines: %v", len(resp.Lines), resp.Lines)
	}
	var ev eventtypes.AgentEvent
	if err := json.Unmarshal([]byte(resp.Lines[0]), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != eventtypes.ActionThink {
		t.Fatalf("expected type %q, got %q", eventtypes.ActionThink, ev.Type)
	}
	if ev.Text != "The user wants me to act" {
		t.Fatalf("expected merged text %q, got %q", "The user wants me to act", ev.Text)
	}
}
```