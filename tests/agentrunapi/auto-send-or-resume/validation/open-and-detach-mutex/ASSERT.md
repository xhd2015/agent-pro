## Expected

- API error about mutual exclusion of open and detach.
- Zero dispatch hook calls.

## Side Effects

- None.

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
	// Parity with CLI: "--detach and --open are mutually exclusive; cannot use both"
	lower := strings.ToLower(resp.ErrString)
	ok := strings.Contains(lower, "mutual") ||
		(strings.Contains(lower, "open") && strings.Contains(lower, "detach") &&
			(strings.Contains(lower, "exclusive") || strings.Contains(lower, "both")))
	if !ok {
		t.Fatalf("error should mention open/detach mutual exclusion, got %q", resp.ErrString)
	}
	assertZeroHooks(t, resp)
}
```
