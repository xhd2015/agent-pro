## Expected
- The command succeeds.
- The completed codex event contains the file contents.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"command_execution"`)
    assertContains(t, resp.Stdout, `"file contents here"`)
}
```
