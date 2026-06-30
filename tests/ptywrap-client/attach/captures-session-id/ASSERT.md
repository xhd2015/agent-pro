## Expected

- Attach succeeds with fake TTY.
- `AttachResult.SessionID` equals `session-42`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.AttachErr != "" {
		t.Fatalf("unexpected attach error: %s", resp.AttachErr)
	}
	if resp.AttachID != "session-42" {
		t.Fatalf("session id: got %q want session-42", resp.AttachID)
	}
}
```