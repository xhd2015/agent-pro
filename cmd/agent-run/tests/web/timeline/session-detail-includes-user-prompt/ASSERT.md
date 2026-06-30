## Expected

- HTTP **200** on session detail GET.
- `events` contains a `message` entry with `role=user` and `text=fix the bug`.

## Side Effects

- Background agent run may append assistant events after the user entry.

```go
import (
	"encoding/json"
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
	body := waitForUserEvent(t, req, req.SessionRunner, req.SessionID, req.CreatePrompt, 3*time.Second)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("parse detail JSON: %v", err)
	}
	if !eventsContainUserText(body, req.CreatePrompt) {
		t.Fatalf("events missing user prompt %q: %s", req.CreatePrompt, body)
	}
}
```