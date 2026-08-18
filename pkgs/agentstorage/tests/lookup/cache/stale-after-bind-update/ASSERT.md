## Expected

- After update bind, Find returns `bind-later`.
- Stale warm cache is not trusted once generation bumps.

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
		t.Fatalf("Find error after bind update: %v", resp.Err)
	}
	if resp.Meta.SessionID != "bind-later" {
		t.Fatalf("SessionID=%q, want bind-later", resp.Meta.SessionID)
	}
	if resp.Meta.RunnerSessionID != req.QueryID {
		t.Fatalf("RunnerSessionID=%q, want %q", resp.Meta.RunnerSessionID, req.QueryID)
	}
}
```
