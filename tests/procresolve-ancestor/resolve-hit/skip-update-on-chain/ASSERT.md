## Expected

- No error.
- Ancestor pid = 4242 (real grok), not 4300 (`grok update`).
- `SessionID` is the real grok uuid, not the update process’s decoy path.

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
	if resp.Ancestor.PID == pidUpdate {
		t.Fatal("ancestor is grok update; must skip update utilities")
	}
	assertHardGrokHit(t, assertResult(t, resp), pidGrok, fixtureGrokSessionID)
}
```
