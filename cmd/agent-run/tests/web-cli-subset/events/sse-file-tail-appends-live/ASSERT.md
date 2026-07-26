---
label: e2e
---

## Expected

- SSE stream includes appended assistant text `Tail appended while SSE open`.
- Demonstrates file-tail delivery (WatchEvents), not poll-only replay.

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
		if ev["type"] == "message" && ev["role"] == "assistant" && ev["text"] == "Tail appended while SSE open" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("SSE did not tail appended event; events=%v", resp.SSEEvents)
	}
}
```
