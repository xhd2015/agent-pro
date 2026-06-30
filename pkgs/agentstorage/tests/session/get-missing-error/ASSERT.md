## Expected

- `GetSession` returns a non-nil error.
- Returned session pointer is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, resp.Err)
	if resp.Session != nil {
		t.Fatalf("expected nil session, got %+v", resp.Session)
	}
}
```