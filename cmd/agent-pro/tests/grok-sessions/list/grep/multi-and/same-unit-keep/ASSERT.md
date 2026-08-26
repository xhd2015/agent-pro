## Expected

- Exactly one session kept: `01900030-aaaa-...`.
- Split-unit session `01900030-bbbb-...` omitted (patterns on different lines).
- Hit line is the user message containing both tokens.
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
	if resp.Sessions[0].ID != "01900030-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q", resp.Sessions[0].ID)
	}
	assertNotContains(t, resp.Output, "01900030-bbbb-7bbb-bbbb-bbbbbbbbbbbb")
	assertNotContains(t, resp.Output, "\x1b[")
	assertContains(t, resp.Output, "  chat_history.jsonl:1:user:")
	if !strings.Contains(resp.Output, "AND_ALPHA") || !strings.Contains(resp.Output, "AND_BETA") {
		t.Fatalf("hit should include both tokens:\n%s", resp.Output)
	}
}
```
