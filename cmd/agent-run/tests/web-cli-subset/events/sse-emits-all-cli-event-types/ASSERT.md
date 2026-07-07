## Expected

- SSE includes at least one `done` typed event (CLI parity — not filtered to messages only).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !sseHasEventType(resp.SSEEvents, "done") {
		t.Fatalf("SSE missing done event type; events=%v", resp.SSEEvents)
	}
	if !sseHasUserPrompt(resp.SSEEvents, req.Prompt) {
		t.Fatalf("SSE missing user prompt %q", req.Prompt)
	}
}
```
