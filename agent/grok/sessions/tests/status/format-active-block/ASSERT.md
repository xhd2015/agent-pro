## Expected

- No error; Status is running.
- `FormatActiveBlock` output is non-empty and indicates:
  - file-active (yes/true/active wording)
  - live pid `5001` and/or name `grok`
- Suitable to append after existing `FormatInfoText` for `session info`.

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
	if strings.TrimSpace(out) == "" {
		t.Fatal("FormatActiveBlock output empty")
	}
	lower := strings.ToLower(out)
	// Dual signal: file + pid
	if !strings.Contains(lower, "file") && !strings.Contains(lower, "active") {
		t.Fatalf("active block missing file/active wording:\n%s", out)
	}
	if !strings.Contains(out, "5001") && !strings.Contains(lower, "grok") {
		t.Fatalf("active block missing pid/name:\n%s", out)
	}
}
```
