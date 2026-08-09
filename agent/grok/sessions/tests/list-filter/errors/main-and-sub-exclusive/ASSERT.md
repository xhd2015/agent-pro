## Expected

- `resp.Err` non-nil.
- Error message indicates mutual exclusion of main-agent and sub-agent
  (mentions main and sub, or exclusive / mutually).
- No panic.

## Errors

- Expected validation error from ListWithOptions.

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
	hasMain := strings.Contains(msg, "main")
	hasSub := strings.Contains(msg, "sub")
	hasExclusive := strings.Contains(msg, "exclusive") || strings.Contains(msg, "mutually") || strings.Contains(msg, "both")
	if !(hasMain && hasSub) && !hasExclusive {
		t.Fatalf("error %q should mention main/sub mutual exclusion", resp.Err.Error())
	}
}
```
