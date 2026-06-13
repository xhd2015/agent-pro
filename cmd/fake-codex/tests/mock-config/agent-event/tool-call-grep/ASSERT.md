## Expected
- The command succeeds.
- The completed codex event contains the grep match output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"command_execution"`)
    assertContains(t, resp.Stdout, `"needle"`)
}
```
