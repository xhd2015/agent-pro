## Expected
- Lock file is removed.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { expectOK(t, resp, err) }
```

