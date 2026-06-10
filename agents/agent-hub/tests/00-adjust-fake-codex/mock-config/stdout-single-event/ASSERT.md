## Expected
- The command succeeds.
- stdout contains the configured event.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"single response"`)
}
```

