## Expected
- Exit code 0.
- `require("os").homedir()` returns the temporary `HOME` directory.

```go
import (
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertHomeDir(t, req, resp, err)
}
```
