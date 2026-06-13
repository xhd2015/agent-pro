## Expected
- Two codex events: one `item.started` and one `item.completed` with item type `command_execution`.
- The completed event contains the mock output and exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"item.started"`)
	assertContains(t, resp.Output, `"type":"item.completed"`)
	assertContains(t, resp.Output, `"type":"command_execution"`)
	assertContains(t, resp.Output, `"aggregated_output":"hello"`)
	assertContains(t, resp.Output, `"exit_code":0`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
