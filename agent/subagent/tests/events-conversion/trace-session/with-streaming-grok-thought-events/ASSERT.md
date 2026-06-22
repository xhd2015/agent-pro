## Expected
- Six per-word think deltas coalesce into exactly one numbered thinking block in trace output.
- Full text "The user wants me to act" appears in the output.
- No fractional per-word `💭` headers.

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
		t.Fatalf("expected non-empty output")
	}

	thinkBlocks := strings.Count(resp.Stdout, "]  💭")
	if thinkBlocks != 1 {
		t.Fatalf("expected exactly 1 coalesced think block, got %d, output:\n%s", thinkBlocks, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "The user wants me to act") {
		t.Fatalf("expected coalesced text 'The user wants me to act' in output, got:\n%s", resp.Stdout)
	}
}
```