---
label: e2e
---

## Expected

- HTTP 200.
- Terminal status is available despite `session.status:"finished"`.
- Session detail still reports `status:"finished"` and `terminal_session_id:"session-1"`.

## Side Effects

- None beyond isolated fixture files.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero until finished status no longer hides live terminal availability.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	requireTerminalMappingAvailable(t, req, resp.HTTPBody)
	status, detail := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.Runner+"/"+req.ChatSessionID, req.WebToken, "", "")
	if status != 200 {
		t.Fatalf("session detail status=%d body=%s", status, detail)
	}
	if !strings.Contains(detail, `"status":"finished"`) {
		t.Fatalf("session detail should remain finished: %s", detail)
	}
	if !strings.Contains(detail, `"terminal_session_id":"`+req.TerminalSessionID+`"`) {
		t.Fatalf("session detail should expose terminal_session_id: %s", detail)
	}
}
```
