## Expected

- Non-zero exit code.
- Combined output mentions session not found (or similar).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertNonZeroExit(t, resp)
	lower := strings.ToLower(resp.Combined)
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "not found") {
		t.Fatalf("expected session-not-found error, got combined %q", resp.Combined)
	}
}
```