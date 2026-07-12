## Expected

- One matching session returned.
- Output shows exactly five indented hit lines of the form
  `  chat_history.jsonl:<n>:<part>: ...`.
- Output includes exactly one overflow line: `  ... and 3 more matches`.
- Structured `Matches[0].Hits` has total hit count 8 (all hits before display cap).

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

	if len(resp.Matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(resp.Matches))
	}
	if got := len(resp.Matches[0].Hits); got != 8 {
		t.Fatalf("len(hits) = %d, want 8 (full hit list before display cap)", got)
	}

	out := resp.Output
	assertContains(t, out, "  ... and 3 more matches")

	// Count indented hit lines (file:line:part:), not the overflow line.
	hitLines := 0
	for _, line := range strings.Split(out, "\n") {
		trim := strings.TrimPrefix(line, "  ")
		if strings.HasPrefix(line, "  ") && strings.Contains(trim, ":") && !strings.HasPrefix(trim, "... ") {
			// hit shape starts with filename basename
			if strings.HasPrefix(trim, "chat_history.jsonl:") || strings.HasPrefix(trim, "summary.json:") {
				hitLines++
			}
		}
	}
	if hitLines != 5 {
		t.Fatalf("displayed hit lines = %d, want 5; output:\n%s", hitLines, out)
	}

	// First five physical chat lines should be preferred in file order.
	assertContains(t, out, "  chat_history.jsonl:1:user:")
	assertContains(t, out, "  chat_history.jsonl:5:")
	// Lines 6-8 should not appear as displayed hits when cap is 5.
	assertNotContains(t, out, "  chat_history.jsonl:6:")
	assertNotContains(t, out, "  chat_history.jsonl:7:")
	assertNotContains(t, out, "  chat_history.jsonl:8:")
	assertNotContains(t, out, "\x1b[")
}
```
