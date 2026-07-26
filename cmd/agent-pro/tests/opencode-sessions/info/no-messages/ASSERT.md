## Expected

- `Info` succeeds and returns summary fields.
- `NumMessages` is 0.
- `FormatInfoText` includes Files section with session path.
- Output does **not** include a `Tokens:` section.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}
	if resp.Info.ID != noMessagesSessionID {
		t.Fatalf("ID = %q, want %q", resp.Info.ID, noMessagesSessionID)
	}
	if resp.Info.NumMessages != 0 {
		t.Fatalf("NumMessages = %d, want 0", resp.Info.NumMessages)
	}

	assertContains(t, resp.Output, "Session: "+noMessagesSessionID)
	assertContains(t, resp.Output, "Title: Docs cleanup")
	assertContains(t, resp.Output, "Messages: 0")
	assertContains(t, resp.Output, "Files:")
	assertNotContains(t, resp.Output, "Tokens:")
}
```