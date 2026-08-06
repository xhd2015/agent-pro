## Expected

- No error.
- `State` = `running` (live PIDs present even if not file-active).
- `PIDChecked` = true.
- Exactly two PIDs, sorted by PID ascending: `7001` then `7002`.
- Both `Name` = `grok`.
- Cmd for 7001 contains `/opt/homebrew/bin/grok`; for 7002 contains
  `/usr/local/bin/grok`.

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

	assertEqualString(t, "State", st.State, "running")
	assertEqualBool(t, "FileActive", st.FileActive, false)
	assertEqualBool(t, "PIDChecked", st.PIDChecked, true)
	if len(st.PIDs) != 2 {
		t.Fatalf("PIDs len = %d, want 2: %+v", len(st.PIDs), st.PIDs)
	}
	assertEqualInt(t, "PIDs[0].PID", st.PIDs[0].PID, 7001)
	assertEqualInt(t, "PIDs[1].PID", st.PIDs[1].PID, 7002)
	assertEqualString(t, "PIDs[0].Name", st.PIDs[0].Name, "grok")
	assertEqualString(t, "PIDs[1].Name", st.PIDs[1].Name, "grok")
	if !strings.Contains(st.PIDs[0].Cmd, "/opt/homebrew/bin/grok") {
		t.Fatalf("PIDs[0].Cmd = %q", st.PIDs[0].Cmd)
	}
	if !strings.Contains(st.PIDs[1].Cmd, "/usr/local/bin/grok") {
		t.Fatalf("PIDs[1].Cmd = %q", st.PIDs[1].Cmd)
	}
}
```
