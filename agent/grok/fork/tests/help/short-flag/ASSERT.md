## Expected

- Same contract as `--help`: exit 0, required flags, no `-n` / `--new-terminal`.

## Side Effects

- None.

## Errors

- None.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	helpMentions(t, resp.Stdout)
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
}
```
