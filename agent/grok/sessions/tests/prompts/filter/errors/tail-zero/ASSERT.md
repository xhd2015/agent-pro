## Errors

- Non-nil error.
- Message mentions tail and/or positive / >= 1 / invalid.

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
	if !strings.Contains(msg, "tail") && !strings.Contains(msg, "positive") && !strings.Contains(msg, "invalid") {
		t.Fatalf("error should mention invalid tail N: %q", resp.Err)
	}
}
```
