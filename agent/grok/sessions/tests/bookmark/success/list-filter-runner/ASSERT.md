## Expected

- No error.
- Exactly one view; AgentRunner=grok; SessionID = fixture primary id.
- No codex row in Views.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Views) != 1 {
		t.Fatalf("Views len=%d want 1: %+v", len(resp.Views), resp.Views)
	}
	assertEqualString(t, "AgentRunner", resp.Views[0].AgentRunner, "grok")
	assertEqualString(t, "SessionID", resp.Views[0].SessionID, req.SessionID)
	for _, v := range resp.Views {
		if v.AgentRunner == "codex" {
			t.Fatalf("codex row should be filtered out: %+v", v)
		}
	}
}
```
