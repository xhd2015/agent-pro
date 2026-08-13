## Expected

- Error contains `grok session not found`.
- Exit 1.
- No open / no foreground.

## Side Effects

- None.

## Errors

- `grok session not found`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "grok session not found")
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
}
```
