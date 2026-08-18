## Expected

- Find succeeds with `SessionID` `hello-gsid`, `Runner` `grok-tty`.
- `FindErr` / `Err` are nil.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil || resp.FindErr != nil {
		t.Fatalf("Find error: %v", resp.Err)
	}
	if resp.Meta.SessionID != "hello-gsid" {
		t.Fatalf("SessionID=%q, want hello-gsid", resp.Meta.SessionID)
	}
	if resp.Meta.Runner != "grok-tty" {
		t.Fatalf("Runner=%q, want grok-tty", resp.Meta.Runner)
	}
	if resp.Meta.RunnerSessionID != req.QueryID {
		t.Fatalf("RunnerSessionID=%q, want %q", resp.Meta.RunnerSessionID, req.QueryID)
	}
}
```
