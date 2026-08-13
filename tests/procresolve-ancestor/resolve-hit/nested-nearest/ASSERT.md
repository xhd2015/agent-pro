## Expected

- No error.
- Ancestor pid = 4242 (subagent), not 3000 (main).
- `SessionID` = nearest fixture uuid, **not** the main uuid.

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
	r := assertResult(t, resp)
	assertHardGrokHit(t, r, pidGrok, fixtureGrokSessionID)
	if r.SessionID == fixtureMainGrokSessionID {
		t.Fatalf("SessionID picked topmost main grok %q, want nearest %q", r.SessionID, fixtureGrokSessionID)
	}
	if resp.Ancestor.PID == pidMainGrok {
		t.Fatal("ancestor is topmost main grok; want nearest subagent")
	}
}
```
