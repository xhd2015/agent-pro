## Expected
- Lock file exists.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { expectOK(t, resp, err) }
```

