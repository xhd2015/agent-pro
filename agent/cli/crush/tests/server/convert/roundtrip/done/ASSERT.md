## Expected
- Round-trip produces exactly 1 event.
- Event type is `"done"` (`ActionDone`).
- Event text is empty (`ToCrush` does not store `Text` in `RunCompletePayload`, so it is lost in round-trip).
- Event ID starts with `"crush:"`.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("roundtrip done failed: %v", err)
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
	if events[0].Type != types.ActionDone {
		t.Fatalf("expected type %q, got %q", types.ActionDone, events[0].Type)
	}
	if events[0].Text != "" {
		t.Fatalf("expected empty text (Text not preserved in done round-trip), got %q", events[0].Text)
	}
	if !strings.HasPrefix(events[0].ID, "crush:") {
		t.Fatalf("expected ID to start with 'crush:', got %q", events[0].ID)
	}
}
```
