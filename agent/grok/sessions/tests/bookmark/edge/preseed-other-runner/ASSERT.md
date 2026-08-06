## Expected

- No error.
- Views length >= 2.
- Contains AgentRunner=grok with primary session id.
- Contains AgentRunner=codex with secondary session id and title
  `codex preseed title`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.Views) < 2 {
		t.Fatalf("Views len=%d want >= 2: %+v", len(resp.Views), resp.Views)
	}
	var sawGrok, sawCodex bool
	for _, v := range resp.Views {
		if v.AgentRunner == "grok" && v.SessionID == req.SessionID {
			sawGrok = true
		}
		if v.AgentRunner == "codex" && v.SessionID == fixtureBookmarkSessionID2 {
			sawCodex = true
			assertEqualString(t, "codex title", v.Title, "codex preseed title")
		}
	}
	if !sawGrok {
		t.Fatalf("missing grok view: %+v", resp.Views)
	}
	if !sawCodex {
		t.Fatalf("missing codex view: %+v", resp.Views)
	}
}
```
