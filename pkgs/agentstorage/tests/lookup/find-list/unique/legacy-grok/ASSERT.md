## Expected

- Find succeeds with `SessionID` `legacy-gsid`, `Runner` `grok`.

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
	if resp.Meta.SessionID != "legacy-gsid" {
		t.Fatalf("SessionID=%q, want legacy-gsid", resp.Meta.SessionID)
	}
	if resp.Meta.Runner != "grok" {
		t.Fatalf("Runner=%q, want grok", resp.Meta.Runner)
	}
}
```
