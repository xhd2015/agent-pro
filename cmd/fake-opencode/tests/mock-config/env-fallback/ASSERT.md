## Expected
- The env fallback mock config is used.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"from env"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_env"`)
}
```

