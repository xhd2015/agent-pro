## Expected

- API error (message mentions empty or prompt).
- Zero LookPath calls and zero launch calls (no exec).

## Errors

- Expected API error only.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	lower := strings.ToLower(resp.ErrString)
	if !strings.Contains(lower, "empty") && !strings.Contains(lower, "prompt") {
		t.Fatalf("error should mention empty/prompt, got %q", resp.ErrString)
	}
	assertEqual(t, "LookPathCalls", resp.LookPathCalls, 0)
	assertEqual(t, "LaunchCalls", resp.LaunchCalls, 0)
}
```
