## Expected
- The emitted event uses the session flag.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"sessionID":"sess_arg"`)
}
```

