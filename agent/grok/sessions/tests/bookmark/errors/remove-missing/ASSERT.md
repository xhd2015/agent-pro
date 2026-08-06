## Expected

- Error containing `not found`.

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
}
```
