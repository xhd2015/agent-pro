## Expected

- Error contains `--color` and `--no-color` and `cannot be specified together`.
- Exit 1.

## Side Effects

- None.

## Errors

- `--color and --no-color cannot be specified together`

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainErr(t, resp, "--color", "--no-color", "cannot be specified together")
	assertNoOpen(t, resp)
}
```
