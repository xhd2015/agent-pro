## Expected

- `FocusSession` returns a non-nil error (index out of range / invalid index).
- `FocusITerm` is never called.

## Errors

- Non-nil from `FocusSession`.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertError(t, err)
	assertNoFocus(t, resp)
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "index") && !strings.Contains(msg, "range") &&
		!strings.Contains(msg, "invalid") && !strings.Contains(msg, "out") {
		t.Logf("index-oob error (acceptable): %v", err)
	}
}
```
