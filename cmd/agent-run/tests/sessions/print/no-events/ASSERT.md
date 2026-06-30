## Expected

- Exit code 0.
- Stdout contains `(no events yet)` and `Done (session finished)`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `
<contains>
(no events yet)
Done (session finished)
</contains>`)
}
```