## Expected

- Third attach cannot write after writer detached.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ObserverCanWrite { t.Fatal("third attach must be read-only") }
}
```
