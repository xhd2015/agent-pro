## Errors

- Non-nil error.
- Message mentions exclude and/or empty / pattern.

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
	if !strings.Contains(msg, "exclude") && !strings.Contains(msg, "pattern") && !strings.Contains(msg, "empty") {
		t.Fatalf("error should mention empty exclude pattern: %q", resp.Err)
	}
}
```
