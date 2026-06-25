## Preconditions
- `print.FormatState.FormatLine` coalesces consecutive streaming events for trace display.
- `traceSession` and `FollowEventLog` loop events through `FormatState`; this harness mirrors that loop.

## Steps
1. Leaf tests set `req.Lines` to AgentEvent JSONL strings (simulating events.jsonl).
2. The shared harness feeds each line through `FormatState.FormatLine`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = countSubstring
	return nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func countSubstring(t *testing.T, got string, sub string) int {
	t.Helper()
	return strings.Count(got, sub)
}
```
