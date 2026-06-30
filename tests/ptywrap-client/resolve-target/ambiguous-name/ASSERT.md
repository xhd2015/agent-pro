## Expected

- Resolve returns error mentioning ambiguous name.
- Error lists matching session ids.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ResolveErr == "" {
		t.Fatal("expected ambiguity error")
	}
	lower := strings.ToLower(resp.ResolveErr)
	if !strings.Contains(lower, "ambiguous") {
		t.Fatalf("error should mention ambiguous: %q", resp.ResolveErr)
	}
	if !strings.Contains(resp.ResolveErr, "session-3") || !strings.Contains(resp.ResolveErr, "session-4") {
		t.Fatalf("error should list ids: %q", resp.ResolveErr)
	}
}
```