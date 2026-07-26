## Expected

- `RelocateCWD` returns a non-nil error.
- Result is nil.

## Errors

- Empty / required session id (or not found for empty id).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	assertError(t, resp)
	if resp.Result != nil {
		t.Fatalf("expected nil Result on error, got %+v", resp.Result)
	}
}
```
