---
label: e2e
---

## Expected

- SSE payload includes at least one `think` typed event.
- SSE payload includes at least one `error` typed event whose text starts with
  `Cannot resolve session id:`.
- SSE includes user `message` for the submitted prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !sseHasEventType(resp.SSEEvents, "think") {
		t.Fatalf("SSE missing think event type; events=%v", resp.SSEEvents)
	}
	if !sseHasEventType(resp.SSEEvents, "error") {
		t.Fatalf("SSE missing error event type; events=%v", resp.SSEEvents)
	}
	if !eventsHaveErrorPrefix(resp.SSEEvents, resolveErrorPrefix) {
		t.Fatalf("SSE missing error prefix %q; events=%v", resolveErrorPrefix, resp.SSEEvents)
	}
	foundUser := false
	for _, ev := range resp.SSEEvents {
		if ev["type"] == "message" && ev["role"] == "user" && ev["text"] == req.Prompt {
			foundUser = true
			break
		}
	}
	if !foundUser {
		t.Fatalf("SSE missing user prompt %q; events=%v", req.Prompt, resp.SSEEvents)
	}
}
```