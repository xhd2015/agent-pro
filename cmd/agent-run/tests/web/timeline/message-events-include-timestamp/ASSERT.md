## Expected

- HTTP **200** on session detail GET.
- `events` contains a `message` with `role=user` and numeric `timestamp` > 0.

## Side Effects

- Assistant events may appear after the user entry once fake-codex completes.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	body, ts := waitForMessageTimestamp(t, req, req.SessionRunner, req.SessionID, "user", 5*time.Second)
	if !eventsContainUserText(body, req.CreatePrompt) {
		t.Fatalf("events missing user prompt %q: %s", req.CreatePrompt, body)
	}
	if ts <= 0 {
		t.Fatalf("expected user message timestamp > 0, got %v in %s", ts, body)
	}
}
```