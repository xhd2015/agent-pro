## Expected

- Find succeeds with `SessionID` `grok-hit` only (not ambiguous).
- Codex session with the same UUID is ignored by Find.

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
		t.Fatalf("Find error: %v", resp.Err)
	}
	if resp.Meta.SessionID != "grok-hit" {
		t.Fatalf("SessionID=%q, want grok-hit", resp.Meta.SessionID)
	}
	if resp.Meta.Runner != "grok-tty" {
		t.Fatalf("Runner=%q, want grok-tty", resp.Meta.Runner)
	}
}
```
