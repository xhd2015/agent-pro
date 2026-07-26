## Expected
- Two message deltas coalesce into exactly one ASSISTANT block.
- Full text "Hello" appears in the output.

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

	assistantHeaders := strings.Count(resp.Output, "ASSISTANT")
	if assistantHeaders != 1 {
		t.Fatalf("expected exactly 1 ASSISTANT block (coalesced), got %d:\n%s", assistantHeaders, resp.Output)
	}
	if !strings.Contains(resp.Output, "Hello") {
		t.Fatalf("expected coalesced text 'Hello' in output, got:\n%s", resp.Output)
	}
}
```