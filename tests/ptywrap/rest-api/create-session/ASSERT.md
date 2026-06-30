## Expected

- POST returns success status (200 or 201).
- Response JSON includes non-empty `id` matching `session-N` pattern.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 && resp.HTTPStatus != 201 {
		t.Fatalf("unexpected status %d body %v", resp.HTTPStatus, resp.CreateBody)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id in create response")
	}
	if !strings.HasPrefix(resp.SessionID, "session-") {
		t.Fatalf("id format: got %q", resp.SessionID)
	}
}
```