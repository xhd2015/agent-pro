## Expected
- Grok native tool lines survive the write-events pipeline into trace output.
- Stdout contains READ and GREP tool blocks plus one ASSISTANT message block.

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

	for _, label := range []string{"READ", "GREP"} {
		if !strings.Contains(resp.Stdout, label) {
			t.Fatalf("stdout missing tool label %q after grok pipeline, got:\n%s", label, resp.Stdout)
		}
	}

	if !strings.Contains(resp.Stdout, "ASSISTANT") {
		t.Fatalf("stdout missing ASSISTANT block, got:\n%s", resp.Stdout)
	}
	if strings.Count(resp.Stdout, "]  💭") != 1 {
		t.Fatalf("expected 1 think block, got:\n%s", resp.Stdout)
	}
}
```