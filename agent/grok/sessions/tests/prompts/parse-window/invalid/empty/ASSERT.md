## Errors

- Non-nil error.
- Message mentions at least one of `Nd`, `Nh`, `Nm` (guidance for valid forms).

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
