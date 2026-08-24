## Expected

- Live agent-run prefer focuses; no grok `--resume` window.

```go
import "github.com/xhd2015/doctest/assert"

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
	assert.Output(t, resp.Stdout, `---
version: 3
---
focused: agent-run ar-open-live
`)
}
```
