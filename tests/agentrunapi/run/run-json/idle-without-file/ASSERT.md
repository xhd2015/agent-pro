## Expected

- API error (timeout or empty/not-valid JSON).
- No successful JSON return.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	if resp.JSON != "" {
		t.Fatalf("idle without file must not return JSON, got %q", resp.JSON)
	}
	lower := strings.ToLower(resp.ErrString)
	if !strings.Contains(lower, "timeout") &&
		!strings.Contains(lower, "empty") &&
		!strings.Contains(lower, "json") {
		t.Fatalf("error should mention timeout/empty/json, got %q", resp.ErrString)
	}
}
```
