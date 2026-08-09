## Expected

- `resp.Err` non-nil.
- Error message mentions grep and empty (or must not be empty).

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
	if !strings.Contains(msg, "grep") {
		t.Fatalf("error %q should mention grep", resp.Err.Error())
	}
}
```
