## Expected

- `ListMessages` returns two messages in append order.
- Texts match `first msg` and `second msg`.
- Each message has non-empty `id` and matching `session_id`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(resp.Messages))
	}
	assertEqual(t, "messages[0].Text", resp.Messages[0].Text, "first msg")
	assertEqual(t, "messages[1].Text", resp.Messages[1].Text, "second msg")
	assertEqual(t, "messages[0].SessionID", resp.Messages[0].SessionID, req.SessionID)
	if resp.Messages[0].ID == "" || resp.Messages[1].ID == "" {
		t.Fatal("expected non-empty message IDs")
	}
}
```