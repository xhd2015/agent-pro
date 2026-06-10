## Expected
- Peek does not advance.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { expectOK(t, resp, err) }
```

