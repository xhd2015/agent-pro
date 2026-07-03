## Expected

- Both writer and observer receive PTY output.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	p := resp.MultiAttachProbe
	if len(p.WriterReceived) == 0 { t.Fatal("writer received no output") }
	if len(p.ObserverReceived) == 0 { t.Fatal("observer received no output") }
}
```
