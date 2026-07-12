## Expected

- One session returned with at least four hits in display order:
  1. `summary.json:1:title:` containing `GREP_MULTI_TOKEN`
  2. `summary.json:1:session_summary:` containing `GREP_MULTI_TOKEN`
  3. `chat_history.jsonl:2:user:` containing `GREP_MULTI_TOKEN`
  4. `chat_history.jsonl:3:assistant:` containing `GREP_MULTI_TOKEN`
- No `... and N more matches` line (hits ≤ 5).
- Hit lines appear under the session row with two-space indent.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900012-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q", resp.Sessions[0].ID)
	}

	wantHits := []string{
		"  summary.json:1:title:",
		"  summary.json:1:session_summary:",
		"  chat_history.jsonl:2:user:",
		"  chat_history.jsonl:3:assistant:",
	}
	out := resp.Output
	pos := 0
	for _, h := range wantHits {
		i := strings.Index(out[pos:], h)
		if i < 0 {
			t.Fatalf("missing hit %q (after pos %d) in:\n%s", h, pos, out)
		}
		pos += i + len(h)
	}
	assertContains(t, out, "GREP_MULTI_TOKEN")
	assertNotContains(t, out, "... and ")
	assertNotContains(t, out, "\x1b[")

	if len(resp.Matches) == 1 && len(resp.Matches[0].Hits) < 4 {
		t.Fatalf("len(hits) = %d, want >= 4", len(resp.Matches[0].Hits))
	}
}
```
