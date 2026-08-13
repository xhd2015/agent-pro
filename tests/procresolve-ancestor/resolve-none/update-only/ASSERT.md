## Expected

- No error.
- `FindAncestorGrok` ok=false.
- `Kind=none`, empty session id.

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	assertNoAncestor(t, resp)
	assertNone(t, assertResult(t, resp), pidStart)
}
```
