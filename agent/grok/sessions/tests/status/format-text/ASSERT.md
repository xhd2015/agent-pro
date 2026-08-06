## Expected

- No error; structured Status is running.
- `FormatStatusText` output (case-insensitive where noted) includes:
  - state value `running`
  - file-active indicated as yes / true (word `yes` or `true` or `File` line)
  - pid `5001` and name `grok` on a PID line
- No requirement on exact column layout beyond readable lines.

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

	out := resp.Output
	if out == "" {
		t.Fatal("FormatStatusText output empty")
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "running") {
		t.Fatalf("text missing state running:\n%s", out)
	}
	// File active yes/true
	if !strings.Contains(lower, "yes") && !strings.Contains(lower, "true") && !strings.Contains(lower, "file") {
		t.Fatalf("text missing file-active indication:\n%s", out)
	}
	assertContains(t, out, "5001")
	assertContains(t, out, "grok")
}
```
