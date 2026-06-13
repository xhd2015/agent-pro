## Expected
- The command succeeds.
- stdout contains started and completed command_execution codex events.
- The completed event contains the actual command output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"command_execution"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"hello-from-bash"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
    assertContains(t, resp.Stdout, `"exit_code":0`)
}
```
