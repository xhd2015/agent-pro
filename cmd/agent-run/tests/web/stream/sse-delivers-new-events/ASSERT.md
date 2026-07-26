---
label: e2e
---

## Expected

- SSE payload includes user `message` with text `sse hello`.
- SSE payload includes at least one assistant `message` event before stream ends.

## Errors

- None from `Run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SSEEvents) == 0 {
		t.Fatal("expected at least one SSE event")
	}
	if !sseHasUserPrompt(resp.SSEEvents, req.CreatePrompt) {
		t.Fatalf("SSE missing user prompt %q: %v", req.CreatePrompt, resp.SSEEvents)
	}
	if !sseHasAssistantMessage(resp.SSEEvents) {
		t.Fatalf("SSE missing assistant message: %v", resp.SSEEvents)
	}
}
```