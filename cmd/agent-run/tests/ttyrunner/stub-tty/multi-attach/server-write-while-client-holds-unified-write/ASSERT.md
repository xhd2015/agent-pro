## Expected

- Server `tty send` succeeds concurrently with client unified writer.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.ServerSendWhileWriter {
		t.Fatal("server send should succeed while writer holds unified write")
	}
}
```
