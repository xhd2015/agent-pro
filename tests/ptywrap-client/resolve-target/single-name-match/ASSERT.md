## Expected

- Resolve by name `my-grok` returns session-2.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResolveErr != "" {
		t.Fatalf("unexpected resolve error: %s", resp.ResolveErr)
	}
	if resp.Resolved == nil || resp.Resolved.ID != "session-2" {
		t.Fatalf("resolved: %+v", resp.Resolved)
	}
}
```