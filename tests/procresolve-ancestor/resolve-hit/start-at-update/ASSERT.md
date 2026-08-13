## Expected

- No error.
- Ancestor is the parent real grok (4242), not the update start pid.
- Hard session id from the parent grok’s open files.

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
	assertAncestor(t, resp, pidGrok)
	assertHardGrokHit(t, assertResult(t, resp), pidGrok, fixtureGrokSessionID)
}
```
