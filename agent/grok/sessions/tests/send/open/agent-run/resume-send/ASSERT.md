## Expected

- Agent-run resume window (agent-run argv, not bare grok --resume); no SendText.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.AgentRunCalls) != 1 {
		t.Fatalf("AgentRunCalls=%v", resp.AgentRunCalls)
	}
	if len(resp.Opened) != 1 {
		t.Fatalf("Opened=%v, want 1", resp.Opened)
	}
	if !strings.Contains(resp.Opened[0], "agent-run") {
		t.Fatalf("Opened want agent-run: %q", resp.Opened[0])
	}
	if strings.Contains(resp.Opened[0], "--resume "+req.SessionID) && !strings.Contains(resp.Opened[0], "agent-run") {
		t.Fatalf("must not bare grok --resume: %q", resp.Opened[0])
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
