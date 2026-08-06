## Expected

- No error.
- `State` = `marked-active`.
- `FileActive` = true.
- `PIDChecked` = true.
- `PIDs` empty (length 0).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	st := assertStatus(t, resp)

	assertEqualString(t, "SessionID", st.SessionID, req.SessionID)
	assertEqualString(t, "State", st.State, "marked-active")
	assertEqualBool(t, "FileActive", st.FileActive, true)
	assertEqualBool(t, "PIDChecked", st.PIDChecked, true)
	if len(st.PIDs) != 0 {
		t.Fatalf("PIDs = %+v, want empty", st.PIDs)
	}
}
```
