## Expected

- Tail-collected events are semantically equal to `grok_session.FromUpdatesJSONL`
  on the same wire fixture (tool_call_id, status, turn_index, ActionDone).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ExpectedEvents) == 0 {
		t.Fatal("ExpectedEvents not populated by Run")
	}
	assertSemanticEqual(t, resp.Events, resp.ExpectedEvents)
}
```