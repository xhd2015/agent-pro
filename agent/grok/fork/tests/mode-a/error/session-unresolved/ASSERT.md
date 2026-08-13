## Expected

- Error contains `session not resolved`.
- Exit 1; no open.

## Side Effects

- None.

## Errors

- `session not resolved`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "session not resolved")
	assertNoOpen(t, resp)
}
```
