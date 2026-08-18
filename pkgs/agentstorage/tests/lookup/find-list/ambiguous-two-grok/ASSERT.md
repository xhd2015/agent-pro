## Expected

- Find error exactly:
  `ambiguous grok-session-id <uuid>: multiple matches: sess-a, sess-b`
  (session ids ascending, comma-space joined).
- List succeeds with length 2; both `sess-a` and `sess-b` present.

```go
import (
	"fmt"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	want := fmt.Sprintf("ambiguous grok-session-id %s: multiple matches: sess-a, sess-b", req.QueryID)
	assertExactErr(t, resp.FindErr, want)
	if resp.ListErr != nil {
		t.Fatalf("List error: %v", resp.ListErr)
	}
	if len(resp.Metas) != 2 {
		t.Fatalf("List len=%d, want 2", len(resp.Metas))
	}
	seen := map[string]bool{}
	for _, m := range resp.Metas {
		seen[m.SessionID] = true
	}
	if !seen["sess-a"] || !seen["sess-b"] {
		t.Fatalf("List metas missing sess-a/sess-b: %+v", resp.Metas)
	}
}
```
