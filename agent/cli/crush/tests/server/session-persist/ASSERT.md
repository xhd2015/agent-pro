## Expected
- First `Ask` returns the capital of France ("paris").
- `LastSessionID` is non-empty after the first call.
- Second `Ask` (resumed with `SessionID`) returns an answer that references the first query context ("french" or "capital").

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("server-ask session-persist failed: %v", err)
	}
	lower := strings.ToLower(resp.Answer)
	if !strings.Contains(lower, "french") && !strings.Contains(lower, "capital") {
		t.Fatalf("expected answer to reference 'french' or 'capital', got:\n%s", resp.Answer)
	}
}
```
