## Expected

- Registry file still exists after observer disconnect.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.SessionStillAlive { t.Fatal("session should remain after observer disconnect") }
}
```
