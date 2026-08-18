## Expected

- List succeeds with length 2.
- Both `list-grok` and `list-codex` are present.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("List error: %v", resp.Err)
	}
	if len(resp.Metas) != 2 {
		t.Fatalf("List len=%d, want 2", len(resp.Metas))
	}
	seen := map[string]bool{}
	for _, m := range resp.Metas {
		seen[m.SessionID] = true
	}
	if !seen["list-grok"] || !seen["list-codex"] {
		t.Fatalf("List metas missing expected ids: %+v", resp.Metas)
	}
}
```
