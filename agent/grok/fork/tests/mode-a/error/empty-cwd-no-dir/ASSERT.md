## Expected

- Error contains `pass --dir`.
- Exit 1; no open.

## Side Effects

- None.

## Errors

- `pass --dir`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "pass --dir")
	assertNoOpen(t, resp)
}
```
