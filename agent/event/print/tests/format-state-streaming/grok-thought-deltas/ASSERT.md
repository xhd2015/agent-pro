## Expected
- All six think deltas coalesce into exactly one thinking header block.
- Full text "The user wants me to act" appears in the output.
- No fractional per-word think headers.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output == "" {
		t.Fatalf("expected non-empty output")
	}

	thinkHeaders := strings.Count(resp.Output, "💭")
	if thinkHeaders != 1 {
		t.Fatalf("expected exactly 1 coalesced think header, got %d:\n%s", thinkHeaders, resp.Output)
	}
	if !strings.Contains(resp.Output, "The user wants me to act") {
		t.Fatalf("expected coalesced text 'The user wants me to act' in output, got:\n%s", resp.Output)
	}
}
```