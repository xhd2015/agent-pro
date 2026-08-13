## Expected

- Error contains `--pid` and `--session-id` and `cannot be specified together`.
- Exit 1; no launch.

## Side Effects

- None.

## Errors

- `--pid and --session-id cannot be specified together`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "--pid", "--session-id", "cannot be specified together")
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
}
```
