## Expected

- No error; pin Created=true.
- Views length 1; SessionID, AgentRunner=grok, Title, NumChatMessages match;
  Orphaned=false; tags include `roundtrip`.
- Warnings empty (session still present).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertEqualBool(t, "Created", resp.Created, true)
	if len(resp.Views) != 1 {
		t.Fatalf("Views len=%d want 1: %+v", len(resp.Views), resp.Views)
	}
	v := resp.Views[0]
	assertEqualString(t, "AgentRunner", v.AgentRunner, "grok")
	assertEqualString(t, "SessionID", v.SessionID, req.SessionID)
	assertEqualString(t, "Title", v.Title, req.Title)
	assertEqualInt(t, "NumChatMessages", v.NumChatMessages, req.NumChatMessages)
	assertEqualBool(t, "Orphaned", v.Orphaned, false)
	assertTagsEqual(t, v.Tags, []string{"roundtrip"})
	if len(resp.Warnings) != 0 {
		t.Fatalf("Warnings=%v want empty", resp.Warnings)
	}
}
```
