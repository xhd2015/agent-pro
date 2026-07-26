## Expected

- Resolve returns session with id `session-7`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResolveErr != "" {
		t.Fatalf("unexpected resolve error: %s", resp.ResolveErr)
	}
	if resp.Resolved == nil || resp.Resolved.ID != "session-7" {
		t.Fatalf("resolved: %+v", resp.Resolved)
	}
}
```