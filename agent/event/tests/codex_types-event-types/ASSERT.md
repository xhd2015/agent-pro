## Expected
- All 8 constant values are correct.
- The marshaled JSON contains proper codex event structure with `item.completed`, `command_execution`, and the command output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "EventStarted=item.started")
	assertContains(t, resp.Stdout, "EventUpdated=item.updated")
	assertContains(t, resp.Stdout, "EventCompleted=item.completed")
	assertContains(t, resp.Stdout, "EventError=error")
	assertContains(t, resp.Stdout, "ItemReasoning=reasoning")
	assertContains(t, resp.Stdout, "ItemCommandExecution=command_execution")
	assertContains(t, resp.Stdout, "ItemFileChange=file_change")
	assertContains(t, resp.Stdout, "ItemMessage=message")

	assertContains(t, resp.Stdout, `"type":"item.completed"`)
	assertContains(t, resp.Stdout, `"type":"command_execution"`)
	assertContains(t, resp.Stdout, `"command":"go test"`)
	assertContains(t, resp.Stdout, `"aggregated_output":"ok"`)
	assertContains(t, resp.Stdout, `"exit_code":0`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
