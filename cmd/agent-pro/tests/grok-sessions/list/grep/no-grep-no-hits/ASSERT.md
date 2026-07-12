## Expected

- Classic list succeeds and includes the session row.
- Output has table headers and session id/title.
- Output does **not** contain indented hit prefixes `  summary.json:` or
  `  chat_history.jsonl:`.
- Output does not contain `... and ` overflow lines.

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
	assertContains(t, resp.Output, "SESSION ID")
	assertContains(t, resp.Output, "01900015-aaaa-7aaa-aaaa-aaaaaaaaaaaa")
	assertContains(t, resp.Output, "Session with history but no grep")

	for _, line := range strings.Split(resp.Output, "\n") {
		if strings.HasPrefix(line, "  summary.json:") || strings.HasPrefix(line, "  chat_history.jsonl:") {
			t.Fatalf("unexpected hit line without grep: %q", line)
		}
		if strings.HasPrefix(line, "  ... and ") {
			t.Fatalf("unexpected overflow line without grep: %q", line)
		}
	}
	// Matches should not be populated on classic path.
	if len(resp.Matches) != 0 {
		t.Fatalf("len(matches) = %d, want 0 when Grep is empty", len(resp.Matches))
	}
}
```
