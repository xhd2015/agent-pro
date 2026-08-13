## Expected

- `Main` error contains `no ancestor grok`.
- Exit 1.
- No OpenInNewTerminal.

## Side Effects

- None.

## Errors

- `no ancestor grok`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "no ancestor grok")
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
}
```
