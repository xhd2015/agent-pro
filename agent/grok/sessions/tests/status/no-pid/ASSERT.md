## Expected

- No error.
- `PIDChecked` = false.
- `PIDs` empty (injectables ignored).
- `FileActive` = true.
- `State` = `marked-active` (rollup from file only; not `running`).

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
	assertEqualBool(t, "PIDChecked", st.PIDChecked, false)
	assertEqualBool(t, "FileActive", st.FileActive, true)
	assertEqualString(t, "State", st.State, "marked-active")
	if len(st.PIDs) != 0 {
		t.Fatalf("PIDs = %+v, want empty when NoPID", st.PIDs)
	}
}
```
