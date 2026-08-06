## Expected

- No error.
- Window equals 2 hours.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Window != 2*time.Hour {
		t.Fatalf("Window=%v want 2h", resp.Window)
	}
}
```
