## Expected
- The answer references "french" or "capital" (case-insensitive).
- Response.SessionID is non-empty (a session was created and reused).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected non-empty SessionID after Ask()")
	}
	lower := strings.ToLower(resp.Answer)
	if !strings.Contains(lower, "french") && !strings.Contains(lower, "capital") {
		t.Fatalf("expected answer to reference 'french' or 'capital', got:\n%s", resp.Answer)
	}
}
```
