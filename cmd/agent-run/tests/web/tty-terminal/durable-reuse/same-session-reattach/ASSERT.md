---
label: e2e
---

## Expected

- Both terminal status responses are available.
- Both responses identify the same runner/session and do not imply a new backend terminal.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	first := decodeJSONBody(t, resp.HTTPBody)
	if !boolField(first, "available") {
		t.Fatalf("first terminal status unavailable: %s", resp.HTTPBody)
	}
	_, _ = doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.Runner+"/"+req.SessionID, req.WebToken, "", "")
	status, secondBody := doHTTP(t, "GET", req.WebBaseURL+terminalStatusPath(req.Runner, req.SessionID), req.WebToken, "", "")
	if status != 200 {
		t.Fatalf("second terminal status=%d body=%s", status, secondBody)
	}
	second := decodeJSONBody(t, secondBody)
	if !boolField(second, "available") {
		t.Fatalf("second terminal status unavailable: %s", secondBody)
	}
	if stringField(second, "session_id") != stringField(first, "session_id") || stringField(second, "runner") != stringField(first, "runner") {
		t.Fatalf("terminal identity changed across navigation: first=%s second=%s", resp.HTTPBody, secondBody)
	}
}
```
