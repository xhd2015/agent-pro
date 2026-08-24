## Expected

- No error.
- `State` = `running`.
- `FileActive` = true.
- `PIDChecked` = true.
- `SessionID` = fixture id.
- Exactly one PID: pid `5001`, `Name` = `grok`, `Cmd` contains `grok`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	st := assertStatus(t, resp)

	assertEqualString(t, "SessionID", st.SessionID, req.SessionID)
	assertEqualString(t, "State", st.State, "running")
	assertEqualBool(t, "FileActive", st.FileActive, true)
	assertEqualBool(t, "PIDChecked", st.PIDChecked, true)
	if !strings.HasSuffix(st.Path, "summary.json") {
		t.Fatalf("Path = %q, want …/summary.json", st.Path)
	}
	if len(st.PIDs) != 1 {
		t.Fatalf("PIDs len = %d, want 1: %+v", len(st.PIDs), st.PIDs)
	}
	p := st.PIDs[0]
	assertEqualInt(t, "PID", p.PID, 5001)
	assertEqualString(t, "Name", p.Name, "grok")
	if !strings.Contains(p.Cmd, "grok") {
		t.Fatalf("Cmd = %q, want substring grok", p.Cmd)
	}
}
```
