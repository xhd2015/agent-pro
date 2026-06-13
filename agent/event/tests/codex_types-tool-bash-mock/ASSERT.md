## Expected
- Two codex events: one `item.started` and one `item.completed` with item type `command_execution`.
- The completed event contains the mock output and exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"item.started"`)
	assertContains(t, resp.Stdout, `"type":"item.completed"`)
	assertContains(t, resp.Stdout, `"type":"command_execution"`)
	assertContains(t, resp.Stdout, `"aggregated_output":"hello"`)
	assertContains(t, resp.Stdout, `"exit_code":0`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
