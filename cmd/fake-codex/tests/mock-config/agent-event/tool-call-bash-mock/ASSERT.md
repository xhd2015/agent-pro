## Expected
- The command succeeds.
- The completed codex event contains the mocked output and exit code, not the real command output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"type":"command_execution"`)
    assertContains(t, resp.Stdout, `"mocked-output"`)
    assertContains(t, resp.Stdout, `"exit_code":2`)
}
```
