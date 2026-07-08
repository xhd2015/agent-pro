## Expected

- SSE stream includes assistant `message` with `CHAT_TAIL_ASSISTANT_MARKER`.
- Demonstrates web path receives full tool-using turn, not only initial pending tool/progress cards.

## Errors

- Must not end SSE before delayed assistant when session status flips `finished`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasAssistantMarker {
		t.Fatalf("SSE missing assistant marker %q; sse=%v events.jsonl=%v",
			chatTailAssistantMarker, resp.SSEEvents, resp.EventsParsed)
	}
}
```