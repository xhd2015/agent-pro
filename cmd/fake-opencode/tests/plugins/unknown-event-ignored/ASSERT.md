## Expected
- Exit code 0.
- No errors in stderr about unknown events.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
}
```
