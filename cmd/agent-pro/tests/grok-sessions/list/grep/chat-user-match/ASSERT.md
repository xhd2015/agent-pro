## Expected

- Exactly one matching session: `01900011-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
- Non-matching session `01900011-bbbb-...` is omitted.
- Output contains indented hit for chat user line:
  `  chat_history.jsonl:2:user:` and the pattern `GREP_CHAT_USER_TOKEN`.
- No `summary.json` hit lines for this pattern.
- No ANSI escapes.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900011-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q, want 01900011-aaaa-7aaa-aaaa-aaaaaaaaaaaa", resp.Sessions[0].ID)
	}
	assertNotContains(t, resp.Output, "01900011-bbbb-7bbb-bbbb-bbbbbbbbbbbb")
	assertContains(t, resp.Output, "  chat_history.jsonl:2:user:")
	assertContains(t, resp.Output, "GREP_CHAT_USER_TOKEN")
	assertNotContains(t, resp.Output, "summary.json:")
	assertNotContains(t, resp.Output, "\x1b[")

	// Snippet should include the user message text (newlines collapsed if any).
	if !strings.Contains(resp.Output, "please implement GREP_CHAT_USER_TOKEN") {
		t.Fatalf("missing user snippet text in:\n%s", resp.Output)
	}
}
```
