## Expected

- No error.
- `Kind` = `none`.
- `SessionID` = empty string.
- `Confidence` = empty string.
- `RunnerPID` = 0 (or unset).
- `InputPID` = 400.

## Side Effects

- None.

## Errors

- None (soft miss, not hard failure).

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	r := assertResult(t, resp)

	assertEqualInt(t, "InputPID", r.InputPID, 400)
	assertEqualString(t, "Kind", r.Kind, "none")
	assertEqualString(t, "SessionID", r.SessionID, "")
	assertEqualString(t, "Confidence", r.Confidence, "")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 0)
}
```
