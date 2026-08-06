## Errors

- Non-nil error.
- Message mentions Nd/Nh/Nm guidance.

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
	msg := resp.Err.Error()
	if !strings.Contains(msg, "Nd") && !strings.Contains(msg, "Nh") && !strings.Contains(msg, "Nm") {
		t.Fatalf("error should mention Nd/Nh/Nm guidance: %q", msg)
	}
}
```
