## Expected

- No error.
- `FindAncestorGrok` ok=false.
- `Kind=none`, empty `SessionID`.
- Must not return the descendant decoy session.

## Side Effects

- None.

## Errors

- None (soft miss).

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
