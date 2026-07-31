## Expected

- `FocusSession` returns a non-nil error (missing session id).
- `FocusITerm` is never called.

## Errors

- Non-nil from `FocusSession`.

## Exit Code

- N/A (library)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertError(t, err)
	assertNoFocus(t, resp)
}
```
