## Expected

- Session reaches `finished` status.
- At least one `message` event with `role=assistant` has `timestamp` > 0.

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
	body, ts := waitForMessageTimestamp(t, req, req.SessionRunner, req.SessionID, "assistant", 2*time.Second)
	if ts <= 0 {
		t.Fatalf("expected assistant message timestamp > 0, got %v in %s", ts, body)
	}
}
```