---
label: e2e
---

## Expected

- HTTP **200** on session detail after follow-up POST.
- `events` includes `role=user` message with `text=second prompt`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	body := waitForUserEvent(t, req, req.SessionRunner, req.SessionID, req.CreatePrompt, 3*time.Second)
	if !eventsContainUserText(body, req.CreatePrompt) {
		t.Fatalf("events missing follow-up user prompt %q: %s", req.CreatePrompt, body)
	}
}
```