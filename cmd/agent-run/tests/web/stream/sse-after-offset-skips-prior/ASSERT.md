---
label: e2e
---

## Expected

- When subscribing at end-of-file offset on a **finished** session, SSE delivers zero parsed events (idle tail).

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
	if len(resp.SSEEvents) != 0 {
		t.Fatalf("expected no SSE events after end offset on finished session, got %d: %v", len(resp.SSEEvents), resp.SSEEvents)
	}
}
```