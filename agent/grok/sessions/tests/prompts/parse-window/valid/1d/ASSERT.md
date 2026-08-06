## Expected

- No error.
- Window equals 24 hours.

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
	if resp.Window != 24*time.Hour {
		t.Fatalf("Window=%v want 24h (1d rolling)", resp.Window)
	}
}
```
