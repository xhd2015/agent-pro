## Expected

- No error.
- `FindAncestorGrok` ok, `Ancestor.PID` = 4242.
- `Kind` = `grok`.
- `SessionID` = `019fabcdef-1234-5678-9abc-def012345678` (Lsof, not `--resume`).
- `Confidence` = `hard`.
- `RunnerPID` / `InputPID` = 4242.

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
