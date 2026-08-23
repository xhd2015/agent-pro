## Expected

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	want := strings.TrimPrefix(sessions.ForkHelp, "\n")
	if !strings.HasSuffix(want, "\n") {
		want += "\n"
	}
	if resp.Stdout != want {
		t.Fatalf("help mismatch:\ngot:\n%s\nwant:\n%s", resp.Stdout, want)
	}
	for _, need := range []string{"--tab", "--tab-index", "--new-window", "--new-terminal", "--dry-run"} {
		if !strings.Contains(resp.Stdout, need) {
			t.Fatalf("help missing %q", need)
		}
	}
}
```
