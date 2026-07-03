## Expected

- After writer detach, earlier second attach cannot write.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ObserverCanWrite { t.Fatal("second attach must remain read-only after writer detach") }
}
```
