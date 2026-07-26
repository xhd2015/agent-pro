## Expected

- `Info` returns empty title in struct; formatted output shows `(untitled)`.
- Message line shows `0 total, 1 chat`.
- `updates.jsonl` is reported as missing.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}
	if resp.Info.Title != "" {
		t.Fatalf("Title = %q, want empty", resp.Info.Title)
	}
	if resp.Info.NumChatMessages != 1 {
		t.Fatalf("NumChatMessages = %d, want 1", resp.Info.NumChatMessages)
	}

	assertContains(t, resp.Output, "Session: "+req.SessionID)
	assertContains(t, resp.Output, "Title: (untitled)")
	assertContains(t, resp.Output, "Messages: 0 total, 1 chat")
	assertContains(t, resp.Output, "updates.jsonl (missing)")
}
```