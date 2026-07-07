## Expected

- SSE includes user prompt and assistant message.
- No SSE payload includes `phase` field.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SSEEvents) == 0 {
		t.Fatal("expected SSE events")
	}
	if !sseHasUserPrompt(resp.SSEEvents, req.Prompt) {
		t.Fatalf("SSE missing user prompt %q: %v", req.Prompt, resp.SSEEvents)
	}
	hasAssistant := false
	for _, ev := range resp.SSEEvents {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			hasAssistant = true
		}
		if _, ok := ev["phase"]; ok {
			t.Fatalf("SSE emitted phased event after CLI-parity refactor: %v", ev)
		}
	}
	if !hasAssistant {
		t.Fatalf("SSE missing assistant message: %v", resp.SSEEvents)
	}
}
```
