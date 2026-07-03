## Expected

- First interactive attach accepts keyboard input.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe == nil || !resp.MultiAttachProbe.WriterCanWrite {
		t.Fatal("expected interactive attach to accept input")
	}
}
```
