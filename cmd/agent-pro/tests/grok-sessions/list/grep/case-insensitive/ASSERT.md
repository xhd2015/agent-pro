## Expected

- Session is listed despite case mismatch between pattern and content.
- Output includes title hit and/or chat hit with original casing `LocalAgentXyz`.
- Pattern was lowercase `localagentxyz` (req.Grep).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if req.Grep != "localagentxyz" {
		t.Fatalf("req.Grep = %q, want localagentxyz", req.Grep)
	}
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900017-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q", resp.Sessions[0].ID)
	}
	assertContains(t, resp.Output, "LocalAgentXyz")
	// At least one hit line present.
	if !strings.Contains(resp.Output, "  summary.json:1:title:") &&
		!strings.Contains(resp.Output, "  chat_history.jsonl:") {
		t.Fatalf("expected at least one hit line in:\n%s", resp.Output)
	}
	// Prefer both hits when title and chat both match.
	assertContains(t, resp.Output, "  summary.json:1:title:")
	assertContains(t, resp.Output, "  chat_history.jsonl:1:user:")
}
```
