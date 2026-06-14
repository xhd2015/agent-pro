## Expected
- Exit code 0.
- Stdout contains the delayed message.
- Elapsed wall time >= 1500ms (verified in the overridden Run).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"text":"delayed"`)
}
```
