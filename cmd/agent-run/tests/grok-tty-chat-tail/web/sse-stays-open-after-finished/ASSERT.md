---
label: e2e
---

## Expected

- SSE delivers assistant `message` with `CHAT_TAIL_SSE_AFTER_FINISHED_MARKER`.
- Session was already `finished` before append — proves status is not a streaming gate.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ev := range resp.SSEEvents {
		if ev["type"] == "message" && ev["role"] == "assistant" && ev["text"] == chatTailSSEAfterFinishedMarker {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("SSE did not deliver append after finished status; events=%v", resp.SSEEvents)
	}
}
```