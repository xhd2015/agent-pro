## Expected
- Trace output shows numbered blocks for think, three tool calls, and assistant message.
- Tool blocks use READ, GREP, and EDIT labels (not only thinking and ASSISTANT).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Stdout == "" {
		t.Fatalf("expected non-empty trace output")
	}

	for _, label := range []string{"READ", "GREP", "EDIT"} {
		if !strings.Contains(resp.Stdout, label) {
			t.Fatalf("stdout missing tool label %q, got:\n%s", label, resp.Stdout)
		}
	}

	toolBlocks := strings.Count(resp.Stdout, "📖") +
		strings.Count(resp.Stdout, "🔎") +
		strings.Count(resp.Stdout, "✏️")
	if toolBlocks < 3 {
		t.Fatalf("expected at least 3 tool blocks in trace, got %d, output:\n%s", toolBlocks, resp.Stdout)
	}

	if !strings.Contains(resp.Stdout, "ASSISTANT") {
		t.Fatalf("stdout missing ASSISTANT block, got:\n%s", resp.Stdout)
	}
	if strings.Count(resp.Stdout, "]  💭") != 1 {
		t.Fatalf("expected exactly 1 think block, got:\n%s", resp.Stdout)
	}
}
```