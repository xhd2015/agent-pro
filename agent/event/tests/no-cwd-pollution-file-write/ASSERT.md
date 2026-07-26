## Expected
- No files are created in the working directory.
- Only "DONE" is printed (no "FILE:" lines).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "DONE")
	assertNotContains(t, resp.Output, "FILE:")
}
```
