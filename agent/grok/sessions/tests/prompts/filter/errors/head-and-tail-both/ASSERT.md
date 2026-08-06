## Errors

- Non-nil error.
- Message mentions both head and tail (or mutual exclusion / cannot both).

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
	if !(strings.Contains(msg, "head") && strings.Contains(msg, "tail")) {
		// also accept "mutually exclusive" style
		if !strings.Contains(msg, "mutual") && !strings.Contains(msg, "both") && !strings.Contains(msg, "exclusive") {
			t.Fatalf("error should mention head/tail conflict: %q", resp.Err)
		}
	}
}
```
