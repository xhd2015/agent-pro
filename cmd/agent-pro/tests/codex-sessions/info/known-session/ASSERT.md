## Expected

- `Info` returns populated `SessionInfo` with session metadata.
- `LineCount` reflects total JSONL lines; `NumDisplayEvents` is 1 (one agent_message).
- `TotalInputTokens` is 1500 and `TotalOutputTokens` is 600 (summed token_count events).
- `FormatInfoText` includes session id, title, relative last active (`2h ago`),
  status, line count, rollout file path, recent messages, and Tokens section.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}

	info := resp.Info
	if info.ID != knownInfoSessionID {
		t.Fatalf("ID = %q, want %q", info.ID, knownInfoSessionID)
	}
	if info.Title != "Refactor auth module" {
		t.Fatalf("Title = %q, want Refactor auth module", info.Title)
	}
	if info.LineCount < 6 {
		t.Fatalf("LineCount = %d, want at least 6", info.LineCount)
	}
	if info.NumDisplayEvents != 1 {
		t.Fatalf("NumDisplayEvents = %d, want 1", info.NumDisplayEvents)
	}
	if info.TotalInputTokens != 1500 || info.TotalOutputTokens != 600 {
		t.Fatalf("tokens = %d in / %d out, want 1500 / 600", info.TotalInputTokens, info.TotalOutputTokens)
	}
	if info.Path == "" {
		t.Fatal("rollout path is empty")
	}

	assertContains(t, resp.Output, "Session: "+knownInfoSessionID)
	assertContains(t, resp.Output, "Title: Refactor auth module")
	assertContains(t, resp.Output, "2h ago")
	assertContains(t, resp.Output, "Status:")
	assertContains(t, resp.Output, "Lines:")
	assertContains(t, resp.Output, "File:")
	assertContains(t, resp.Output, "Recent messages:")
	assertContains(t, resp.Output, "Analyzing auth flow")
	assertContains(t, resp.Output, "Tokens:")
	assertContains(t, resp.Output, "1500")
	assertContains(t, resp.Output, "600")
}
```