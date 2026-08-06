## Expected

- Error containing `not found`.
- View is nil.

## Errors

- not found

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	msg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(msg, "not found") {
		t.Fatalf("error %q missing not found", resp.Err.Error())
	}
	if resp.View != nil {
		t.Fatalf("View should be nil, got %+v", resp.View)
	}
}
```
