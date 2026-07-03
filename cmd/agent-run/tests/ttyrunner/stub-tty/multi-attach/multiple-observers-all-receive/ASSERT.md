## Expected

- Two observers both receive PTY output.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	p := resp.MultiAttachProbe
	if len(p.ObserverReceived) == 0 || len(p.Observer2Received) == 0 { t.Fatal("both observers should receive output") }
}
```
