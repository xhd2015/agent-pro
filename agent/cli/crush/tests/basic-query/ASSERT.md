## Expected
- The answer contains "paris" (case-insensitive).
- Response.SessionID is non-empty (a session was created by the crush server).

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
	if !strings.Contains(strings.ToLower(resp.Answer), "paris") {
		t.Fatalf("expected answer to contain 'paris', got:\n%s", resp.Answer)
	}
}
```
