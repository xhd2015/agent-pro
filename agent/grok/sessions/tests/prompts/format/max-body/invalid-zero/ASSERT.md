## Errors

- Non-nil error from format Write* (or harness-surfaced package validation).
- Message mentions `max-body` and/or positive / `>= 1` / invalid.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	msg := strings.ToLower(resp.Err.Error())
	hasBody := strings.Contains(msg, "max-body") || strings.Contains(msg, "maxbody") || strings.Contains(msg, "body")
	hasBound := strings.Contains(msg, ">=") || strings.Contains(msg, "positive") ||
		strings.Contains(msg, "invalid") || strings.Contains(msg, "at least") ||
		strings.Contains(msg, ">= 1") || strings.Contains(msg, ">=1")
	if !hasBody && !hasBound {
		t.Fatalf("error should mention invalid max-body N: %q", resp.Err)
	}
	// Prefer explicit max-body mention when implementer uses flag name in errors.
	if !hasBody {
		// Accept generic invalid/positive if body keyword missing but message is clear.
		if !strings.Contains(msg, "invalid") && !strings.Contains(msg, "positive") {
			t.Fatalf("error too vague for MaxBody=0: %q", resp.Err)
		}
	}
}
```
