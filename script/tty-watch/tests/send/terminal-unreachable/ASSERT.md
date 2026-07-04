## Expected

- Non-zero exit code.
- Stale registry file removed.
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
	if resp.RegistryExists {
		t.Fatalf("stale registry %s should be pruned", resp.SessionID)
	}
	lower := strings.ToLower(resp.Combined)
	if !strings.Contains(lower, "session") && !strings.Contains(lower, "not found") {
		t.Fatalf("expected session-not-found error, got combined %q", resp.Combined)
	}
}
```