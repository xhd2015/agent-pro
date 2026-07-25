## Expected

- No error.
- `Kind` = `none` (`grok update` is not a session runner).
- `SessionID` empty.
- `Confidence` empty.
- `RunnerPID` = 0.
- `InputPID` = 500.

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
	r := assertResult(t, resp)

	assertEqualInt(t, "InputPID", r.InputPID, 500)
	assertEqualString(t, "Kind", r.Kind, "none")
	assertEqualString(t, "SessionID", r.SessionID, "")
	assertEqualString(t, "Confidence", r.Confidence, "")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 0)
}
```
