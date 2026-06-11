## Expected
- Exit code 0.
- Stdout contains the sub-agent's response text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, "I have implemented the feature.")
}
```
