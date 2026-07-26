## Expected

- `PopMessages` returns all three messages.
- Order is `oldest`, `middle`, `newest` (FIFO).
- A subsequent `ListMessages` would return empty (queue drained).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.Messages) != 3 {
		t.Fatalf("expected 3 popped messages, got %d", len(resp.Messages))
	}
	assertEqual(t, "messages[0].Text", resp.Messages[0].Text, "oldest")
	assertEqual(t, "messages[1].Text", resp.Messages[1].Text, "middle")
	assertEqual(t, "messages[2].Text", resp.Messages[2].Text, "newest")
}
```