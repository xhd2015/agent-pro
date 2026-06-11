## Expected
- The command succeeds.
- stderr contains the configured diagnostic.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stderr, "diagnostic line")
}
```

