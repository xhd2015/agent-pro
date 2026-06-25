## Preconditions
- Each leaf sets `req.Line` to the JSONL event line to format.

## Steps
1. Call `print.FormatTraceLine(req.Line)` directly.
2. Return the formatted output in `resp.Output`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/event/print"
)

func Setup(t *testing.T, req *Request) error {
	_ = assertContains
	_ = assertEmpty
	_ = assertNotContains
	return nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertEmpty(t *testing.T, got string) {
	t.Helper()
	if strings.TrimSpace(got) != "" {
		t.Fatalf("expected empty output, got:\n%s", got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
