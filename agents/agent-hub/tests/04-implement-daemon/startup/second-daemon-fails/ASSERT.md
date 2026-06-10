## Expected
- Second daemon fails with lock error.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { expectErrContains(t, resp, "lock") }
```

