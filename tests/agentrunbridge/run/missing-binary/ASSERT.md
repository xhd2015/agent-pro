## Expected

- API error indicating binary not found / PATH failure.
- LookPath called at least once.
- LaunchCalls remains 0.

## Errors

- Expected API error only.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	lower := strings.ToLower(resp.ErrString)
	if !strings.Contains(lower, "not found") && !strings.Contains(lower, "path") && !strings.Contains(lower, "agent-run") {
		t.Fatalf("error should indicate missing binary, got %q", resp.ErrString)
	}
	if resp.LookPathCalls < 1 {
		t.Fatal("expected LookPath to be called")
	}
	assertEqual(t, "LaunchCalls", resp.LaunchCalls, 0)
}
```
