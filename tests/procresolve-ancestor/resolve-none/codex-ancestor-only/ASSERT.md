## Expected

- No error.
- `FindAncestorGrok` ok=false.
- `Kind=none` (not `codex`).
- Empty `SessionID`.

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
	r := assertResult(t, resp)
	assertNone(t, r, pidStart)
	if r.Kind == "codex" {
		t.Fatal("Kind=codex; ancestor resolve is grok-only")
	}
}
```
