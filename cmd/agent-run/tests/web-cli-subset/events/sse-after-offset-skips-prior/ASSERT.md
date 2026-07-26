---
label: e2e
---

## Expected

- SSE at EOF offset on finished session delivers zero parsed events.

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
		t.Fatalf("expected no SSE replay after offset, got %d: %v", len(resp.SSEEvents), resp.SSEEvents)
	}
}
```
