## Expected

- Agent-run resume window; prompt delivered in child; no SendText.

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
	if !strings.Contains(resp.Opened[0], "agent-run") || strings.Contains(resp.Opened[0], "--resume "+req.SessionID) {
		t.Fatalf("Opened want agent-run command, got %q", resp.Opened[0])
	}
	assertNoSend(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; agent-run resume ar-resume-1
sent to session `+req.SessionID+`
`)
}
```
