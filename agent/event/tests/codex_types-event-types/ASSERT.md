## Expected
- All 8 constant values are correct.
- The marshaled JSON contains proper codex event structure with `item.completed`, `command_execution`, and the command output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "EventStarted=item.started")
	assertContains(t, resp.Output, "EventUpdated=item.updated")
	assertContains(t, resp.Output, "EventCompleted=item.completed")
	assertContains(t, resp.Output, "EventError=error")
	assertContains(t, resp.Output, "ItemReasoning=reasoning")
	assertContains(t, resp.Output, "ItemCommandExecution=command_execution")
	assertContains(t, resp.Output, "ItemFileChange=file_change")
	assertContains(t, resp.Output, "ItemMessage=message")

	assertContains(t, resp.Output, `"type":"item.completed"`)
	assertContains(t, resp.Output, `"type":"command_execution"`)
	assertContains(t, resp.Output, `"command":"go test"`)
	assertContains(t, resp.Output, `"aggregated_output":"ok"`)
	assertContains(t, resp.Output, `"exit_code":0`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
