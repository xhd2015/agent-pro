## Expected

- No error.
- `FindAncestorGrok` ok, `Ancestor.PID` = 4242 (self).
- Hard grok hit; `SessionID` from Lsof on 4242.

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
