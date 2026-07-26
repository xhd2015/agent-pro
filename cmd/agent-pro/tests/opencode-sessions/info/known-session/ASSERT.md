## Expected

- `Info` returns populated `SessionInfo` with session JSON fields.
- `NumMessages` is 3 (message file count).
- Token totals sum input/output across message files; cost is summed.
- `FormatInfoText` includes session id, title, relative last active (`2h ago`),
  CWD, Messages count, Files block with session and message paths, and Tokens/Cost section.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}

	info := resp.Info
	if info.ID != knownSessionID {
		t.Fatalf("ID = %q, want %q", info.ID, knownSessionID)
	}
	if info.Title != "Refactor auth module" {
		t.Fatalf("Title = %q, want Refactor auth module", info.Title)
	}
	if info.NumMessages != 3 {
		t.Fatalf("NumMessages = %d, want 3", info.NumMessages)
	}
	if info.TotalInputTokens != 350 || info.TotalOutputTokens != 150 {
		t.Fatalf("tokens = %d in / %d out, want 350 / 150", info.TotalInputTokens, info.TotalOutputTokens)
	}
	if info.TotalCost < 0.034 || info.TotalCost > 0.036 {
		t.Fatalf("TotalCost = %v, want ~0.035", info.TotalCost)
	}
	if info.SessionPath == "" || info.MessageDir == "" {
		t.Fatal("expected session and message paths")
	}

	assertContains(t, resp.Output, "Session: "+knownSessionID)
	assertContains(t, resp.Output, "Title: Refactor auth module")
	assertContains(t, resp.Output, "2h ago")
	assertContains(t, resp.Output, "Messages: 3")
	assertContains(t, resp.Output, "Files:")
	assertContains(t, resp.Output, "Tokens:")
	assertContains(t, resp.Output, "Cost:")
}
```