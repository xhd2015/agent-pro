## Expected

- No error; View non-nil; Orphaned=false.
- Output contains: grok, session id, title, chat/message count, session dir,
  tags (`show`/`detail`), description `user note for show`.

## Errors

- None.

```go
import (
	"strconv"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.View == nil {
		t.Fatal("View is nil")
	}
	assertEqualBool(t, "Orphaned", resp.View.Orphaned, false)
	assertEqualString(t, "SessionID", resp.View.SessionID, req.SessionID)
	out := resp.Output
	if out == "" {
		t.Fatal("show output empty")
	}
	assertContains(t, out, "grok")
	assertContains(t, out, req.SessionID)
	assertContains(t, out, req.Title)
	assertContains(t, out, strconv.Itoa(req.NumChatMessages))
	assertContains(t, out, req.SessionDir)
	assertContains(t, out, "show")
	assertContains(t, out, "detail")
	assertContains(t, out, "user note for show")
}
```
