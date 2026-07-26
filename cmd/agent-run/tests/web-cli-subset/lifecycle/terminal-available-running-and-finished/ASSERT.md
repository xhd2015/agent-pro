---
label: e2e
---

## Expected

- Terminal status `available:true` while session is running.
- Terminal status `available:true` after session finishes (keep-tty).

```go
import (
	"net/http"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	_, runningBody := doHTTP(t, http.MethodGet, req.WebBaseURL+terminalStatusPath(req.Runner, req.ChatSessionID), req.WebToken, "", "")
	running := decodeJSONBody(t, runningBody)
	if !boolField(running, "available") {
		t.Fatalf("terminal not available while running: %s", runningBody)
	}
	waitForSessionStatus(t, req, req.Runner, req.ChatSessionID, "finished", 60*time.Second)
	_, finishedBody := doHTTP(t, http.MethodGet, req.WebBaseURL+terminalStatusPath(req.Runner, req.ChatSessionID), req.WebToken, "", "")
	finished := decodeJSONBody(t, finishedBody)
	if !boolField(finished, "available") {
		t.Fatalf("terminal not available after finish (keep-tty): %s", finishedBody)
	}
}
```
