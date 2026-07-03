## Expected

- Observer resize does not change PTY (resize dropped).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ResizeAccepted { t.Fatal("observer resize must be ignored") }
}
```
