## Expected
- Exit code 0.
- `require("os").homedir()` returns the temporary `HOME` directory.

```go
import (
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertHomeDir(t, req, resp, err)
}
```
