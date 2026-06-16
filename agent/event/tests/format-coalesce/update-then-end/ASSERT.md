## Expected
- Output[0]: non-empty (PhaseUpdate formatted).
- Output[1]: empty string (PhaseEnd suppressed by coalescer).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if len(resp.Output) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(resp.Output))
	}
	if resp.Output[0] == "" {
		t.Fatalf("PhaseUpdate must produce formatted output")
	}
	if !strings.Contains(resp.Output[0], "hello") {
		t.Fatalf("PhaseUpdate output must contain delta text, got: %q", resp.Output[0])
	}
	if resp.Output[1] != "" {
		t.Fatalf("PhaseEnd after PhaseUpdate must be suppressed, got: %q", resp.Output[1])
	}
}
```
