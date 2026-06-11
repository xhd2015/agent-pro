## Expected
- Exit code 0.
- The followup is delivered on the same session (the mock session responds with completion).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, "resumed and done")
}
```
