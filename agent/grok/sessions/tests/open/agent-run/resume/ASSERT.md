## Expected

- Exited agent-run prefer opens agent-run resume window (not bare grok).

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Opened) != 1 {
		t.Fatalf("Opened=%v, want 1", resp.Opened)
	}
	if !strings.Contains(resp.Opened[0], "agent-run") {
		t.Fatalf("Opened want agent-run: %q", resp.Opened[0])
	}
	if strings.Contains(resp.Opened[0], "--resume "+req.SessionID) {
		t.Fatalf("must not bare grok --resume: %q", resp.Opened[0])
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; agent-run resume ar-open-resume
`)
}
```
