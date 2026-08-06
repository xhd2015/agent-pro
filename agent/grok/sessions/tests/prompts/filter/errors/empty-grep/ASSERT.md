## Errors

- Non-nil error.
- Message mentions grep and/or empty / pattern.

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
	if !strings.Contains(msg, "grep") && !strings.Contains(msg, "pattern") && !strings.Contains(msg, "empty") {
		t.Fatalf("error should mention empty grep pattern: %q", resp.Err)
	}
}
```
