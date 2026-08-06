## Expected

- No error.
- `State` = `inactive`.
- `FileActive` = false.
- `PIDChecked` = true.
- `PIDs` empty.

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
	assertEqualString(t, "State", st.State, "inactive")
	assertEqualBool(t, "FileActive", st.FileActive, false)
	assertEqualBool(t, "PIDChecked", st.PIDChecked, true)
	if len(st.PIDs) != 0 {
		t.Fatalf("PIDs = %+v, want empty", st.PIDs)
	}
}
```
