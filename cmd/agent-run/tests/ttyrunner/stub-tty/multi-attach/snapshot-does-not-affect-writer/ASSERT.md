## Expected

- Writer can still write after snapshot attach.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.WriterCanWrite { t.Fatal("writer should retain write after snapshot probe") }
}
```
