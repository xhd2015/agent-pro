## Expected
- Exit code 0.
- `os.UserHomeDir()` returns the temporary `HOME` directory.

```go
import (
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertHomeDir(t, req, resp, err)
}
```
