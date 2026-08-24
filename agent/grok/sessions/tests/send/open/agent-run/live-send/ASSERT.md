## Expected

- Prefer live agent-run send; no grok `--resume` window; no SendText.

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
	if len(resp.Opened) != 0 {
		t.Fatalf("Opened=%v, want none", resp.Opened)
	}
	assertNoSend(t, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
sent to session `+req.SessionID+`
`)
	if strings.Contains(resp.Stdout, "resuming") {
		t.Fatalf("must not print bare resume ack:\n%s", resp.Stdout)
	}
}
```
