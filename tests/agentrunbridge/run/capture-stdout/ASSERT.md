## Expected

- No API error.
- `resp.Stdout` is exactly `hello out` (trimmed).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertEqual(t, "Stdout", resp.Stdout, "hello out")
}
```
