## Expected

- No error.
- Window equals 30 minutes.

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
	if resp.Window != 30*time.Minute {
		t.Fatalf("Window=%v want 30m", resp.Window)
	}
}
```
