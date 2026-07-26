## Expected

- API error mentioning timeout and/or ready/banner/sendable.
- At least one StatusFn poll.

## Side Effects

- StatusFn only.

## Errors

- Expected timeout API error.

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
	hasTimeout := strings.Contains(lower, "timeout")
	hasReadyHint := strings.Contains(lower, "banner") ||
		strings.Contains(lower, "sendable") ||
		strings.Contains(lower, "ready")
	if !hasTimeout && !hasReadyHint {
		t.Fatalf("error should mention timeout/ready failure, got %q", resp.ErrString)
	}
	if resp.StatusPollCalls < 1 {
		t.Fatal("expected at least one status poll")
	}
}
```
