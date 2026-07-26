## Expected

- `Info` succeeds and returns summary fields.
- `FormatInfoText` includes Files section with `summary.json`.
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
	if resp.Info.ID != req.SessionID {
		t.Fatalf("ID = %q, want %q", resp.Info.ID, req.SessionID)
	}

	assertContains(t, resp.Output, "Session: "+req.SessionID)
	assertContains(t, resp.Output, "Title: Docs cleanup")
	assertContains(t, resp.Output, "Messages: 5 total, 3 chat")
	assertContains(t, resp.Output, "Files:")
	assertContains(t, resp.Output, "summary.json")
	assertNotContains(t, resp.Output, "Tokens:")
}
```