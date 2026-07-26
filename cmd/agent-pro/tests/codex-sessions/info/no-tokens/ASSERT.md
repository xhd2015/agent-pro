## Expected

- `Info` succeeds and returns summary fields.
- `TotalInputTokens` and `TotalOutputTokens` are zero.
- `FormatInfoText` includes session metadata and Recent messages.
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
	if resp.Info.TotalInputTokens != 0 || resp.Info.TotalOutputTokens != 0 {
		t.Fatalf("tokens = %d in / %d out, want 0 / 0",
			resp.Info.TotalInputTokens, resp.Info.TotalOutputTokens)
	}

	assertContains(t, resp.Output, "Session: "+noTokensSessionID)
	assertContains(t, resp.Output, "Title: Docs cleanup")
	assertContains(t, resp.Output, "Recent messages:")
	assertNotContains(t, resp.Output, "Tokens:")
}
```