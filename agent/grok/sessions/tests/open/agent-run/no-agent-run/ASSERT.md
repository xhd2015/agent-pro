## Expected

- `--no-agent-run` forces bare `grok --resume`; prefer hook not called.

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.AgentRunCalls) != 0 {
		t.Fatalf("AgentRunCalls=%v, want none", resp.AgentRunCalls)
	}
	assertResumeOpened(t, req, resp)
	assert.Output(t, resp.Stdout, `---
version: 3
---
opened: new window; resuming `+req.SessionID+`
`)
}
```
